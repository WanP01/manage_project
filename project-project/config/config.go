package config

import (
	"log"
	"os"

	"project-common/logs"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

// AppConf 全局配置变量
var AppConf = InitConfig()

type Config struct {
	viper *viper.Viper
	Sc    *ServerConf
	Gc    *GrpcConf
	Ec    *EtcdConf
	Mc    *MysqlConf
	Jc    *JwtConf
}
type ServerConf struct {
	Name string
	Addr string
}

type GrpcConf struct {
	Name    string
	Addr    string
	Version string
	Weight  string
}

type EtcdConf struct {
	Addrs []string
}

type MysqlConf struct {
	Username string
	Password string
	Host     string
	Port     int
	Db       string
}

type JwtConf struct {
	AccessExp     int64
	RefreshExp    int64
	AccessSecret  string
	RefreshSecret string
}

func InitConfig() *Config {
	conf := &Config{viper: viper.New()}
	workDir, _ := os.Getwd()
	conf.viper.SetConfigName("config")
	conf.viper.SetConfigType("yaml")
	conf.viper.AddConfigPath("/etc/manage_project/project-project") // docker linux 环境下读取配置
	conf.viper.AddConfigPath(workDir + "/config")
	conf.viper.AddConfigPath(workDir + "/project-project/config") // windows 源代码下读取配置
	err := conf.viper.ReadInConfig()
	if err != nil {
		log.Fatalln(err)
		return nil
	}
	conf.InitServerConfig()
	conf.InitZapLog()
	conf.InitGrpcConfig()
	conf.InitEtcdConfig()
	conf.InitMysqlConfig()
	conf.InitJwtConfig()
	return conf
}

// InitServerConfig Server配置读取
func (c *Config) InitServerConfig() {
	sc := &ServerConf{
		Name: c.viper.GetString("server.name"),
		Addr: c.viper.GetString("server.addr"),
	}
	c.Sc = sc
}

// InitGrpcConfig GRPC服务配置读取
func (c *Config) InitGrpcConfig() {
	gc := &GrpcConf{
		Addr:    c.viper.GetString("grpc.addr"),
		Name:    c.viper.GetString("grpc.name"),
		Version: c.viper.GetString("grpc.version"),
		Weight:  c.viper.GetString("grpc.weight"),
	}
	c.Gc = gc
}

// InitEtcdConfig Etcd 配置读取
func (c *Config) InitEtcdConfig() {
	ec := &EtcdConf{}
	var addrs []string
	err := c.viper.UnmarshalKey("etcd.addrs", &addrs)
	if err != nil {
		log.Fatalln(err)
	}
	ec.Addrs = addrs
	c.Ec = ec
}

// InitZapLog Zaplog读取配置并初始化
func (c *Config) InitZapLog() {
	//从配置中读取日志配置，初始化日志
	lg := &logs.LogConfig{
		DebugFileName: c.viper.GetString("zap.debugFileName"),
		InfoFileName:  c.viper.GetString("zap.infoFileName"),
		WarnFileName:  c.viper.GetString("zap.warnFileName"),
		MaxSize:       c.viper.GetInt("zap.maxSize"),
		MaxAge:        c.viper.GetInt("zap.maxAge"),
		MaxBackups:    c.viper.GetInt("zap.MaxBackups"),
	}

	if err := logs.InitLogger(lg); err != nil {
		log.Fatalln(err)
	}
}

// InitRedisOptions Redis配置初始化
func (c *Config) InitRedisOptions() *redis.Options {
	return &redis.Options{
		Addr:     c.viper.GetString("redis.host") + ":" + c.viper.GetString("redis.port"),
		Password: c.viper.GetString("redis.password"),
		DB:       c.viper.GetInt("redis.db"),
	}
}

// InitMysqlConfig Mysql 配置初始化
func (c *Config) InitMysqlConfig() {
	mc := &MysqlConf{
		Username: c.viper.GetString("mysql.username"),
		Password: c.viper.GetString("mysql.password"),
		Host:     c.viper.GetString("mysql.host"),
		Port:     c.viper.GetInt("mysql.port"),
		Db:       c.viper.GetString("mysql.db"),
	}
	c.Mc = mc
}

// InitJwtConfig Jwt配置读取
func (c *Config) InitJwtConfig() {
	jc := &JwtConf{
		AccessExp:     c.viper.GetInt64("jwt.accessExp"),
		RefreshExp:    c.viper.GetInt64("jwt.refreshExp"),
		AccessSecret:  c.viper.GetString("jwt.accessSecret"),
		RefreshSecret: c.viper.GetString("jwt.refreshSecret"),
	}
	c.Jc = jc
}
