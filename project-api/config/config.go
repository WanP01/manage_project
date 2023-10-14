package config

import (
	"log"
	"os"
	"project-common/logs"

	"github.com/spf13/viper"
)

var AppConf = InitConfig()

type Config struct {
	viper *viper.Viper
	Sc    *ServerConf
	Ec    *EtcdConf
}
type ServerConf struct {
	Name string
	Addr string
}

type EtcdConf struct {
	Addrs []string
}

func InitConfig() *Config {
	conf := &Config{viper: viper.New()}
	workDir, _ := os.Getwd()
	conf.viper.SetConfigName("config")
	conf.viper.SetConfigType("yaml")
	conf.viper.AddConfigPath("/etc/manage_project/project-api")
	conf.viper.AddConfigPath(workDir + "/project-api/config")
	err := conf.viper.ReadInConfig()
	if err != nil {
		log.Fatalln(err)
		return nil
	}
	conf.InitServerConfig()
	conf.InitZapLog()
	conf.InitEtcdConfig()
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

// Etcd 配置读取
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
