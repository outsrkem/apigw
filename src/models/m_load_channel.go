package models

import "time"

// CreateLoadChannel creates a new load channel record in database
func (d *DB) CreateLoadChannel(item *OrmLoadChannel) error {
	now := time.Now().UnixMilli()
	item.CreateTime = now
	item.UpdateTime = now
	return d.db.Create(item).Error
}

// GetLoadChannelByKid queries load channel by primary key kid
func (d *DB) GetLoadChannelByKid(kid int64) (*OrmLoadChannel, error) {
	var data OrmLoadChannel
	err := d.db.Where("kid = ?", kid).First(&data).Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}

// GetLoadChannelByUUID queries load channel by unique channel uuid
func (d *DB) GetLoadChannelByUUID(uuid string) (*OrmLoadChannel, error) {
	var data OrmLoadChannel
	err := d.db.Where("id = ?", uuid).First(&data).Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}

// UpdateLoadChannel updates existing load channel record
func (d *DB) UpdateLoadChannel(item *OrmLoadChannel) error {
	item.UpdateTime = time.Now().UnixMilli()
	return d.db.Save(item).Error
}

// DeleteLoadChannel deletes load channel record by primary key kid
func (d *DB) DeleteLoadChannel(kid int64) error {
	return d.db.Delete(&OrmLoadChannel{}, "kid = ?", kid).Error
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
	err = query.Limit(limit).Offset(offset).Find(&list).Error
	return list, total, err
}

// ListAllEnabledLoadChannel queries all load channels with enabled status(status = 1)
func (d *DB) ListAllEnabledLoadChannel() ([]*OrmLoadChannel, error) {
	var list []*OrmLoadChannel
	err := d.db.Where("status = ?", 1).Find(&list).Error
	return list, err
}
