package docker

type LogEntry struct {
	Stream string `json:"stream"`
	Log    string `json:"log"`
}
