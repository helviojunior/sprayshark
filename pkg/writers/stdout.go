package writers

import (
	//"fmt"
	"os"

	"github.com/helviojunior/sprayshark/pkg/models"
	log "github.com/sirupsen/logrus"
)

// StdoutWriter is a Stdout writer
type StdoutWriter struct {
}

// NewStdoutWriter initialises a stdout writer
func NewStdoutWriter() (*StdoutWriter, error) {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.TraceLevel)
	return &StdoutWriter{}, nil
}

// Write results to stdout
func (s *StdoutWriter) Write(result *models.Result) error {
	if result.Failed {
		log.Errorf("[%s] username: %s", result.FailedReason, result.User)
		return nil
	}
	if result.ValidCredential {
		log.Warnf("[Credential found] %s:%s", result.User, result.Password)
		return nil
	}
	if result.UserExists {
		log.Infof("[Invalid Creds] %s:%s", result.User, result.Password)
		return nil
	}
	log.Infof("[Invalid User] username: %s", result.User)
	return nil
}
