package writers

import (
	"sync"

	"github.com/helviojunior/sprayshark/internal/islazy"
	"github.com/helviojunior/sprayshark/pkg/database"
	//"github.com/helviojunior/sprayshark/pkg/log"
	"github.com/helviojunior/sprayshark/pkg/models"
	"gorm.io/gorm"
)

var hammingThreshold = 10

// DbWriter is a Database writer
type DbWriter struct {
	URI           string
	conn          *gorm.DB
	mutex         sync.Mutex
	hammingGroups []islazy.HammingGroup
}

// NewDbWriter initialises a database writer
func NewDbWriter(uri string, debug bool) (*DbWriter, error) {
	c, err := database.Connection(uri, false, debug)
	if err != nil {
		return nil, err
	}

	return &DbWriter{
		URI:           uri,
		conn:          c,
		mutex:         sync.Mutex{},
		hammingGroups: []islazy.HammingGroup{},
	}, nil
}

// Write results to the database
func (dw *DbWriter) Write(result *models.Result) error {
	dw.mutex.Lock()
	defer dw.mutex.Unlock()
	return dw.conn.Create(result).Error
}

