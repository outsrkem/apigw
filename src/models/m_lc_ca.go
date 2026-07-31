package models

// CreateLcCa creates a new CA record
func (d *DB) CreateLcCa(item *OrmLcCa) error {
	return d.db.Create(item).Error
}

// GetLcCaByID query record by unique certificate ID
func (d *DB) GetLcCaByID(id string) (*OrmLcCa, error) {
	var data OrmLcCa
	err := d.db.Where("id = ?", id).First(&data).Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}

// DeleteLcCa delete record by primary key kid
func (d *DB) DeleteLcCa(kid int64) error {
	return d.db.Delete(&OrmLcCa{}, "kid = ?", kid).Error
}

// ListLcCa paginated query for CA list, return data list and total count
func (d *DB) ListLcCa(limit, offset int) ([]*OrmLcCa, int64, error) {
	var list []*OrmLcCa
	var total int64
	query := d.db.Model(&OrmLcCa{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
