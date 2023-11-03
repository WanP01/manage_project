package config

import (
	"bytes"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"log"
	"os"

	"project-common/logs"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

// AppConf 全局配置变量
var AppConf = InitConfig()

type Config struct {
	viper   *viper.Viper
	Sc      *ServerConf
	Gc      *GrpcConf
	Ec      *EtcdConf
	Mc      *MysqlConf
	Jc      *JwtConf
	Dc      *DbConf
	Kc      *KafkaConf
	JaegerC *JaegerConfig
}
type ServerConf struct {
	Name string
	Addr string
}

type GrpcConf struct {
	Name     string
	Addr     string
	EtcdAddr string
	Version  string
	Weight   string
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
	Name     string
}

// 主从复制
type DbConf struct {
	Master     MysqlConf
	Slave      []MysqlConf
	Separation bool
}

type JwtConf struct {
	AccessExp     int64
	RefreshExp    int64
	AccessSecret  string
	RefreshSecret string
}

type KafkaConf struct {
	Addr  []string
	Group string
	Topic string
}

type JaegerConfig struct {
	Endpoints string
}

func InitConfig() *Config {
	conf := &Config{viper: viper.New()}
	conf.viper.SetConfigType("yaml") //提前设置读取的文件类型
	//先从nacos读取配置，如果读取不到 在本地读取
	nacosClient := InitNacosClient() //连接
	configYaml, err2 := nacosClient.confClient.GetConfig(vo.ConfigParam{
		DataId: nacosClient.NacosConf.DataId,
		Group:  nacosClient.NacosConf.Group,
	}) //读取string类型的配置文件
	if err2 != nil {
		log.Fatalln(err2)
	}

	//监听配置变化
	err2 = nacosClient.confClient.ListenConfig(vo.ConfigParam{
		DataId: nacosClient.NacosConf.DataId,
		Group:  nacosClient.NacosConf.Group,
		OnChange: func(namespace, group, dataId, data string) {
			//重新读取nacos的最新文件
			log.Printf("load nacos config changed %s \n", data)
			err := conf.viper.ReadConfig(bytes.NewBuffer([]byte(data)))
			if err != nil {
				log.Printf("load nacos config changed err : %s \n", err.Error())
			}
			//所有的配置应该重新读取
			conf.ReLoadAllConfig()
		},
	})
	if err2 != nil {
		log.Fatalln(err2)
	}

	if configYaml != "" {
		err := conf.viper.ReadConfig(bytes.NewBuffer([]byte(configYaml)))
		if err != nil {
			log.Fatalln(err)
		}
		log.Printf("load nacos config %s \n", configYaml)
	} else { //如果远程读取不到，就在本地读取
		workDir, _ := os.Getwd()
		conf.viper.SetConfigName("config")                              //设置文件名称
		conf.viper.AddConfigPath("/etc/manage_project/project-project") // docker linux 环境下读取配置
		conf.viper.AddConfigPath(workDir + "/config")
		conf.viper.AddConfigPath(workDir + "/project-project/config") // windows 源代码下读取配置
		err := conf.viper.ReadInConfig()
		if err != nil {
			log.Fatalln(err)
			return nil
		}
	}

	//配置载入conf =》 Appconf
	conf.ReLoadAllConfig()
	return conf
}

func (c *Config) ReLoadAllConfig() {
	//读取最新配置
	c.InitServerConfig()
	c.InitZapLog()
	c.InitGrpcConfig()
	c.InitEtcdConfig()
	c.InitMysqlConfig()
	c.InitJwtConfig()
	c.InitDbConfig()
	c.InitKafkaConfig()
	c.InitJaegerConfig()

	//重新创建相关的客户端
	c.ReConnRedis()
	c.ReConnMysql()
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
		Addr:     c.viper.GetString("grpc.addr"),
		Name:     c.viper.GetString("grpc.name"),
		EtcdAddr: c.viper.GetString("grpc.etcdAddr"),
		Version:  c.viper.GetString("grpc.version"),
		Weight:   c.viper.GetString("grpc.weight"),
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

func (c *Config) InitDbConfig() {
	mc := &DbConf{}
	mc.Separation = c.viper.GetBool("db.separation")
	var slaves []MysqlConf
	err := c.viper.UnmarshalKey("db.slave", &slaves)
	if err != nil {
		panic(err)
	}
	master := MysqlConf{
		Username: c.viper.GetString("db.master.username"),
		Password: c.viper.GetString("db.master.password"),
		Host:     c.viper.GetString("db.master.host"),
		Port:     c.viper.GetInt("db.master.port"),
		Db:       c.viper.GetString("db.master.db"),
	}
	mc.Master = master
	mc.Slave = slaves
	c.Dc = mc
}

func (c *Config) InitKafkaConfig() {
	kc := &KafkaConf{}
	var addr []string
	err := c.viper.UnmarshalKey("kafka.addr", &addr)
	if err != nil {
		log.Fatalln(err)
	}
	kc.Addr = addr
	kc.Topic = c.viper.GetString("kafka.topic")
	kc.Group = c.viper.GetString("kafka.group")
	c.Kc = kc
}

func (c *Config) InitJaegerConfig() {
	mc := &JaegerConfig{
		Endpoints: c.viper.GetString("jaeger.endpoints"),
	}
	c.JaegerC = mc
}
