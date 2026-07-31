package models

import "time"

// CreateApiInterface insert a new api interface record into database
func (d *DB) CreateApiInterface(item *OrmApiInterface) error {
	now := time.Now().UnixMilli()
	item.CreateTime = now
	item.UpdateTime = now
	return d.db.Create(item).Error
}

// ListApiInterfaceByGroupID query api list under specified group unique uuid
func (d *DB) ListApiInterfaceByGroupID(groupID string) ([]*OrmApiInterface, error) {
	var list []*OrmApiInterface
	err := d.db.Where("group_id = ?", groupID).Find(&list).Error
	return list, err
}

// ListApiInterfaceByLcID query all apis bound to specified load channel uuid
func (d *DB) ListApiInterfaceByLcID(lcID string) ([]*OrmApiInterface, error) {
	var list []*OrmApiInterface
	err := d.db.Where("lc_id = ?", lcID).Find(&list).Error
	return list, err
}

// UpdateApiInterface update existing api interface record
func (d *DB) UpdateApiInterface(item *OrmApiInterface) error {
	item.UpdateTime = time.Now().UnixMilli()
	return d.db.Save(item).Error
}

// DeleteApiInterface physically delete api record by primary id
func (d *DB) DeleteApiInterface(id int64) error {
	return d.db.Delete(&OrmApiInterface{}, "id = ?", id).Error
}

// ListApiInterface paginate query api interface list, return record list and total count
func (d *DB) ListApiInterface(limit, offset int) ([]*OrmApiInterface, int64, error) {
	var list []*OrmApiInterface
	var total int64
	query := d.db.Model(&OrmApiInterface{})

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = query.Limit(limit).Offset(offset).Find(&list).Error
	return list, total, err
}

// ListAllEnabledApiInterface query all apis which are enabled and officially published
// Filter condition: status=1(enabled), publish_status=2(published)
func (d *DB) ListAllEnabledApiInterface() ([]*OrmApiInterface, error) {
	var list []*OrmApiInterface
	err := d.db.Where("status = ? AND publish_status = ?", 1, 2).Find(&list).Error
	return list, err
}

// OfflineApiById temporarily offline single api, set publish_status to 3
func (d *DB) OfflineApiById(apiId string) error {
	now := time.Now().UnixMilli()
	return d.db.Model(&OrmApiInterface{}).
		Where("id = ?", apiId).Updates(map[string]any{
		"publish_status": 3,
		"update_time":    now,
	}).Error
}

// OnlineApiById restore offline api online, set publish_status to 2
func (d *DB) OnlineApiById(apiId string) error {
	now := time.Now().UnixMilli()
	return d.db.Model(&OrmApiInterface{}).Where("id = ?", apiId).
		Updates(map[string]any{
			"publish_status": 2,
			"update_time":    now,
		}).Error
}

// DisableApiById permanently disable api, set status to 0
func (d *DB) DisableApiById(apiId string) error {
	now := time.Now().UnixMilli()
	return d.db.Model(&OrmApiInterface{}).Where("id = ?", apiId).
		Updates(map[string]any{
			"status":      0,
			"update_time": now,
		}).Error
}

// GetApiById query complete latest single api data by unique api id
func (d *DB) GetApiById(apiId string) (*OrmApiInterface, error) {
	var item OrmApiInterface
	err := d.db.Where("id = ?", apiId).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// BatchOfflineApiByIds batch offline apis, set publish_status to 3
func (d *DB) BatchOfflineApiByIds(apiIDs []string) error {
	now := time.Now().UnixMilli()
	return d.db.Model(&OrmApiInterface{}).
		Where("id IN (?)", apiIDs).
		Updates(map[string]any{
			"publish_status": 3,
			"update_time":    now,
		}).Error
}

// BatchOnlineApiByIds batch restore apis online, set publish_status to 2
func (d *DB) BatchOnlineApiByIds(apiIDs []string) error {
	now := time.Now().UnixMilli()
	return d.db.Model(&OrmApiInterface{}).
		Where("id IN (?)", apiIDs).
		Updates(map[string]any{
			"publish_status": 2,
			"update_time":    now,
		}).Error
}

// BatchDisableApiByIds batch permanently disable apis, set status to 0
func (d *DB) BatchDisableApiByIds(apiIDs []string) error {
	now := time.Now().UnixMilli()
	return d.db.Model(&OrmApiInterface{}).
		Where("id IN (?)", apiIDs).
		Updates(map[string]any{
			"status":      0,
			"update_time": now,
		}).Error
}
