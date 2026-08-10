// src/core/service/core.go
package service

import (
	"apigw/src/cache"
	"context"
)

// Route matching mode constants
const (
	RouteModeExact  = "EXACT"  // Exact path matching mode
	RouteModePrefix = "PREFIX" // Prefix longest matching mode
	RouteMethodAny  = "Any"    // Represent matching all HTTP request methods
)

// RouteSvc global singleton RouteService instance
var RouteSvc *RouteService

// RouteService provides route matching capabilities for gateway requests
type RouteService struct {
	globalCache *cache.GatewayCache // reference to global gateway runtime cache
}

// NewRouteService constructor of RouteService, returns service pointer instance
func NewRouteService(c *cache.GatewayCache) *RouteService {
	return &RouteService{
		globalCache: c,
	}
}

// MatchRoute matches incoming request with cached valid routes
// Matching priority: exact match first, then longest prefix match
// ctx: request context
// path: request uri path from client
// method: HTTP request method
// return: matched route pointer, nil if no route matched; nil error always
func (s *RouteService) MatchRoute(ctx context.Context, host, path, method string) (*cache.CachedRoute, error) {
	route := s.globalCache.MatchRoute(host, path, method)
	return route, nil
}

// GetChannel 获取通道
func (s *RouteService) GetChannel(lcID string) *cache.CachedChannel {
	return s.globalCache.GetChannel(lcID)
}
