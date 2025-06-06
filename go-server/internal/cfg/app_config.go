package cfg

var (
	GlobalServerConfig *ApplicationConfig
)

// ApplicationConfig 配置结构体
type ApplicationConfig struct {
	// ServerConfig 服务器配置
	IP       string
	Port     string
	SecKey   string
	Debug    bool
	DbType   string
	DbConfig DbConfig `mapstructure:"db-config"`
}

// DbConfig 数据库配置
type DbConfig struct {
	Mysql  Mysql  `mapstructure:"mysql"`
	Sqlite Sqlite `mapstructure:"sqlite"`
}
