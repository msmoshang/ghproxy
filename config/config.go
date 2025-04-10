package config

import (
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Server    ServerConfig
	Httpc     HttpcConfig
	GitClone  GitCloneConfig
	Shell     ShellConfig
	Pages     PagesConfig
	Log       LogConfig
	Auth      AuthConfig
	Blacklist BlacklistConfig
	Whitelist WhitelistConfig
	RateLimit RateLimitConfig
	Outbound  OutboundConfig
}

/*
[server]
host = "0.0.0.0"  # 监听地址
port = 8080  # 监听端口
sizeLimit = 125 # 125MB
H2C = true # 是否开启H2C传输
*/

type ServerConfig struct {
	Port      int    `toml:"port"`
	Host      string `toml:"host"`
	SizeLimit int    `toml:"sizeLimit"`
	MemLimit  int64  `toml:"memLimit"`
	H2C       bool   `toml:"H2C"`
	Cors      string `toml:"cors"`
	Debug     bool   `toml:"debug"`
}

/*
[httpc]
mode = "auto" # "auto" or "advanced"
maxIdleConns = 100 # only for advanced mode
maxIdleConnsPerHost = 60 # only for advanced mode
maxConnsPerHost = 0 # only for advanced mode
*/
type HttpcConfig struct {
	Mode                string `toml:"mode"`
	MaxIdleConns        int    `toml:"maxIdleConns"`
	MaxIdleConnsPerHost int    `toml:"maxIdleConnsPerHost"`
	MaxConnsPerHost     int    `toml:"maxConnsPerHost"`
}

/*
[gitclone]
mode = "bypass" # bypass / cache
smartGitAddr = "http://127.0.0.1:8080"
ForceH2C = true
*/
type GitCloneConfig struct {
	Mode         string `toml:"mode"`
	SmartGitAddr string `toml:"smartGitAddr"`
	ForceH2C     bool   `toml:"ForceH2C"`
}

/*
[shell]
editor = true
rewriteAPI = false
*/
type ShellConfig struct {
	Editor     bool `toml:"editor"`
	RewriteAPI bool `toml:"rewriteAPI"`
}

/*
[pages]
mode = "internal" # "internal" or "external"
theme = "bootstrap" # "bootstrap" or "nebula"
staticDir = "/data/www"
*/
type PagesConfig struct {
	Mode      string `toml:"mode"`
	Theme     string `toml:"theme"`
	StaticDir string `toml:"staticDir"`
}

type LogConfig struct {
	LogFilePath  string `toml:"logFilePath"`
	MaxLogSize   int    `toml:"maxLogSize"`
	Level        string `toml:"level"`
	HertZLogPath string `toml:"hertzLogPath"`
}

/*
[auth]
Method = "parameters" # "header" or "parameters"
Key = ""
Token = "token"
enabled = false
passThrough = false
ForceAllowApi = true
*/
type AuthConfig struct {
	Enabled       bool   `toml:"enabled"`
	Method        string `toml:"method"`
	Key           string `toml:"key"`
	Token         string `toml:"token"`
	PassThrough   bool   `toml:"passThrough"`
	ForceAllowApi bool   `toml:"ForceAllowApi"`
}

type BlacklistConfig struct {
	Enabled       bool   `toml:"enabled"`
	BlacklistFile string `toml:"blacklistFile"`
}

type WhitelistConfig struct {
	Enabled       bool   `toml:"enabled"`
	WhitelistFile string `toml:"whitelistFile"`
}

type RateLimitConfig struct {
	Enabled       bool   `toml:"enabled"`
	RateMethod    string `toml:"rateMethod"`
	RatePerMinute int    `toml:"ratePerMinute"`
	Burst         int    `toml:"burst"`
}

/*
[outbound]
enabled = false
url = "socks5://127.0.0.1:1080" # "http://127.0.0.1:7890"
*/
type OutboundConfig struct {
	Enabled bool   `toml:"enabled"`
	Url     string `toml:"url"`
}

// LoadConfig 从 TOML 配置文件加载配置
func LoadConfig(filePath string) (*Config, error) {
	if !FileExists(filePath) {
		// 楔入配置文件
		err := DefaultConfig().WriteConfig(filePath)
		if err != nil {
			return nil, err
		}
		return DefaultConfig(), nil
	}

	var config Config
	if _, err := toml.DecodeFile(filePath, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// 写入配置文件
func (c *Config) WriteConfig(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	return encoder.Encode(c)
}

// 检测文件是否存在
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// 默认配置结构体
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:      8080,
			Host:      "0.0.0.0",
			SizeLimit: 1024,
			MemLimit:  0,
			H2C:       true,
			Cors:      "*",
			Debug:     false,
		},
		Httpc: HttpcConfig{
			Mode:                "auto",
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 60,
			MaxConnsPerHost:     0,
		},
		GitClone: GitCloneConfig{
			Mode:         "bypass",
			SmartGitAddr: "http://127.0.0.1:8080",
			ForceH2C:     false,
		},
		Shell: ShellConfig{
			Editor:     false,
			RewriteAPI: false,
		},
		Pages: PagesConfig{
			Mode:      "internal",
			Theme:     "aurora",
			StaticDir: "/data/www",
		},
		Log: LogConfig{
			LogFilePath:  "/data/ghproxy/log/ghproxy.log",
			MaxLogSize:   10,
			Level:        "info",
			HertZLogPath: "/data/ghproxy/log/hertz.log",
		},
		Auth: AuthConfig{
			Enabled:       false,
			Method:        "parameters",
			Key:           "",
			Token:         "token",
			PassThrough:   false,
			ForceAllowApi: false,
		},
		Blacklist: BlacklistConfig{
			Enabled:       false,
			BlacklistFile: "/data/ghproxy/config/blacklist.txt",
		},
		Whitelist: WhitelistConfig{
			Enabled:       false,
			WhitelistFile: "/data/ghproxy/config/whitelist.txt",
		},
		RateLimit: RateLimitConfig{
			Enabled:       false,
			RateMethod:    "total",
			RatePerMinute: 100,
			Burst:         10,
		},
		Outbound: OutboundConfig{
			Enabled: false,
			Url:     "socks5://127.0.0.1:1080",
		},
	}
}
