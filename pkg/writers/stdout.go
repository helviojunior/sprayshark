package writers

import (
	"fmt"
	"os"

	"github.com/helviojunior/sprayshark/pkg/models"
	logger "github.com/helviojunior/sprayshark/pkg/log"
)

// StdoutWriter is a Stdout writer
type StdoutWriter struct {
}

// NewStdoutWriter initialises a stdout writer
func NewStdoutWriter() (*StdoutWriter, error) {
	return &StdoutWriter{}, nil
}

// Write results to stdout
func (s *StdoutWriter) Write(result *models.Result) error {
	fmt.Fprintf(os.Stderr, "                                                                               \r")
	if result.Failed {
		logger.Errorf("[%s] user=%s", result.FailedReason, result.User)
		return nil
	}
	if result.ValidCredential {
		logger.Warnf("[Credential found] %s:%s", result.User, result.Password)
		return nil
	}
	if result.UserExists {
		logger.Infof("[Invalid Creds] %s:%s", result.User, result.Password)
		return nil
	}
	logger.Infof("[Invalid User] username: %s", result.User)
	return nil
}
