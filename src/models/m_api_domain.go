package models

import (
	"apigw/src/pkg/common"
	"time"
)

type IApiDomainDao interface {
	Create(item *OrmApiDomain) error
	Delete(id string) (int64, error)
	ListByGroupId(groupID string) ([]*OrmApiDomain, error)
}

func (ad *ApiDomainDao) Create(item *OrmApiDomain) error {
	now := common.CreateTimestamp()
	item.CreateTime = now
	item.UpdateTime = now
	return ad.db.Create(item).Error
}

func (ad *ApiDomainDao) Delete(id string) (int64, error) {
	result := ad.db.Delete(&OrmApiDomain{}, "id = ?", id)
	return result.RowsAffected, result.Error
}

func (ad *ApiDomainDao) ListByGroupId(groupID string) ([]*OrmApiDomain, error) {
	var list []*OrmApiDomain
	err := ad.db.Where("group_id = ?", groupID).Order("kid DESC").Find(&list).Error
	return list, err
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

// UpdateApiDomain updates existing domain configuration record
func (d *DB) UpdateApiDomain(item *OrmApiDomain) error {
	item.UpdateTime = time.Now().UnixMilli()
	return d.db.Save(item).Error
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
	err = query.Limit(limit).Offset(offset).Order("kid DESC").Find(&list).Error
	return list, total, err
}

// ListAllEnabledApiDomain queries all domains with enabled status(status = 1)
func (d *DB) ListAllEnabledApiDomain() ([]*OrmApiDomain, error) {
	var list []*OrmApiDomain
	err := d.db.Where("status = ?", 1).Find(&list).Error
	return list, err
}
