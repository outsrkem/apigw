package core

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
)

// SetReqHeaders 透传请求头
func SetReqHeaders(ctx context.Context, c *app.RequestContext) map[string]string {
	headers := make(map[string]string)
	c.Request.Header.VisitAll(func(key, value []byte) {
		headers[string(key)] = string(value)
	})
	return headers
}

// SetResHeaders 透传响应头
func SetResHeaders(ctx context.Context, c *app.RequestContext, header http.Header) {
	for key, value := range header {
		c.Response.Header.Set(key, strings.Join(value, ""))
	}
}

// Response http 请求返回的结果
type Response struct {
	StatusCode int // e.g. 200
	Body       []byte
	Header     http.Header
}

// SendHttpRequest 发送 HTTP 请求并安全处理响应
// 注意：返回的 responseBody 是已读取的完整内容，后续无需再操作原始响应体
func SendHttpRequest(ctx context.Context, method, url string, body []byte, headers map[string]string, timeout time.Duration) (*Response, error) {
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	_resp := &Response{}
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return _resp, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return _resp, err
	}

	// 先读取内容再关闭 Body
	defer func() {
		if resp != nil && resp.Body != nil {
			// 安全关闭并丢弃剩余内容（如果存在）
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	}()

	// 读取全部内容到内存
	_resp.Body, err = io.ReadAll(resp.Body)
	if err != nil {
		return _resp, err
	}

	_resp.Header = resp.Header
	_resp.StatusCode = resp.StatusCode
	return _resp, nil
}
