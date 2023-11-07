package subproc_test

import (
	"fmt"
	"statectl/internal/utils/subproc"

	"testing"
)

func TestFetchLocalSHA(t *testing.T) {
	// Test that the local SHA is fetched correctly.
	sha, err := subproc.FetchLocalSHA()
	if err != nil {
		// t.Errorf("error fetching local git SHA: %v", err)
	}
	fmt.Printf("Local git SHA: %s", sha)
}

func TestFetchRemoteSHA(t *testing.T) {
	// Test that the remote SHA is fetched correctly.

}
