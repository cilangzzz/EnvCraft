package cfg

var GlobalServerConfig *ServerConfig

// ServerConfig 配置结构体
type ServerConfig struct {
	IP     string
	Port   string
	SecKey string
	Debug  bool
}
