package models

import (
	"apigw/src/pkg/common"
	"errors"

	"gorm.io/gorm"
)

type ILcCaDao interface {
	CreateLcCa(item *OrmLcCa) error
}

// CreateLcCa creates a new CA record
func (d *LcCaDao) CreateLcCa(item *OrmLcCa) error {
	return d.db.Create(item).Error
}

// UpdateLcCa Update channel certificate information
func (d *DB) UpdateLcCa(id string, data map[string]any) error {
	data["update_time"] = common.CreateTimestamp()
	return d.db.Model(&OrmLcCa{}).Where("id=?", id).Updates(data).Error
}

// GetLcCaByID query record by unique certificate ID
func (d *DB) GetLcCaByID(id string) (*OrmLcCa, error) {
	var data OrmLcCa
	err := d.db.Where("id = ?", id).First(&data).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &data, nil
}

// DeleteLcCa delete record by primary key kid
func (d *DB) DeleteLcCa(id string) error {
	return d.db.Delete(&OrmLcCa{}, "id = ?", id).Error
}

// ListLcCa paginated query for CA list, return data list and total count
func (d *DB) ListLcCa(limit, offset int) ([]*OrmLcCa, int64, error) {
	var list []*OrmLcCa
	var total int64
	query := d.db.Model(&OrmLcCa{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Limit(limit).Offset(offset).Order("kid DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
