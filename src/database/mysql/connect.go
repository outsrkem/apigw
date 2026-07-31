package mysql

import (
	"apigw/src/cfgtypes"
	"apigw/src/slog"
	"context"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// OrmDB global gorm database instance
var OrmDB *gorm.DB

// InitDB initializes mysql connection pool with automatic retry
func InitDB(cfg *cfgtypes.Apigw) {
	klog := slog.FromCtx(context.Background())
	dbcfg := cfg.Database
	var (
		retries   int           = 100
		backoff   time.Duration = time.Second
		dsn       string
		err       error
		dbConnect *gorm.DB
	)
	// Assemble mysql DSN string
	dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=true&loc=Local",
		dbcfg.User, dbcfg.Passwd, dbcfg.Host, dbcfg.Port, dbcfg.Name)
	klog.Debugf("database passwd: %s", dbcfg.Passwd)
	klog.Infof("Connect database: %s:%s", dbcfg.Host, dbcfg.Port)

	// GORM global configuration
	ormCfg := gorm.Config{
		Logger:                 NewGormLogger(cfg),
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "",
			SingularTable: true,
		},
	}

	// Retry loop for database connection
	for i := 0; i <= retries; i++ {
		dbConnect, err = gorm.Open(mysql.Open(dsn), &ormCfg)
		if err == nil {
			sqlDB, _ := dbConnect.DB()
			// Configure connection pool parameters
			sqlDB.SetMaxOpenConns(50)
			sqlDB.SetMaxIdleConns(5)
			sqlDB.SetConnMaxLifetime(time.Hour * 2)
			sqlDB.SetConnMaxIdleTime(time.Hour)
			// Test connection availability
			if err = sqlDB.Ping(); err == nil {
				OrmDB = dbConnect
				break
			}
		}

		klog.Warnf("Connection attempt %d/%d failed: %v. Retrying in %v...", i+1, retries, err, backoff)
		time.Sleep(backoff)
		// Exponential backoff upper limit
		if backoff <= 30*time.Second {
			backoff += time.Second
		}
	}

	// Exit program if connection still fails after all retries
	if err != nil {
		panic(err)
	}
}
