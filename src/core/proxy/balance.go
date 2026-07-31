package proxy

import (
	"apigw/src/cache"
	"apigw/src/models"
	"errors"
	"strings"
)

// SelectTarget selects a downstream backend node according to load balancing strategy
// route: current matched api route configuration
// return: selected backend address string, error message if selection failed
func SelectTarget(route *models.OrmApiInterface) (string, error) {
	lcID := route.LcID
	if lcID == "" {
		return "", errors.New("load channel id empty")
	}

	if cache.GlobalCache == nil {
		return "", errors.New("gateway cache uninitialized")
	}
	channel := cache.GlobalCache.GetChannelByLCID(lcID)
	if channel == nil {
		return "", errors.New("load channel not found")
	}
	if channel.Status != 1 {
		return "", errors.New("load channel disabled")
	}

	backendStr := strings.TrimSpace(channel.Backend)
	if backendStr == "" {
		return "", errors.New("no backend nodes configured")
	}

	// Filter blank strings to get valid available backend node list
	var validNodes []string
	for _, node := range strings.Split(backendStr, ",") {
		node = strings.TrimSpace(node)
		if node != "" {
			validNodes = append(validNodes, node)
		}
	}
	nodeCount := len(validNodes)
	if nodeCount == 0 {
		return "", errors.New("valid backend nodes empty")
	}

	idx := cache.GlobalCache.SelectNodeIndex(lcID, nodeCount)
	if idx < 0 || idx >= nodeCount {
		return "", errors.New("select backend index invalid")
	}

	return validNodes[idx], nil
}
