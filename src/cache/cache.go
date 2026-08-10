// src/cache/cache.go
package cache

import (
	"apigw/src/models"
	"strings"
	"sync"
	"sync/atomic"
)

const (
	RouteModeExact  = "EXACT"  // Exact path matching mode
	RouteModePrefix = "PREFIX" // Prefix the longest matching mode
	RouteMethodAny  = "Any"    // Represent matching all HTTP request methods
)

// ==================== Data Structures ====================
// CachedRoute cached api route rule
type CachedRoute struct {
	ID, GroupId, Protocol, Method, ReqUri, BackendUri, Mode, Auth, LcID string
	RateLimit                                                           int
}

// CachedChannel cached load‑balance backend channel
type CachedChannel struct {
	ID, Name, CaCert    string
	Timeout, HcInterval int
	Status              int8
	Nodes               []string
}

// CachedGroup cached api group
type CachedGroup struct {
	ID, Name   string
	Status     int8
	ValidHosts map[string]bool
}

// MethodRoutes route container grouped by match mode
type MethodRoutes struct {
	exactRoutes     []*CachedRoute // Mode = "exact"
	prefixRoutes    []*CachedRoute // Mode = "prefix"
	anyExactRoutes  []*CachedRoute // Method = "Any" + Mode = "exact"
	anyPrefixRoutes []*CachedRoute // Method = "Any" + Mode = "prefix"
}

// GatewayCache in‑memory cache for gateway config
type GatewayCache struct {
	sync.RWMutex
	hostIndex     map[string]string // host → groupID
	defaultGroups []string          // groups without bound domain
	groupRoutes   map[string]*MethodRoutes
	groupMap      map[string]*CachedGroup
	channelMap    map[string]*CachedChannel
	roundRobinIdx sync.Map // channelID → round‑robin counter
}

var GlobalCache *GatewayCache

// NewGatewayCache create new gateway cache instance
func NewGatewayCache() *GatewayCache {
	return &GatewayCache{
		hostIndex:     make(map[string]string),
		defaultGroups: make([]string, 0),
		groupRoutes:   make(map[string]*MethodRoutes),
		groupMap:      make(map[string]*CachedGroup),
		channelMap:    make(map[string]*CachedChannel),
	}
}

// SetGlobalCache set global cache singleton
func SetGlobalCache(c *GatewayCache) { GlobalCache = c }

// ==================== Private Helper Functions ====================

// isValidRoute check whether route can be loaded into cache
func isValidRoute(r *models.OrmApiInterface) bool {
	return r != nil && r.Status == 1 && r.Publish == 2
}

// cleanChannelSchedule remove channel and its round‑robin state
func (c *GatewayCache) cleanChannelSchedule(lcID string) {
	delete(c.channelMap, lcID)
	c.roundRobinIdx.Delete(lcID)
}

// toCachedChannel convert db model to cached channel struct
func toCachedChannel(ch *models.OrmLoadChannel) *CachedChannel {
	nodes := strings.Split(ch.Backend, ",")
	valid := make([]string, 0, len(nodes))
	for _, n := range nodes {
		if n = strings.TrimSpace(n); n != "" {
			valid = append(valid, n)
		}
	}
	return &CachedChannel{
		ID:         ch.ID,
		Name:       ch.Name,
		Timeout:    ch.Timeout,
		HcInterval: ch.HcInterval,
		Status:     ch.Status,
		CaCert:     ch.CaCert,
		Nodes:      valid,
	}
}

// filterRoute remove target route by id from slice
func filterRoute(routes []*CachedRoute, id string) []*CachedRoute {
	result := make([]*CachedRoute, 0, len(routes))
	for _, r := range routes {
		if r.ID != id {
			result = append(result, r)
		}
	}
	return result
}

// matchSingleGroupRoute match route inside one group
func matchSingleGroupRoute(mr *MethodRoutes, path, method string) *CachedRoute {
	if mr == nil {
		return nil
	}
	// priority 1: specific method + exact match
	for _, r := range mr.exactRoutes {
		if r.Method == method && r.ReqUri == path {
			return r
		}
	}

	// priority 2: specific method + prefix match，Select the longest prefix
	var bestPrefix *CachedRoute
	for _, r := range mr.prefixRoutes {
		if r.Method == method && strings.HasPrefix(path, r.ReqUri) {
			if bestPrefix == nil || len(r.ReqUri) > len(bestPrefix.ReqUri) {
				bestPrefix = r
			}
		}
	}
	if bestPrefix != nil {
		return bestPrefix
	}

	// priority 3: Any method + exact match
	for _, r := range mr.anyExactRoutes {
		if r.ReqUri == path {
			return r
		}
	}

	// priority 4: Any method + prefix match，Select the longest prefix
	var bestAnyPrefix *CachedRoute
	for _, r := range mr.anyPrefixRoutes {
		if strings.HasPrefix(path, r.ReqUri) {
			if bestAnyPrefix == nil || len(r.ReqUri) > len(bestAnyPrefix.ReqUri) {
				bestAnyPrefix = r
			}
		}
	}
	return bestAnyPrefix
}

// ==================== Core Query ====================

// MatchRoute find matched route by host, path and http method
func (c *GatewayCache) MatchRoute(host, path, method string) *CachedRoute {
	c.RLock()
	defer c.RUnlock()

	gid, ok := c.hostIndex[host]
	if ok {
		return matchSingleGroupRoute(c.groupRoutes[gid], path, method)
	}

	// host not found, iterate default groups in order
	for _, defGid := range c.defaultGroups {
		route := matchSingleGroupRoute(c.groupRoutes[defGid], path, method)
		if route != nil {
			return route
		}
	}
	return nil
}

// GetChannel get load‑balance channel by id
func (c *GatewayCache) GetChannel(lcID string) *CachedChannel {
	c.RLock()
	defer c.RUnlock()
	return c.channelMap[lcID]
}

// ==================== Full Refresh Methods ====================

// Refresh full reload: groups + domains + routes + channels
func (c *GatewayCache) Refresh(
	groups []*models.OrmApiGroup,
	domains []*models.OrmApiDomain,
	routes []*models.OrmApiInterface,
	channels []*models.OrmLoadChannel,
) {
	c.Lock()
	defer c.Unlock()
	c.refreshGroups(groups, domains)
	c.refreshRoutes(routes)
	c.refreshChannels(channels)
}

// RefreshRoute reload groups, domains and routes, keep channels unchanged
func (c *GatewayCache) RefreshRoute(
	groups []*models.OrmApiGroup,
	domains []*models.OrmApiDomain,
	routes []*models.OrmApiInterface,
) {
	c.Lock()
	defer c.Unlock()
	c.refreshGroups(groups, domains)
	c.refreshRoutes(routes)
}

// RefreshChannel reload load‑balance channels only
func (c *GatewayCache) RefreshChannel(channels []*models.OrmLoadChannel) {
	c.Lock()
	defer c.Unlock()
	c.refreshChannels(channels)
}

// ==================== Internal Full‑Refresh Private Methods ====================

func (c *GatewayCache) refreshGroups(groups []*models.OrmApiGroup, domains []*models.OrmApiDomain) {
	c.hostIndex = make(map[string]string)
	c.defaultGroups = make([]string, 0)
	c.groupMap = make(map[string]*CachedGroup)

	domainCount := make(map[string]int)
	for _, d := range domains {
		if d.Status != 1 || domainCount[d.GroupID] >= 8 {
			continue
		}
		c.hostIndex[d.Name] = d.GroupID
		domainCount[d.GroupID]++
	}

	for _, g := range groups {
		if g.Status != 1 {
			continue
		}
		hosts := make(map[string]bool)
		for _, d := range domains {
			if d.Status == 1 && d.GroupID == g.ID {
				hosts[d.Name] = true
			}
		}
		if len(hosts) == 0 {
			c.defaultGroups = append(c.defaultGroups, g.ID)
		}
		c.groupMap[g.ID] = &CachedGroup{
			ID:         g.ID,
			Name:       g.Name,
			Status:     g.Status,
			ValidHosts: hosts,
		}
	}
}

func (c *GatewayCache) refreshRoutes(routes []*models.OrmApiInterface) {
	c.groupRoutes = make(map[string]*MethodRoutes)

	for _, r := range routes {
		if !isValidRoute(r) {
			continue
		}
		if _, ok := c.groupMap[r.GroupID]; !ok {
			continue
		}
		c.addRouteLocked(r)
	}
}

func (c *GatewayCache) refreshChannels(channels []*models.OrmLoadChannel) {
	c.channelMap = make(map[string]*CachedChannel)
	for _, ch := range channels {
		if ch.Status != 1 {
			continue
		}
		c.channelMap[ch.ID] = toCachedChannel(ch)
	}
}

// ==================== Route Incremental Update ====================

// AddRoute add single route to cache
func (c *GatewayCache) AddRoute(r *models.OrmApiInterface) {
	c.Lock()
	defer c.Unlock()
	if !isValidRoute(r) {
		return
	}
	if _, ok := c.groupMap[r.GroupID]; !ok {
		return
	}
	c.addRouteLocked(r)
}

// UpdateRoute update existing single route
func (c *GatewayCache) UpdateRoute(r *models.OrmApiInterface) {
	c.Lock()
	defer c.Unlock()
	c.deleteRoute(r.ID)
	if isValidRoute(r) {
		if _, ok := c.groupMap[r.GroupID]; ok {
			c.addRouteLocked(r)
		}
	}
}

// DeleteRoute remove route by id
func (c *GatewayCache) DeleteRoute(routeID string) {
	c.Lock()
	defer c.Unlock()
	c.deleteRoute(routeID)
}

// addRouteLocked add route, must hold write lock before calling
func (c *GatewayCache) addRouteLocked(r *models.OrmApiInterface) {
	route := &CachedRoute{
		ID:         r.ID,
		GroupId:    r.GroupID,
		Protocol:   r.Protocol,
		Method:     r.Method,
		ReqUri:     r.ReqUri,
		BackendUri: r.BackendUri,
		Mode:       r.Mode,
		Auth:       r.Auth,
		LcID:       r.LcID,
		RateLimit:  r.RateLimit,
	}
	mr := c.groupRoutes[r.GroupID]
	if mr == nil {
		mr = &MethodRoutes{}
		c.groupRoutes[r.GroupID] = mr
	}
	if r.Method == RouteMethodAny {
		if r.Mode == RouteModeExact {
			mr.anyExactRoutes = append(mr.anyExactRoutes, route)
		} else {
			mr.anyPrefixRoutes = append(mr.anyPrefixRoutes, route)
		}
	} else {
		if r.Mode == RouteModeExact {
			mr.exactRoutes = append(mr.exactRoutes, route)
		} else {
			mr.prefixRoutes = append(mr.prefixRoutes, route)
		}
	}
}

// deleteRoute remove route by id, must hold write lock before calling
func (c *GatewayCache) deleteRoute(routeID string) {
	for gid, mr := range c.groupRoutes {
		mr.exactRoutes = filterRoute(mr.exactRoutes, routeID)
		mr.prefixRoutes = filterRoute(mr.prefixRoutes, routeID)
		mr.anyExactRoutes = filterRoute(mr.anyExactRoutes, routeID)
		mr.anyPrefixRoutes = filterRoute(mr.anyPrefixRoutes, routeID)
		if len(mr.exactRoutes)+len(mr.prefixRoutes)+
			len(mr.anyExactRoutes)+len(mr.anyPrefixRoutes) == 0 {
			delete(c.groupRoutes, gid)
		}
	}
}

// ==================== Channel Incremental Update ====================

// AddChannel add single load‑balance channel
func (c *GatewayCache) AddChannel(ch *models.OrmLoadChannel) {
	c.Lock()
	defer c.Unlock()
	if ch.Status != 1 {
		return
	}
	c.channelMap[ch.ID] = toCachedChannel(ch)
}

// UpdateChannel update single load‑balance channel
func (c *GatewayCache) UpdateChannel(ch *models.OrmLoadChannel) {
	c.Lock()
	defer c.Unlock()
	if ch.Status != 1 {
		c.cleanChannelSchedule(ch.ID)
		return
	}
	c.channelMap[ch.ID] = toCachedChannel(ch)
}

// DeleteSingleChannel public entry to delete channel and its scheduler state
func (c *GatewayCache) DeleteSingleChannel(lcID string) {
	c.Lock()
	defer c.Unlock()
	c.cleanChannelSchedule(lcID)
}

// RefreshSingleRoute reload one route
func (c *GatewayCache) RefreshSingleRoute(apiID string, newRoute *models.OrmApiInterface) {
	c.Lock()
	defer c.Unlock()
	c.deleteRoute(apiID)
	if isValidRoute(newRoute) {
		if _, ok := c.groupMap[newRoute.GroupID]; ok {
			c.addRouteLocked(newRoute)
		}
	}
}

// RefreshSingleChannel reload one load‑balance channel
func (c *GatewayCache) RefreshSingleChannel(lcID string, newChannel *models.OrmLoadChannel) {
	c.Lock()
	defer c.Unlock()
	if newChannel == nil || newChannel.Status != 1 {
		c.cleanChannelSchedule(lcID)
		return
	}
	c.channelMap[newChannel.ID] = toCachedChannel(newChannel)
}

// ==================== Round‑Robin Load Balancer ====================

// SelectNode pick backend node index using round‑robin
func (c *GatewayCache) SelectNode(lcID string) int {
	ch := c.GetChannel(lcID)
	if ch == nil || len(ch.Nodes) == 0 {
		return -1
	}
	val, _ := c.roundRobinIdx.LoadOrStore(lcID, new(int64))
	cnt := val.(*int64)
	old := atomic.AddInt64(cnt, 1)
	return int(old-1) % len(ch.Nodes)
}

// ClearChannelCache delete round‑robin counter for channel
func (c *GatewayCache) ClearChannelCache(lcID string) {
	c.Lock()
	defer c.Unlock()
	c.roundRobinIdx.Delete(lcID)
}
