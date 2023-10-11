package config

import (
	"log"
	"os"

	"project-common/logs"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

var AppConf = InitConfig()

type Config struct {
	viper *viper.Viper
	Sc    *ServerConf
	Gc    *GrpcConf
}
type ServerConf struct {
	Name string
	Addr string
}

type GrpcConf struct {
	Name string
	Addr string
}

func InitConfig() *Config {
	conf := &Config{viper: viper.New()}
	workDir, _ := os.Getwd()
	conf.viper.SetConfigName("config")
	conf.viper.SetConfigType("yaml")
	// conf.viper.AddConfigPath("/etc/manage_project/project-user")
	conf.viper.AddConfigPath(workDir + "/project-user/config")
	err := conf.viper.ReadInConfig()
	if err != nil {
		log.Fatalln(err)
		return nil
	}
	conf.InitServerConfig()
	conf.InitZapLog()
	conf.InitGrpcConfig()
	return conf
}

// Server配置读取
func (c *Config) InitServerConfig() {
	sc := &ServerConf{
		Name: c.viper.GetString("server.name"),
		Addr: c.viper.GetString("server.addr"),
	}
	c.Sc = sc
}

// GRPC服务配置读取
func (c *Config) InitGrpcConfig() {
	gc := &GrpcConf{
		Addr: c.viper.GetString("grpc.addr"),
		Name: c.viper.GetString("grpc.name"),
	}
	c.Gc = gc
}

// Zaplog读取配置并初始化
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

// Redis配置初始化
func (c *Config) InitRedisOptions() *redis.Options {
	return &redis.Options{
		Addr:     c.viper.GetString("redis.host") + ":" + c.viper.GetString("redis.port"),
		Password: c.viper.GetString("redis.password"),
		DB:       c.viper.GetInt("redis.db"),
	}
}