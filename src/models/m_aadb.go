package models

import (
	"apigw/src/database/mysql"
	"apigw/src/pkg/common"
	"errors"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ApiDomainDao struct{ db *gorm.DB }
type ApiGroupDao struct{ db *gorm.DB }
type ApiInterfaceDao struct{ db *gorm.DB }
type LcCaDao struct{ db *gorm.DB }
type LoadChannelDao struct{ db *gorm.DB }

// DB encapsulates gorm instance, logger and environment information for database operations
type DB struct {
	db           *gorm.DB      // GORM database instance
	Klog         *logrus.Entry // Log object for recording runtime information
	instanceId   string        // Unique identification of service instance
	CurrentTime  int64         // Unified timestamp used in business logic
	Domain       IApiDomainDao
	ApiGroup     IApiGroupDao
	ApiInterface IApiInterfaceDao
	LcCa         ILcCaDao
	LoadChannel  ILoadChannelDao
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

func NewSystemDBModel() (*DB, error) {
	_t := &DB{db: mysql.OrmDB}
	_t.Domain = &ApiDomainDao{db: _t.db}
	_t.ApiGroup = &ApiGroupDao{db: _t.db}
	_t.ApiInterface = &ApiInterfaceDao{db: _t.db}
	_t.LcCa = &LcCaDao{db: _t.db}
	_t.LoadChannel = &LoadChannelDao{db: _t.db}
	return _t, nil
}
