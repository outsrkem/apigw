package models

import (
	"apigw/src/pkg/common"
	"errors"
	"time"

	"gorm.io/gorm"
)

type IApiGroupDao interface {
	Create(group *OrmApiGroup) error
}

// Create inserts a new api group record into database
func (d *ApiGroupDao) Create(group *OrmApiGroup) error {
	now := time.Now().UnixMilli()
	group.CreateTime = now
	group.UpdateTime = now
	return d.db.Create(group).Error
}

// GetApiGroupByID queries api group record by unique uuid
func (d *DB) GetApiGroupByID(id string) (*OrmApiGroup, error) {
	var item OrmApiGroup
	err := d.db.Where("id = ?", id).First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (d *DB) GetApiGroupByName(name string) (*OrmApiGroup, error) {
	var item OrmApiGroup
	err := d.db.Where("name = ?", name).First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

// UpdateMap modifies existing api group record
func (d *DB) UpdateApiGroup(id string, data map[string]any) error {
	data["update_time"] = common.CreateTimestamp()
	return d.db.Model(&OrmApiGroup{}).Where("id=?", id).Updates(data).Error
}

// DeleteApiGroup removes api group record by primary key id
func (d *DB) DeleteApiGroup(id string) error {
	return d.db.Delete(&OrmApiGroup{}, "id = ?", id).Error
}

// ListApiGroup executes paginated query for api groups, returns record list and total quantity
func (d *DB) ListApiGroup(limit, offset int) ([]*OrmApiGroup, int64, error) {
	var list []*OrmApiGroup
	var total int64
	query := d.db.Model(&OrmApiGroup{})

	// Count total number of records
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Limit(limit).Offset(offset).Order("kid DESC").Find(&list).Error
	return list, total, err
}

// ListAllEnabled queries all api groups whose status is enabled(status = 1)
func (d *DB) ListAllEnabled() ([]*OrmApiGroup, error) {
	var list []*OrmApiGroup
	err := d.db.Where("status = ?", 1).Order("kid DESC").Find(&list).Error
	return list, err
}
