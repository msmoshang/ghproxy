package proxy

import (
	"context"
	"fmt"
	"ghproxy/config"
	"io"
	"net/http"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
)

var (
	respHeadersToRemove = map[string]struct{}{
		"Content-Security-Policy":   {},
		"Referrer-Policy":           {},
		"Strict-Transport-Security": {},
		"X-Github-Request-Id":       {},
		"X-Timer":                   {},
		"X-Served-By":               {},
		"X-Fastly-Request-Id":       {},
	}

	reqHeadersToRemove = map[string]struct{}{
		"CF-IPCountry":     {},
		"CF-RAY":           {},
		"CF-Visitor":       {},
		"CF-Connecting-IP": {},
		"CF-EW-Via":        {},
		"CDN-Loop":         {},
		// Git协议需要保留这些头信息
		// "Upgrade":          {},
		// "Connection":       {},
	}
)

func ChunkedProxyRequest(ctx context.Context, c *app.RequestContext, u string, cfg *config.Config, matcher string) {

	var (
		method []byte
		req    *http.Request
		resp   *http.Response
		err    error
	)

	method = c.Request.Method()

	req, err = client.NewRequest(string(method), u, c.Request.BodyStream())
	if err != nil {
		HandleError(c, fmt.Sprintf("Failed to create request: %v", err))
		return
	}

	setRequestHeaders(c, req)
	AuthPassThrough(c, cfg, req)

	resp, err = client.Do(req)
	if err != nil {
		HandleError(c, fmt.Sprintf("Failed to send request: %v", err))
		return
	}

	// 错误处理(404)
	if resp.StatusCode == 404 {
		ErrorPage(c, NewErrorWithStatusLookup(404, "Page Not Found (From Github)"))
		return
	}

	var (
		bodySize      int
		contentLength string
		sizelimit     int
	)
	sizelimit = cfg.Server.SizeLimit * 1024 * 1024
	contentLength = resp.Header.Get("Content-Length")
	if contentLength != "" {
		var err error
		bodySize, err = strconv.Atoi(contentLength)
		if err != nil {
			logWarning("%s %s %s %s %s Content-Length header is not a valid integer: %v", c.ClientIP(), c.Method(), c.Path(), c.UserAgent(), c.Request.Header.GetProtocol(), err)
			bodySize = -1
		}
		if err == nil && bodySize > sizelimit {
			var finalURL string
			finalURL = resp.Request.URL.String()
			err = resp.Body.Close()
			if err != nil {
				logError("Failed to close response body: %v", err)
			}
			c.Redirect(301, []byte(finalURL))
			logWarning("%s %s %s %s %s Final-URL: %s Size-Limit-Exceeded: %d", c.ClientIP(), c.Method(), c.Path(), c.UserAgent(), c.Request.Header.GetProtocol(), finalURL, bodySize)
			return
		}
	}

	// 复制响应头，排除需要移除的 header
	for key, values := range resp.Header {
		if _, shouldRemove := respHeadersToRemove[key]; !shouldRemove {
			for _, value := range values {
				c.Header(key, value)
			}
		}
	}

	switch cfg.Server.Cors {
	case "*":
		c.Header("Access-Control-Allow-Origin", "*")
	case "":
		c.Header("Access-Control-Allow-Origin", "*")
	case "nil":
		c.Header("Access-Control-Allow-Origin", "")
	default:
		c.Header("Access-Control-Allow-Origin", cfg.Server.Cors)
	}

	c.Status(resp.StatusCode)

	if MatcherShell(u) && matchString(matcher, matchedMatchers) && cfg.Shell.Editor {
		// 判断body是不是gzip
		var compress string
		if resp.Header.Get("Content-Encoding") == "gzip" {
			compress = "gzip"
		}

		logDebug("Use Shell Editor: %s %s %s %s %s", c.ClientIP(), method, u, c.Request.Header.Get("User-Agent"), c.Request.Header.GetProtocol())
		c.Header("Content-Length", "")

		var reader io.Reader

		reader, _, err = processLinks(resp.Body, compress, string(c.Request.Host()), cfg)
		c.SetBodyStream(reader, -1)
		if err != nil {
			logError("%s %s %s %s %s Failed to copy response body: %v", c.ClientIP(), method, u, c.Request.Header.Get("User-Agent"), c.Request.Header.GetProtocol(), err)
			ErrorPage(c, NewErrorWithStatusLookup(500, fmt.Sprintf("Failed to copy response body: %v", err)))
			return
		}
	} else {
		if contentLength != "" {
			c.SetBodyStream(resp.Body, bodySize)
			return
		}
		c.SetBodyStream(resp.Body, -1)
	}

}
