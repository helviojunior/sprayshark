package writers

import "github.com/helviojunior/sprayshark/pkg/models"

// Writer is a results writer
type Writer interface {
	Write(*models.Result) error
}
