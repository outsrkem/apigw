package proxy

import (
	"log"
	"net/http"
	"strings"
)

type httpProxy struct {
	Headers map[string]string
	Method  string
	Url     string
	Body    *strings.Reader
}

func NewProxy() (*httpProxy, error) {
	proxy := &httpProxy{}
	return proxy, nil
}

func (r *httpProxy) NewProxyRes(headers map[string]string, method, url string, body *strings.Reader) (*httpProxy, error) {
	// 可以在这里对参数进行判断
	res := &httpProxy{
		Headers: headers,
		Method:  method,
		Url:     url,
		Body:    body,
	}
	return res, nil
}

// 向后端接口发送http请求
func (r *httpProxy) DoHttpV1(res *httpProxy) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest(res.Method, res.Url, res.Body)
	if err != nil {
		log.Println("Get error:", nil)
		return nil, err
	}
	// 设置请求头
	for key, header := range res.Headers {
		req.Header.Set(key, header)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	// 这个defer会导致不能返回数据
	// defer resp.Body.Close()
	return resp, nil
}
