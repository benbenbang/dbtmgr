package test

import (
	"encoding/json"
	"statectl/internal/utils/types"
)

func CreateLockInfo(expectedSHA string) (types.LockInfo, []byte) {
	lockInfo := types.LockInfo{
		LockID:    expectedSHA,
		TimeStamp: "2021-01-01T00:00:00Z",
		Signer:    "test",
		Comments: types.Comments{
			Commit:  "ok",
			Trigger: "ok",
			Extra:   "ok",
		},
	}

	lockInfoRaw, err := json.Marshal(lockInfo)
	if err != nil {
		panic(err)
	}

	return lockInfo, lockInfoRaw
}
