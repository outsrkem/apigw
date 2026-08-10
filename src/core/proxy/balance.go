package proxy

import (
	"apigw/src/cache"
	"errors"
)

// SelectTarget selects a downstream backend node according to load balancing strategy
// route: current matched api route configuration
// return: selected backend address string, error message if selection failed
func SelectTarget(lcID string) (string, error) {
	if lcID == "" {
		return "", errors.New("load channel id is empty")
	}

	if cache.GlobalCache == nil {
		return "", errors.New("gateway cache is uninitialized")
	}

	channel := cache.GlobalCache.GetChannel(lcID)
	if channel == nil {
		return "", errors.New("load channel not found")
	}

	idx := cache.GlobalCache.SelectNode(lcID)
	if idx < 0 || idx >= len(channel.Nodes) {
		return "", errors.New("no available backend nodes")
	}

	return channel.Nodes[idx], nil
}
