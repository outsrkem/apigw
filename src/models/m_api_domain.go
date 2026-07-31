package models

import "time"

// CreateApiDomain creates a new domain configuration record
func (d *DB) CreateApiDomain(item *OrmApiDomain) error {
	now := time.Now().UnixMilli()
	item.CreateTime = now
	item.UpdateTime = now
	return d.db.Create(item).Error
}

// GetApiDomainByID queries domain record by primary key id
func (d *DB) GetApiDomainByID(id int64) (*OrmApiDomain, error) {
	var data OrmApiDomain
	err := d.db.Where("id = ?", id).First(&data).Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}

// GetApiDomainByName queries domain by domain name with unique index
func (d *DB) GetApiDomainByName(name string) (*OrmApiDomain, error) {
	var data OrmApiDomain
	err := d.db.Where("name = ?", name).First(&data).Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}

// ListApiDomainByGroupID queries domain list bound to specified group uuid
func (d *DB) ListApiDomainByGroupID(groupID string) ([]*OrmApiDomain, error) {
	var list []*OrmApiDomain
	err := d.db.Where("group_id = ?", groupID).Find(&list).Error
	return list, err
}

// UpdateApiDomain updates existing domain configuration record
func (d *DB) UpdateApiDomain(item *OrmApiDomain) error {
	item.UpdateTime = time.Now().UnixMilli()
	return d.db.Save(item).Error
}

// DeleteApiDomain deletes domain record by primary key id
func (d *DB) DeleteApiDomain(id int64) error {
	return d.db.Delete(&OrmApiDomain{}, "id = ?", id).Error
}

// ListApiDomain executes paginated domain query, returns record list and total count
func (d *DB) ListApiDomain(limit, offset int) ([]*OrmApiDomain, int64, error) {
	var list []*OrmApiDomain
	var total int64
	query := d.db.Model(&OrmApiDomain{})

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = query.Limit(limit).Offset(offset).Find(&list).Error
	return list, total, err
}

// ListAllEnabledApiDomain queries all domains with enabled status(status = 1)
func (d *DB) ListAllEnabledApiDomain() ([]*OrmApiDomain, error) {
	var list []*OrmApiDomain
	err := d.db.Where("status = ?", 1).Find(&list).Error
	return list, err
}
