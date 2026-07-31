package models

import (
	"apigw/src/database/mysql"
	"apigw/src/pkg/common"
	"errors"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// DB encapsulates gorm instance, logger and environment information for database operations
type DB struct {
	db          *gorm.DB      // GORM database instance
	Klog        *logrus.Entry // Log object for recording runtime information
	instanceId  string        // Unique identification of service instance
	CurrentTime int64         // Unified timestamp used in business logic
}

// NewDBModel creates and initializes a DB operation instance
func NewDBModel(instanceId string) (*DB, error) {
	if instanceId == "" {
		return nil, errors.New("instanceId can not be empty")
	}
	_t := &DB{
		instanceId:  instanceId,
		db:          mysql.OrmDB,
		CurrentTime: common.CreateTimestamp(),
	}
	return _t, nil
}

// SetKlog binds log entry to DB instance, supports chained calls
func (d *DB) SetKlog(klog *logrus.Entry) *DB {
	d.Klog = klog
	return d
}

// SetCurrentTime customizes unified business timestamp, supports chained calls
func (d *DB) SetCurrentTime(currentTime int64) *DB {
	d.CurrentTime = currentTime
	return d
}

// WithInstance returns gorm query builder filtered by current instanceId
func (d *DB) WithInstance() *gorm.DB {
	return d.db.Where("instance_id = ?", d.instanceId)
}
