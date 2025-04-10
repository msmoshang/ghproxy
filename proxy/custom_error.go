package proxy

import (
	"context"
	"fmt"
	"ghproxy/config"
	"io/fs"
	"net/http"
	"os"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/adaptor"
)

// 全局错误处理器
var errorHandler *ErrorHandler

// ErrorHandler 处理自定义错误页面
type ErrorHandler struct {
	cfg        *config.Config
	errorPages fs.FS
}

// NewErrorHandler 创建一个新的错误处理器
func NewErrorHandler(cfg *config.Config, errorPages fs.FS) *ErrorHandler {
	return &ErrorHandler{
		cfg:        cfg,
		errorPages: errorPages,
	}
}

// InitErrorHandler 初始化全局错误处理器
func InitErrorHandler(cfg *config.Config, errorPages fs.FS) {
	errorHandler = NewErrorHandler(cfg, errorPages)
	logInfo("Error handler initialized with custom404: %s", cfg.Pages.Custom404)
}

// Handle404Error 处理404错误，显示自定义页面
func (h *ErrorHandler) Handle404Error(ctx context.Context, c *app.RequestContext, message string, path string) {
	// 记录错误日志
	logWarning("%s %s %s %s %s Invalid URL Format. Path: %s",
		c.ClientIP(), c.Method(), path, c.Request.Header.UserAgent(),
		c.Request.Header.GetProtocol(), path)

	// 检查是否配置了自定义404页面
	if h.cfg.Pages.Custom404 != "" {
		// 使用外部自定义404页面
		if _, err := os.Stat(h.cfg.Pages.Custom404); err == nil {
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.Status(http.StatusNotFound)

			// 读取自定义404页面内容
			content, err := os.ReadFile(h.cfg.Pages.Custom404)
			if err != nil {
				// 如果读取失败，回退到默认错误消息
				logError("Failed to read custom 404 page: %v", err)
				c.String(http.StatusNotFound, "Invalid URL Format. Path: %s", path)
				return
			}

			// 返回自定义404页面内容
			c.Data(http.StatusNotFound, "text/html; charset=utf-8", content)
			return
		}

		// 如果配置的文件不存在，记录错误并回退到内置404页面
		logWarning("Custom 404 page not found at: %s", h.cfg.Pages.Custom404)
	}

	// 使用内置404页面
	if h.errorPages != nil {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.Status(http.StatusNotFound)

		// 使用http.FileServer提供内置404页面
		staticServer := http.FileServer(http.FS(h.errorPages))
		req, err := adaptor.GetCompatRequest(&c.Request)
		if err != nil {
			logError("%s", err)
			c.String(http.StatusNotFound, "Invalid URL Format. Path: %s", path)
			return
		}

		// 修改请求路径为404.html
		req.URL.Path = "/404.html"

		// 使用ResponseWriter适配器提供响应
		staticServer.ServeHTTP(adaptor.GetCompatResponseWriter(&c.Response), req)
		return
	}

	// 如果内置404页面不可用，回退到默认错误消息
	c.String(http.StatusNotFound, "Invalid URL Format. Path: %s", path)
}

// HandleInvalidURL 处理无效URL格式错误
func (h *ErrorHandler) HandleInvalidURL(ctx context.Context, c *app.RequestContext, path string) {
	h.Handle404Error(ctx, c, fmt.Sprintf("Invalid URL Format. Path: %s", path), path)
}
