package service

import (
	"apigw/src/cache"
	"apigw/src/models"
	"context"
	"strings"
)

// Route matching mode constants
const (
	RouteModeExact  = "EXACT"  // Exact path matching mode
	RouteModePrefix = "PREFIX" // Prefix longest matching mode
	RouteMethodAny  = "Any"    // Represent matching all HTTP request methods
)

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
func (s *RouteService) MatchRoute(ctx context.Context, path, method string) (*models.OrmApiInterface, error) {
	s.globalCache.RLock()
	defer s.globalCache.RUnlock()

	allRoutes := s.globalCache.GetAllActiveRoutes()
	var prefixMatchRoute *models.OrmApiInterface

	// Step 1: Exact match with highest priority
	for _, item := range allRoutes {
		if item.Method != RouteMethodAny && item.Method != method {
			continue
		}
		if item.Mode == RouteModeExact && item.ReqUri == path {
			return item, nil
		}
	}

	// Step 2: Longest prefix matching
	for _, item := range allRoutes {
		if item.Method != RouteMethodAny && item.Method != method {
			continue
		}
		if item.Mode == RouteModePrefix && strings.HasPrefix(path, item.ReqUri) {
			// update to longer prefix route
			if prefixMatchRoute == nil || len(item.ReqUri) > len(prefixMatchRoute.ReqUri) {
				prefixMatchRoute = item
			}
		}
	}
	return prefixMatchRoute, nil
}

// RouteSvc global singleton RouteService instance
var RouteSvc *RouteService
