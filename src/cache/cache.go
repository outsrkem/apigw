package cache

import (
	"apigw/src/models"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// sharedDialer global reused tcp dialer with keep-alive configuration
var sharedDialer = &net.Dialer{
	Timeout:   300 * time.Millisecond,
	KeepAlive: 1 * time.Second,
}

// transportWrapper wrap http.Transport with last access timestamp for idle recycle
type transportWrapper struct {
	tr           *http.Transport
	lastAccessTs int64 // unix second timestamp
}

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
	// ipHashMap ip-hash scheduling cache, key format: channelID|clientIP
	ipHashMap map[string]int

	// channelTransport isolated http transport wrapper for each plain http channel
	channelTransport map[string]*transportWrapper
}

// GlobalCache singleton global gateway cache instance
var GlobalCache *GatewayCache

// idleRecycleConfig recycle rule constant
const (
	idleRecycleInterval = 5 * time.Minute
	idleMaxDuration     = 30 * time.Minute
	cleanInterval       = 20 * time.Second
)

// NewGatewayCache create and initialize a new GatewayCache instance
func NewGatewayCache() *GatewayCache {
	cache := &GatewayCache{
		channelMap:       make(map[string]*models.OrmLoadChannel),
		ipHashMap:        make(map[string]int),
		channelTransport: make(map[string]*transportWrapper),
	}
	// start background idle transport recycle goroutine
	go cache.startIdleRecycleTask()
	return cache
}

// SetGlobalCache assign given cache instance to global singleton variable
func SetGlobalCache(c *GatewayCache) {
	GlobalCache = c
}

// GetAllActiveRoutes get a copied slice of all currently active API routes
// use read lock to guarantee thread-safe read operation
func (c *GatewayCache) GetAllActiveRoutes() []*models.OrmApiInterface {
	c.RLock()
	defer c.RUnlock()
	list := make([]*models.OrmApiInterface, len(c.activeRoutes))
	copy(list, c.activeRoutes)
	return list
}

// RefreshRoutes overwrite cached active route list with new route data
// exclusive lock ensures atomic write operation
func (c *GatewayCache) RefreshRoutes(routes []*models.OrmApiInterface) {
	c.Lock()
	defer c.Unlock()
	c.activeRoutes = routes
}

// RefreshChannels rebuild channel map with latest channel configuration list
// clean offline channel transport immediately
func (c *GatewayCache) RefreshChannels(channels []*models.OrmLoadChannel) {
	c.Lock()
	newMap := make(map[string]*models.OrmLoadChannel)
	activeSet := make(map[string]struct{})
	for _, ch := range channels {
		newMap[ch.ID] = ch
		activeSet[ch.ID] = struct{}{}
	}
	c.channelMap = newMap
	// remove transport of deleted channel
	for cid, wrap := range c.channelTransport {
		if _, ok := activeSet[cid]; !ok {
			wrap.tr.CloseIdleConnections()
			delete(c.channelTransport, cid)
		}
	}
	c.Unlock()
}

// GetChannelTransport lazy initialize isolated transport for specified channel
// refresh access timestamp every call to delay recycle time
func (c *GatewayCache) GetChannelTransport(channelID string) *http.Transport {
	now := time.Now().Unix()
	c.RLock()
	wrap, exist := c.channelTransport[channelID]
	c.RUnlock()
	if exist {
		atomic.StoreInt64(&wrap.lastAccessTs, now)
		return wrap.tr
	}

	newTr := &http.Transport{
		DialContext:           sharedDialer.DialContext,
		MaxIdleConns:          1000,
		MaxIdleConnsPerHost:   200,
		MaxConnsPerHost:       400,
		IdleConnTimeout:       1200 * time.Millisecond,
		ResponseHeaderTimeout: 1000 * time.Millisecond,
		ExpectContinueTimeout: 500 * time.Millisecond,
		DisableKeepAlives:     false,
	}
	newWrap := &transportWrapper{
		tr:           newTr,
		lastAccessTs: now,
	}

	c.Lock()
	if existWrap, ok := c.channelTransport[channelID]; ok {
		newTr.CloseIdleConnections()
		c.Unlock()
		atomic.StoreInt64(&existWrap.lastAccessTs, now)
		return existWrap.tr
	}
	c.channelTransport[channelID] = newWrap
	c.Unlock()

	// regular idle connection cleanup inside transport
	go func(t *http.Transport) {
		ticker := time.NewTicker(cleanInterval)
		defer ticker.Stop()
		for range ticker.C {
			t.CloseIdleConnections()
		}
	}(newTr)

	return newTr
}

// CleanStaleTransport manually clean offline channel transport
func (c *GatewayCache) CleanStaleTransport(keepChannelIDs map[string]struct{}) {
	c.Lock()
	defer c.Unlock()
	for cid, wrap := range c.channelTransport {
		if _, keep := keepChannelIDs[cid]; !keep {
			wrap.tr.CloseIdleConnections()
			delete(c.channelTransport, cid)
		}
	}
}

// startIdleRecycleTask background task to recycle idle transport over 30 minutes
func (c *GatewayCache) startIdleRecycleTask() {
	ticker := time.NewTicker(idleRecycleInterval)
	defer ticker.Stop()
	threshold := int64(idleMaxDuration.Seconds())
	for range ticker.C {
		now := time.Now().Unix()
		c.Lock()
		for cid, wrap := range c.channelTransport {
			last := atomic.LoadInt64(&wrap.lastAccessTs)
			if now-last > threshold {
				wrap.tr.CloseIdleConnections()
				delete(c.channelTransport, cid)
			}
		}
		c.Unlock()
	}
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
	prefix := lcID + "|"
	for k := range c.ipHashMap {
		if strings.HasPrefix(k, prefix) {
			delete(c.ipHashMap, k)
		}
	}
}

// simpleHash lightweight string hash function
func simpleHash(s string) int {
	res := 0
	for _, b := range []byte(s) {
		res = res*31 + int(b)
	}
	if res < 0 {
		res = -res
	}
	return res
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

// SelectNodeIndex unified load balancing scheduling entrance
// Only supports round robin polling
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
