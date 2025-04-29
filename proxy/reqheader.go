package proxy

import (
	"net/http"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
)

func setRequestHeaders(c *app.RequestContext, req *http.Request) {
	c.Request.Header.VisitAll(func(key, value []byte) {
		headerKey := string(key)
		headerValue := string(value)
		if _, shouldRemove := reqHeadersToRemove[headerKey]; !shouldRemove {
			req.Header.Set(headerKey, headerValue)
		}
	})

	// 确保Git协议所需的头信息被保留
	// 对于Git请求，强制保留Upgrade和Connection头
	if strings.Contains(req.URL.Path, "/git-upload-pack") || strings.Contains(req.URL.Path, "/git-receive-pack") {
		if c.Request.Header.Get("Upgrade") != "" {
			req.Header.Set("Upgrade", c.Request.Header.Get("Upgrade"))
		}
		if c.Request.Header.Get("Connection") != "" {
			req.Header.Set("Connection", c.Request.Header.Get("Connection"))
		}
	}
}
