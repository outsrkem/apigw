package channel

import (
	"apigw/src/cache"
	"apigw/src/models"
)

func refreshLcCache(lcID string) error {
	dao, err := models.NewDBModel("SYSTEM")
	if err != nil {
		return err
	}

	newLc, err := dao.GetLoadChannelByID(lcID)
	if err != nil {
		return err
	}
	cache.GlobalCache.RefreshSingleChannel(lcID, newLc)
	return nil
}

func deleteLcCache(lcID string) {
	cache.GlobalCache.DeleteSingleChannel(lcID)
}
