package proxy

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/network/standard"
	"github.com/cloudwego/hertz/pkg/protocol"
)

var (
	httpCli  *client.Client     // httpCli global shared plain HTTP client instance
	CacheTTL = 30 * time.Minute // Expiration duration of idle cached client
	MaxSize  = 500              // Maximum capacity of client cache pool
)

// cachedClient stores client instance and last access timestamp for LRU eviction
type cachedClient struct {
	client   *client.Client
	lastUsed time.Time
}

var (
	cliCache map[string]*cachedClient // Key: lcCaID, Value: cached TLS client
	mu       sync.RWMutex             // Read-write lock for cache concurrent access
	stats    struct{ hits, misses int64 }
	stopCh   chan struct{} // Signal channel to terminate cleanup goroutine
)

func init() {
	var err error
	httpCli, err = client.NewClient()
	if err != nil {
		panic(fmt.Sprintf("initialize http client failed: %s", err.Error()))
	}
	cliCache = make(map[string]*cachedClient)
	stopCh = make(chan struct{})
	go cleanupLoop()
}

// SetReqHeaders converts raw Hertz request header to string key-value map
func SetReqHeaders(c *app.RequestContext) map[string]string {
	headerMap := make(map[string]string)
	c.Request.Header.VisitAll(func(key, value []byte) {
		k := string(key)
		v := string(value)
		if existVal, ok := headerMap[k]; ok {
			headerMap[k] = existVal + "," + v
		} else {
			headerMap[k] = v
		}
	})
	return headerMap
}

// GetHttpsCli gets TLS client from cache, creates new client when cache miss
func GetHttpsCli(ctx context.Context, lcCaID string) (*client.Client, error) {
	// Read lock fast query cache
	mu.RLock()
	item, ok := cliCache[lcCaID]
	mu.RUnlock()
	if ok {
		mu.Lock()
		item.lastUsed = time.Now()
		stats.hits++
		mu.Unlock()
		return item.client, nil
	}

	// Exclusive lock for client creation
	mu.Lock()
	defer mu.Unlock()

	// Double check to avoid repeated creation in concurrent scenario
	if item, ok := cliCache[lcCaID]; ok {
		item.lastUsed = time.Now()
		stats.hits++
		return item.client, nil
	}

	tlsCfg, err := BuildtlsCfg(lcCaID)
	if err != nil {
		return nil, err
	}

	cli, err := client.NewClient(
		client.WithDialer(standard.NewDialer()),
		client.WithTLSConfig(tlsCfg),
	)
	if err != nil {
		stats.misses++
		return nil, err
	}

	// Evict least recently used element when cache reaches capacity limit
	if len(cliCache) >= MaxSize {
		evictOne()
	}

	cliCache[lcCaID] = &cachedClient{
		client:   cli,
		lastUsed: time.Now(),
	}
	stats.misses++

	return cli, nil
}

// Stats returns runtime cache hit & miss statistics
func Stats() map[string]interface{} {
	mu.RLock()
	defer mu.RUnlock()
	return map[string]interface{}{
		"total":  len(cliCache),
		"hits":   stats.hits,
		"misses": stats.misses,
	}
}

// cleanupLoop periodically cleans expired idle cached clients
func cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			mu.Lock()
			now := time.Now()
			for uuid, item := range cliCache {
				if now.Sub(item.lastUsed) > CacheTTL {
					delete(cliCache, uuid)
				}
			}
			mu.Unlock()
		case <-stopCh:
			return
		}
	}
}

// Close sends termination signal to stop periodic cleanup goroutine
func Close() {
	close(stopCh)
}

// evictOne removes the least recently used client from cache pool
func evictOne() {
	var oldest string
	var oldestTime time.Time
	for uuid, item := range cliCache {
		if oldest == "" || item.lastUsed.Before(oldestTime) {
			oldest = uuid
			oldestTime = item.lastUsed
		}
	}
	if oldest != "" {
		delete(cliCache, oldest)
	}
}

// HeProxy forwards request via plain HTTP or customized TLS HTTPS client
func HeProxy(ctx context.Context, method string, proto, LcCaID, url string, body []byte, headers map[string]string) (*protocol.Response, error) {
	var cli *client.Client
	var err error

	if proto == "HTTP" {
		cli = httpCli
	} else {
		cli, err = GetHttpsCli(ctx, LcCaID)
		if err != nil {
			return nil, err
		}
	}

	req := protocol.AcquireRequest()
	defer protocol.ReleaseRequest(req)

	req.SetMethod(method)
	req.SetRequestURI(url)
	req.SetHeaders(headers)
	req.SetBody(body)

	resp := protocol.AcquireResponse()
	err = cli.Do(ctx, req, resp)
	if err != nil {
		protocol.ReleaseResponse(resp)
		return nil, fmt.Errorf("request execute failed: %w", err)
	}

	return resp, nil
}
