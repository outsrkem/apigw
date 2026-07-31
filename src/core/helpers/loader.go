package helpers

import (
	"apigw/src/cache"
	"apigw/src/models"
)

// LoadAndRefreshRoute queries api route data from database and refreshes route cache
// It filters valid routes that are enabled and published, then updates global gateway route cache
func LoadAndRefreshRoute(c *cache.GatewayCache) error {
	dao, err := models.NewDBModel("instanceId")
	if err != nil {
		return err
	}

	list, _, err := dao.ListApiInterface(99999, 0)
	if err != nil {
		return err
	}
	// Filter: only keep enabled and published api routes
	var valid []*models.OrmApiInterface
	for _, item := range list {
		if item.Status == 1 && item.PublishStatus == 2 {
			valid = append(valid, item)
		}
	}
	c.RefreshRoutes(valid)
	return nil
}

// LoadAndRefreshChannel loads all channel configurations from database and refreshes channel cache
func LoadAndRefreshChannel(c *cache.GatewayCache) error {
	dao, err := models.NewDBModel("instanceId")
	if err != nil {
		return err
	}

	list, _, err := dao.ListLoadChannel(99999, 0)
	if err != nil {
		return err
	}
	c.RefreshChannels(list)
	return nil
}
