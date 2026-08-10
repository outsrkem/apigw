package models

import (
	"apigw/src/pkg/common"
	"errors"
	"time"

	"gorm.io/gorm"
)

const (
	StatusDisable     int8 = 0 // 禁用
	StatusEnable      int8 = 1 // 启用
	PublishUnReleased int8 = 0 // 未发布
	PublishTesting    int8 = 1 // 测试中
	PublishReleased   int8 = 2 // 已发布
	PublishOffline    int8 = 3 // 已下线
)

type IApiInterfaceDao interface {
	OfflineApiById(apiId string) error
	OnlineApiById(apiId string) error
	UpdateApiStatusById(apiID string, status int8) error
}

// UpdateApiStatusById updates api enable/disable status by api id
func (d *ApiInterfaceDao) UpdateApiStatusById(apiID string, status int8) error {
	updateTime := common.CreateTimestamp()
	return d.db.Model(&OrmApiInterface{}).
		Where("id = ?", apiID).Updates(map[string]any{
		"status":      status,
		"update_time": updateTime,
	}).Error
}

// OfflineApiById temporarily offline single api, set publish_status to 3
func (d *ApiInterfaceDao) OfflineApiById(apiId string) error {
	now := common.CreateTimestamp()
	return d.db.Model(&OrmApiInterface{}).Where("id = ?", apiId).
		Updates(map[string]any{
			"publish":     PublishOffline,
			"update_time": now,
		}).Error
}

// OnlineApiById restore offline api online, set publish_status to 2
func (d *ApiInterfaceDao) OnlineApiById(apiId string) error {
	now := common.CreateTimestamp()
	return d.db.Model(&OrmApiInterface{}).Where("id = ?", apiId).
		Updates(map[string]any{
			"publish":     PublishReleased,
			"update_time": now,
		}).Error
}

// CreateApiInterface insert a new api interface record into database
func (d *DB) CreateApiInterface(item *OrmApiInterface) error {
	now := common.CreateTimestamp()
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
	err := d.db.Where("lc_id = ?", lcID).Order("kid DESC").Find(&list).Error
	return list, err
}

// UpdateApiInterface update existing api interface record
func (d *DB) UpdateApiInterface(id string, data map[string]any) error {
	data["update_time"] = common.CreateTimestamp()
	return d.db.Model(&OrmApiInterface{}).Where("id=?", id).Updates(data).Error
}

// DeleteApiInterface physically delete api record by primary id
func (d *DB) DeleteApiInterface(id string) error {
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
	err = query.Limit(limit).Offset(offset).Order("kid DESC").Find(&list).Error
	return list, total, err
}

// GetApiById query complete latest single api data by unique api id
func (d *DB) GetApiById(apiId string) (*OrmApiInterface, error) {
	var item OrmApiInterface
	err := d.db.Where("id = ?", apiId).First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

// ListAllEnabledApiInterface query all apis which are enabled and officially published
// Filter condition: status=1(enabled), publish_status=2(published)
func (d *DB) ListAllEnabledApiInterface() ([]*OrmApiInterface, error) {
	var list []*OrmApiInterface
	err := d.db.Where("status = ? AND publish_status = ?", 1, 2).Find(&list).Error
	return list, err
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
