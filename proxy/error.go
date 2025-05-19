package proxy

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/WJQSERVER-STUDIO/logger"
	"github.com/cloudwego/hertz/pkg/app"
)

// 日志模块
var (
	logw       = logger.Logw
	logDump    = logger.LogDump
	logDebug   = logger.LogDebug
	logInfo    = logger.LogInfo
	logWarning = logger.LogWarning
	logError   = logger.LogError
)

func HandleError(c *app.RequestContext, message string) {
	ErrorPage(c, NewErrorWithStatusLookup(500, message))
	logError(message)
}

type GHProxyErrors struct {
	StatusCode   int
	StatusDesc   string
	StatusText   string
	HelpInfo     string
	ErrorMessage string
}

var (
	ErrInvalidURL = &GHProxyErrors{
		StatusCode: 400,
		StatusDesc: "Bad Request",
		StatusText: "无效请求",
		HelpInfo:   "请求的URL格式不正确，请检查后重试。",
	}
	ErrAuthHeaderUnavailable = &GHProxyErrors{
		StatusCode: 401,
		StatusDesc: "Unauthorized",
		StatusText: "认证失败",
		HelpInfo:   "缺少或无效的鉴权信息。",
	}
	ErrForbidden = &GHProxyErrors{
		StatusCode: 403,
		StatusDesc: "Forbidden",
		StatusText: "权限不足",
		HelpInfo:   "您没有权限访问此资源。",
	}
	ErrNotFound = &GHProxyErrors{
		StatusCode: 404,
		StatusDesc: "Not Found",
		StatusText: "页面未找到",
		HelpInfo:   "抱歉，您访问的页面不存在。",
	}
	ErrTooManyRequests = &GHProxyErrors{
		StatusCode: 429,
		StatusDesc: "Too Many Requests",
		StatusText: "请求过于频繁",
		HelpInfo:   "您的请求过于频繁，请稍后再试。",
	}
	ErrInternalServerError = &GHProxyErrors{
		StatusCode: 500,
		StatusDesc: "Internal Server Error",
		StatusText: "服务器内部错误",
		HelpInfo:   "服务器处理您的请求时发生错误，请稍后重试或联系管理员。",
	}
)

var statusErrorMap map[int]*GHProxyErrors

func init() {
	statusErrorMap = map[int]*GHProxyErrors{
		ErrInvalidURL.StatusCode:            ErrInvalidURL,
		ErrAuthHeaderUnavailable.StatusCode: ErrAuthHeaderUnavailable,
		ErrForbidden.StatusCode:             ErrForbidden,
		ErrNotFound.StatusCode:              ErrNotFound,
		ErrTooManyRequests.StatusCode:       ErrTooManyRequests,
		ErrInternalServerError.StatusCode:   ErrInternalServerError,
	}
}

func NewErrorWithStatusLookup(statusCode int, errMsg string) *GHProxyErrors {
	baseErr, found := statusErrorMap[statusCode]

	if found {
		return &GHProxyErrors{
			StatusCode:   baseErr.StatusCode,
			StatusDesc:   baseErr.StatusDesc,
			StatusText:   baseErr.StatusText,
			HelpInfo:     baseErr.HelpInfo,
			ErrorMessage: errMsg,
		}
	} else {
		return &GHProxyErrors{
			StatusCode:   statusCode,
			ErrorMessage: errMsg,
		}
	}
}

var (
	errPagesFs    fs.FS
	custom404Path string
	hasCustom404  bool
)

// InitErrPagesFS 初始化错误页面文件系统
// 支持内置页面和自定义404页面
func InitErrPagesFS(pages fs.FS) error {
	var err error
	errPagesFs, err = fs.Sub(pages, "pages/error")
	if err != nil {
		return err
	}
	return nil
}

// SetCustom404Page 设置自定义404页面路径
func SetCustom404Page(path string) {
	if path != "" {
		absPath, err := filepath.Abs(path)
		if err != nil {
			logError("无法解析自定义404页面路径: %v", err)
			return
		}

		// 检查文件是否存在
		if _, err := os.Stat(absPath); err == nil {
			custom404Path = absPath
			hasCustom404 = true
			logInfo("已设置自定义404页面: %s", absPath)
		} else {
			logWarning("自定义404页面不存在: %s, 将使用内置404页面", absPath)
		}
	}
}

type ErrorPageData struct {
	StatusCode   int
	StatusDesc   string
	StatusText   string
	HelpInfo     string
	ErrorMessage string
}

func ErrPageUnwarper(errInfo *GHProxyErrors) ErrorPageData {
	return ErrorPageData{
		StatusCode:   errInfo.StatusCode,
		StatusDesc:   errInfo.StatusDesc,
		StatusText:   errInfo.StatusText,
		HelpInfo:     errInfo.HelpInfo,
		ErrorMessage: errInfo.ErrorMessage,
	}
}

func ErrorPage(c *app.RequestContext, errInfo *GHProxyErrors) {
	// 如果是404错误且有自定义404页面，则使用自定义页面
	if errInfo.StatusCode == 404 && hasCustom404 {
		pageData, err := os.ReadFile(custom404Path)
		if err != nil {
			logError("读取自定义404页面失败: %v, 将使用内置404页面", err)
		} else {
			c.Data(errInfo.StatusCode, "text/html; charset=utf-8", pageData)
			return
		}
	}

	// 使用内置错误页面模板
	pageData, err := htmlTemplateRender(errPagesFs, ErrPageUnwarper(errInfo))
	if err != nil {
		c.JSON(errInfo.StatusCode, map[string]string{"error": errInfo.ErrorMessage})
		logDebug("Error reading page.tmpl: %v", err)
		return
	}
	c.Data(errInfo.StatusCode, "text/html; charset=utf-8", pageData)
	return
}

func htmlTemplateRender(fsys fs.FS, data interface{}) ([]byte, error) {
	tmplPath := "404.html"
	tmpl, err := template.ParseFS(fsys, tmplPath)
	if err != nil {
		return nil, fmt.Errorf("error parsing template: %w", err)
	}
	if tmpl == nil {
		return nil, fmt.Errorf("template is nil")
	}

	// 创建一个 bytes.Buffer 用于存储渲染结果
	var buf bytes.Buffer

	err = tmpl.Execute(&buf, data)
	if err != nil {
		return nil, fmt.Errorf("error executing template: %w", err)
	}

	// 返回 buffer 的内容作为 []byte
	return buf.Bytes(), nil
}
