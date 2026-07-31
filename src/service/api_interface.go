package service

import (
	"apigw/src/cache"
	"apigw/src/models"
)

// OfflineApi takes a single api offline: update database status then refresh single route cache
func OfflineApi(apiID string) error {
	dao, err := models.NewDBModel("SYSTEM")
	if err != nil {
		return err
	}
	// Modify offline status record in database
	err = dao.OfflineApiById(apiID)
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
	dao, err := models.NewDBModel("SYSTEM")
	if err != nil {
		return err
	}
	err = dao.OnlineApiById(apiID)
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

// DisableApi permanently disables specified api
func DisableApi(apiID string) error {
	dao, err := models.NewDBModel("SYSTEM")
	if err != nil {
		return err
	}
	err = dao.DisableApiById(apiID)
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

// BatchOfflineApi takes multiple apis offline in batch mode
func BatchOfflineApi(apiIDs []string) error {
	dao, err := models.NewDBModel("SYSTEM")
	if err != nil {
		return err
	}
	// Execute batch status update in database
	err = dao.BatchOfflineApiByIds(apiIDs)
	if err != nil {
		return err
	}
	// Query all valid online & published api routes
	onlineRoutes, err := dao.ListAllEnabledApiInterface()
	if err != nil {
		return err
	}
	// Fully refresh global route cache
	cache.GlobalCache.RefreshRoutes(onlineRoutes)
	return nil
}

// BatchOnlineApi restores multiple offline apis online in batch mode
func BatchOnlineApi(apiIDs []string) error {
	dao, err := models.NewDBModel("SYSTEM")
	if err != nil {
		return err
	}
	err = dao.BatchOnlineApiByIds(apiIDs)
	if err != nil {
		return err
	}
	onlineRoutes, err := dao.ListAllEnabledApiInterface()
	if err != nil {
		return err
	}
	cache.GlobalCache.RefreshRoutes(onlineRoutes)
	return nil
}

// BatchDisableApi permanently disables multiple apis in batch mode
func BatchDisableApi(apiIDs []string) error {
	dao, err := models.NewDBModel("SYSTEM")
	if err != nil {
		return err
	}
	err = dao.BatchDisableApiByIds(apiIDs)
	if err != nil {
		return err
	}
	onlineRoutes, err := dao.ListAllEnabledApiInterface()
	if err != nil {
		return err
	}
	cache.GlobalCache.RefreshRoutes(onlineRoutes)
	return nil
}
