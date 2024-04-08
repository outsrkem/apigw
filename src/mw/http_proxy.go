package mw

import (
	"fmt"
	"net/http"
	"strings"
)

// 向后端接口发送http请求
func DoHttp(headers map[string]string, method string, url string, raw *strings.Reader) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, url, raw)
	if err != nil {
		fmt.Println("Get error:", nil)
		return nil, err
	}
	// 设置请求头
	for key, header := range headers {
		req.Header.Set(key, header)
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	// 这个defer会导致不能返回数据
	// defer resp.Body.Close()
	return resp, nil
}
