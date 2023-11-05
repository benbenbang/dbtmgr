package lock

import "errors"

type Comments struct {
	Commit  string `json:"commit"`
	Trigger string `json:"trigger"`
	Extra   string `json:"extra"`
}

type LockInfo struct {
	LockID    string   `json:"lock_id"`
	TimeStamp string   `json:"timestamp"`
	Signer    string   `json:"signer"`
	Comments  Comments `json:"comments"`
}

var LockExists = errors.New("lock exists")
