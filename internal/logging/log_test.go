package logging_test

import (
	"statectl/internal/logging"
	"testing"
)

func TestGetLog(t *testing.T) {
	log := logging.GetLogger()
	if log == nil {
		t.Errorf("error getting logger")
	}
}
