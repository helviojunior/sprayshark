package runner

import (
	"fmt"

	"github.com/helviojunior/sprayshark/pkg/models"
)

// ChromeNotFoundError signals that chrome is not available
type ChromeNotFoundError struct {
	Err error
}

func (e ChromeNotFoundError) Error() string {
	return fmt.Sprintf("chrome not found: %v", e.Err)
}

// Driver is the interface browser drivers will implement.
type Driver interface {
	Check(username string, password string, runner *Runner, to int, enumOnly bool) (*models.Result, error)
	Close()
}
