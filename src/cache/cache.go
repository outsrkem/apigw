package cache

import (
	"apigw/src/models"
	"sync"
	"sync/atomic"
)

// GatewayCache global runtime cache for API gateway routing & load balancing data
type GatewayCache struct {
	sync.RWMutex
	// activeRoutes list of all valid & enabled API route definitions
	activeRoutes []*models.OrmApiInterface
	// channelMap key: channel unique ID, value: channel detailed configuration
	channelMap map[string]*models.OrmLoadChannel
	// roundRobinIdx store atomic int64 counter for each load channel
	// key = channelID, value = *int64 atomic counter
	roundRobinIdx sync.Map
}

var GlobalCache *GatewayCache

func NewGatewayCache() *GatewayCache {
	cache := &GatewayCache{
		channelMap: make(map[string]*models.OrmLoadChannel),
	}
	return cache
}

func SetGlobalCache(c *GatewayCache) {
	GlobalCache = c
}

// GetAllActiveRoutes get a copied slice of all currently active API routes
func (c *GatewayCache) GetAllActiveRoutes() []*models.OrmApiInterface {
	c.RLock()
	defer c.RUnlock()
	list := make([]*models.OrmApiInterface, len(c.activeRoutes))
	copy(list, c.activeRoutes)
	return list
}

// RefreshRoutes overwrite cached active route list with new route data
func (c *GatewayCache) RefreshRoutes(routes []*models.OrmApiInterface) {
	c.Lock()
	defer c.Unlock()
	c.activeRoutes = routes
}

// RefreshChannels rebuild channel map with latest channel configuration list
func (c *GatewayCache) RefreshChannels(channels []*models.OrmLoadChannel) {
	c.Lock()
	newMap := make(map[string]*models.OrmLoadChannel)
	for _, ch := range channels {
		newMap[ch.ID] = ch
	}
	c.channelMap = newMap
	c.Unlock()
}

// GetChannelByLCID query channel config by channel unique ID
func (c *GatewayCache) GetChannelByLCID(lcID string) *models.OrmLoadChannel {
	if c == nil {
		return nil
	}
	c.RLock()
	defer c.RUnlock()
	return c.channelMap[lcID]
}

// ClearChannelCache clean up scheduling cache of specified channel
func (c *GatewayCache) ClearChannelCache(lcID string) {
	c.Lock()
	defer c.Unlock()
	c.roundRobinIdx.Delete(lcID)
}

// getOrCreateRRCounter get atomic counter, create new counter if not exist
func (c *GatewayCache) getOrCreateRRCounter(lcID string) *int64 {
	val, ok := c.roundRobinIdx.Load(lcID)
	if ok {
		return val.(*int64)
	}
	newCnt := int64(0)
	actual, _ := c.roundRobinIdx.LoadOrStore(lcID, &newCnt)
	return actual.(*int64)
}

// SelectNodeIndex unified load balancing scheduling entrance, Only supports rr polling
func (c *GatewayCache) SelectNodeIndex(lcID string, nodeCount int) int {
	if c == nil || nodeCount <= 0 {
		return -1
	}
	cnt := c.getOrCreateRRCounter(lcID)
	old := atomic.LoadInt64(cnt)
	newVal := old + 1
	atomic.CompareAndSwapInt64(cnt, old, newVal)
	return int(old) % nodeCount
}

// RefreshSingleRoute updates a single route record by ApiID
func (c *GatewayCache) RefreshSingleRoute(apiID string, newRoute *models.OrmApiInterface) {
	c.Lock()
	defer c.Unlock()
	for i := 0; i < len(c.activeRoutes); i++ {
		if c.activeRoutes[i].ID == apiID {
			c.activeRoutes[i] = newRoute
			return
		}
	}
	c.activeRoutes = append(c.activeRoutes, newRoute)
}

// DeleteRoute completely removes specified route from cache
func (c *GatewayCache) DeleteRoute(apiID string) {
	c.Lock()
	defer c.Unlock()
	newList := make([]*models.OrmApiInterface, 0, len(c.activeRoutes))
	for _, item := range c.activeRoutes {
		if item.ID != apiID {
			newList = append(newList, item)
		}
	}
	c.activeRoutes = newList
}
