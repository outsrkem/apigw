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

	// 1. 加载分组
	groups, _, err := dao.ListApiGroup(99999, 0)
	if err != nil {
		return err
	}

	// 2. 加载域名
	domains, _, err := dao.ListApiDomain(99999, 0)
	if err != nil {
		return err
	}

	// 3. 加载路由
	routes, _, err := dao.ListApiInterface(99999, 0)
	if err != nil {
		return err
	}

	// 4. 过滤有效路由
	var valid []*models.OrmApiInterface
	for _, item := range routes {
		if item.Status == 1 && item.Publish == 2 {
			valid = append(valid, item)
		}
	}

	// 5. 构建域名索引并刷新路由
	c.RefreshRoute(groups, domains, valid)
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
	c.RefreshChannel(list)
	return nil
}
