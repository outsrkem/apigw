package models

import (
	"time"
)

// Create inserts a new api group record into database
func (d *DB) Create(group *OrmApiGroup) error {
	now := time.Now().UnixMilli()
	group.CreateTime = now
	group.UpdateTime = now
	return d.db.Create(group).Error
}

// GetByUUID queries api group record by unique uuid
func (d *DB) GetByUUID(uuid string) (*OrmApiGroup, error) {
	var item OrmApiGroup
	err := d.db.Where("id = ?", uuid).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// Update modifies existing api group record
func (d *DB) Update(group *OrmApiGroup) error {
	group.UpdateTime = time.Now().UnixMilli()
	return d.db.Save(group).Error
}

// Delete removes api group record by primary key kid
func (d *DB) Delete(kid int64) error {
	return d.db.Delete(&OrmApiGroup{}, "kid = ?", kid).Error
}

// List executes paginated query for api groups, returns record list and total quantity
func (d *DB) List(limit, offset int) ([]*OrmApiGroup, int64, error) {
	var list []*OrmApiGroup
	var total int64
	query := d.db.Model(&OrmApiGroup{})

	// Count total number of records
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Limit(limit).Offset(offset).Find(&list).Error
	return list, total, err
}

// ListAllEnabled queries all api groups whose status is enabled(status = 1)
func (d *DB) ListAllEnabled() ([]*OrmApiGroup, error) {
	var list []*OrmApiGroup
	err := d.db.Where("status = ?", 1).Find(&list).Error
	return list, err
}
