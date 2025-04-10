# GHProxy

![pull](https://img.shields.io/docker/pulls/wjqserver/ghproxy.svg)![Docker Image Size (tag)](https://img.shields.io/docker/image-size/wjqserver/ghproxy/latest)[![Go Report Card](https://goreportcard.com/badge/github.com/WJQSERVER-STUDIO/ghproxy)](https://goreportcard.com/report/github.com/WJQSERVER-STUDIO/ghproxy)

使用Go实现的GHProxy,用于加速部分地区Github仓库的拉取,支持速率限制,用户鉴权,支持Docker部署

## 项目说明

### 项目特点

- ⚡ **基于 Go 语言实现，跨平台的同时提供高并发性能**
- 🌐 **使用字节旗下的 [HertZ](https://github.com/cloudwego/hertz) 作为 Web 框架**
- 📡 **使用 [Touka-HTTPC](https://github.com/satomitouka/touka-httpc) 作为 HTTP 客户端**
- 📥 **支持 Git clone、raw、releases 等文件拉取**
- 🎨 **支持多个前端主题**
- 🚫 **支持自定义黑名单/白名单**
- 🗄️ **支持 Git Clone 缓存（配合 [Smart-Git](https://github.com/WJQSERVER-STUDIO/smart-git)）**
- 🐳 **支持 Docker 部署**
- ⚡ **支持速率限制**
- 🔒 **支持用户鉴权**
- 🐚 **支持 shell 脚本嵌套加速**

此项目基于[WJQSERVER-STUDIO/ghproxy: 基于Go的高性能,多功能,可扩展的Github代理](https://github.com/WJQSERVER-STUDIO/ghproxy)赞助原项目

爱发电: https://afdian.com/a/wjqserver

USDT(TRC20): `TNfSYG6F2vkiibd6J6mhhHNWDgWgNdF5hN`

### 捐赠列表

| 赞助人 | 金额           |
| ------ | -------------- |
| starry | 8 USDT (TRC20) |

### 新增

- 新增一个新主题
- 新增查询shell嵌套加速api
- 增加自定义404页面选项

### 待办

- [ ] 新增hf镜像下载加速
