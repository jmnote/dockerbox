package box

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type Box struct {
	cli         *client.Client
	opts        BoxOpts
	containerID string
	statsStart  time.Time
	statsDone   chan bool
	statsResult chan Stats
	result      RunResult
}

type BoxOpts struct {
	*client.Client
	container.Config
	Timeout int
}

type RunResult struct {
	IsTimedOut bool
	CPU        uint64 // cpu time nanoseconds
	MEM        uint64 // bytes
	Time       int64  // milliseconds
	Logs       []LogEntry
	ExitCode  
	Warnings   []error
}

type Stats struct {
	cpu uint64 // cpu time nanoseconds
	mem uint64 // bytes
}

type LogEntry struct {
	Stream string `json:"stream"`
	Log    string `json:"log"`
}

func Run(opts BoxOpts) (*RunResult, error) {
	var cli *client.Client
	if opts.Client == nil {
		newcli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			return nil, fmt.Errorf("newClient err: %w", err)
		}
		cli = newcli
	} else {
		cli = opts.Client
	}
	if opts.Timeout == 0 {
		opts.Timeout = 60000 // 60s
	}
	b := &Box{
		cli:  cli,
		opts: opts,
	}
	if err := b.precheck(); err != nil {
		return nil, fmt.Errorf("precheck err: %w", err)
	}
	if err := b.run(); err != nil {
		return nil, fmt.Errorf("run err: %w", err)
	}
	b.postcheck()
	return &b.result, nil
}

func (b *Box) run() error {
	b.startCollectStats()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(b.opts.Timeout)*time.Millisecond)
	defer cancel()

	if err := b.cli.ContainerStart(ctx, b.containerID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	statusCode, err := b.waitContainer(ctx)
	if err != nil {
		b.result.IsTimedOut = true
	}

	b.stopCollectStats()
	b.collectLogs()
	return nil
}

func (b *Box) waitContainer(ctx context.Context) (int, error) {
	statusCh, errCh := b.cli.ContainerWait(ctx, b.containerID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		return 0, fmt.Errorf("failed to wait for container: %w", err)
	case status <-statusCh:
		return status.StatusCode, fmt.Errorf("failed to wait for container: %w", err)
	}
}

func (b *Box) collectLogs() {
	out, err := b.cli.ContainerLogs(context.Background(), b.containerID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		b.result.Warnings = append(b.result.Warnings, fmt.Errorf("failed to get container logs: %w", err))
		return
	}
	defer out.Close()
	b.parseLogEntries(out)
}

func (b *Box) parseLogEntries(out io.Reader) {
	data, err := io.ReadAll(out)
	if err != nil {
		b.result.Warnings = append(b.result.Warnings, fmt.Errorf("reading logs err: %w", err))
		return
	}

	var logEntries []LogEntry
	for len(data) > 0 {
		streamType := data[0]
		var stream string
		switch streamType {
		case 1:
			stream = "stdout"
		case 2:
			stream = "stderr"
		default:
			b.result.Warnings = append(b.result.Warnings, fmt.Errorf("reading logs err: %w", err))
			return
		}

		msgLength := binary.BigEndian.Uint32(data[4:])
		msg := data[8 : 8+msgLength]

		logEntry := LogEntry{
			Stream: stream,
			Log:    string(msg),
		}
		logEntries = append(logEntries, logEntry)

		data = data[8+msgLength:]
	}
	b.result.Logs = logEntries
}

func (b *Box) startCollectStats() {
	b.statsStart = time.Now()
	b.statsDone = make(chan bool)
	b.statsResult = make(chan Stats)
	go b.collectStats()
}

func (b *Box) stopCollectStats() {
	b.result.Time = time.Since(b.statsStart).Milliseconds()
	b.statsDone <- true
	stats := <-b.statsResult
	b.result.CPU = stats.cpu
	b.result.MEM = stats.mem
}

func (b *Box) collectStats() {
	var cpu uint64
	var mem uint64

	for {
		select {
		case <-b.statsDone:
			b.statsResult <- Stats{cpu, mem}
			close(b.statsResult)
			return
		default:
			b.getStats(&cpu, &mem)
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (b *Box) getStats(cpu *uint64, mem *uint64) {
	stats, err := b.cli.ContainerStatsOneShot(context.Background(), b.containerID)
	if err != nil {
		fmt.Printf("containerStatsOneShot err: %v\n", err)
		return
	}
	defer stats.Body.Close()

	var v container.StatsResponse
	if err := json.NewDecoder(stats.Body).Decode(&v); err != nil {
		fmt.Printf("stats decode err: %v\n", err)
		return
	}
	totalCPU := v.CPUStats.CPUUsage.TotalUsage
	if totalCPU > *cpu {
		*cpu = totalCPU
	}
	if v.MemoryStats.Usage > *mem {
		*mem = v.MemoryStats.Usage
	}
}
