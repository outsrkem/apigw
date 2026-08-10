package apinterface

import (
	"apigw/src/cache"
	"apigw/src/models"
)

// OfflineApi takes a single api offline: update database status then refresh single route cache
func OfflineApi(apiID string) error {
	dao, err := models.NewSystemDBModel()
	if err != nil {
		return err
	}
	// Modify offline status record in database
	err = dao.ApiInterface.OfflineApiById(apiID)
	if err != nil {
		return err
	}
	// Query latest api configuration from database
	newRoute, err := dao.GetApiById(apiID)
	if err != nil {
		return err
	}
	// Update specified single route in global cache
	cache.GlobalCache.RefreshSingleRoute(apiID, newRoute)
	return nil
}

// OnlineApi restores a single offline api to online status
func OnlineApi(apiID string) error {
	dao, err := models.NewSystemDBModel()
	if err != nil {
		return err
	}
	err = dao.ApiInterface.OnlineApiById(apiID)
	if err != nil {
		return err
	}
	newRoute, err := dao.GetApiById(apiID)
	if err != nil {
		return err
	}
	cache.GlobalCache.RefreshSingleRoute(apiID, newRoute)
	return nil
}
