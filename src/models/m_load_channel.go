package models

import (
	"apigw/src/pkg/common"
	"errors"
	"time"

	"gorm.io/gorm"
)

type ILoadChannelDao interface {
	CreateLoadChannel(item *OrmLoadChannel) error
}

// CreateLoadChannel creates a new load channel record in database
func (d *LoadChannelDao) CreateLoadChannel(item *OrmLoadChannel) error {
	now := time.Now().UnixMilli()
	item.CreateTime = now
	item.UpdateTime = now
	return d.db.Create(item).Error
}

// GetLoadChannelByID queries load channel by unique channel uuid
func (d *DB) GetLoadChannelByID(id string) (*OrmLoadChannel, error) {
	var data OrmLoadChannel
	err := d.db.Where("id = ?", id).First(&data).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &data, nil
}

func (d *DB) GetLoadChannelByLcCaID(lcCaID string) ([]*OrmLoadChannel, error) {
	var data []*OrmLoadChannel
	err := d.db.Where("ca_cert = ?", lcCaID).Find(&data).Error
	if err != nil {
		return nil, err
	}
	return data, nil
}

// UpdateLoadChannel updates existing load channel record
func (d *DB) UpdateLoadChannel(id string, data map[string]any) error {
	data["update_time"] = common.CreateTimestamp()
	return d.db.Model(&OrmLoadChannel{}).Where("id=?", id).Updates(data).Error
}

// DeleteLoadChannel deletes load channel record by primary key kid
func (d *DB) DeleteLoadChannel(id string) error {
	return d.db.Delete(&OrmLoadChannel{}, "id = ?", id).Error
}

// ListLoadChannel paginates load channel query, returns record list and total quantity
func (d *DB) ListLoadChannel(limit, offset int) ([]*OrmLoadChannel, int64, error) {
	var list []*OrmLoadChannel
	var total int64
	query := d.db.Model(&OrmLoadChannel{})

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = query.Limit(limit).Offset(offset).Order("kid DESC").Find(&list).Error
	return list, total, err
}

// ListAllEnabledLoadChannel queries all load channels with enabled status(status = 1)
func (d *DB) ListAllEnabledLoadChannel() ([]*OrmLoadChannel, error) {
	var list []*OrmLoadChannel
	err := d.db.Where("status = ?", 1).Find(&list).Error
	return list, err
}

// SetLoadChannelStatusById Set load channel status
func (d *DB) SetLoadChannelStatusById(lcID string, status int8) (int64, error) {
	if lcID == "" {
		return 0, errors.New("id cannot be empty")
	}

	now := common.CreateTimestamp()
	result := d.db.Model(&OrmLoadChannel{}).Where("id = ?", lcID).
		Updates(map[string]any{"status": status, "update_time": now})
	return result.RowsAffected, result.Error
}
