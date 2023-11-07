package types

type State struct {
	VersionID string `json:"version_id"`
	CommitSHA string `json:"commit_sha"`
	Bucket    string `json:"bucket"`
	Key       string `json:"key"`
}
