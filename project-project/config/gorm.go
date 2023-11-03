package config

import (
	"fmt"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
	"project-project/internal/database/gorms"
)

var _db *gorm.DB

func (c *Config) ReConnMysql() {
	if c.Dc.Separation { // 开启读写分离
		//master
		master_username := c.Dc.Master.Username //账号
		master_password := c.Dc.Master.Password //密码
		master_host := c.Dc.Master.Host         //数据库地址，可以是Ip或者域名
		master_port := c.Dc.Master.Port         //数据库端口
		master_Dbname := c.Dc.Master.Db         //数据库名
		master_dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", master_username, master_password, master_host, master_port, master_Dbname)
		var err error
		// 连接Mysql数据库
		_db, err = gorm.Open(mysql.Open(master_dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		//if err != nil {
		//	panic("连接数据库失败, error=" + err.Error())
		//}
		if err != nil {
			zap.L().Error("Use slave err ", zap.Error(err))
			return
		}
		// slave
		replicas := []gorm.Dialector{}
		for _, v := range c.Dc.Slave {
			username := v.Username //账号
			password := v.Password //密码
			host := v.Host         //数据库地址，可以是Ip或者域名
			port := v.Port         //数据库端口
			Dbname := v.Db         //数据库名
			dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", username, password, host, port, Dbname)
			cfg := mysql.Config{
				DSN: dsn,
			}
			replicas = append(replicas, mysql.New(cfg))
		}
		//读写分离
		_db.Use(dbresolver.Register(dbresolver.Config{
			Sources:  []gorm.Dialector{mysql.Open(master_dsn)}, // 或者 mysql.New(mysql.config{DSN:master_dsn})
			Replicas: replicas,
			Policy:   dbresolver.RandomPolicy{}, //策略 随机
		}).SetMaxOpenConns(200).SetMaxIdleConns(10), // 连接池最大连接数200，空闲连接10
		)
		if err != nil {
			zap.L().Error("Use slave err ", zap.Error(err))
			return
		}
	} else { //未启用读写分离
		//配置MySQL连接参数
		username := c.Mc.Username //账号
		password := c.Mc.Password //密码
		host := c.Mc.Host         //数据库地址，可以是Ip或者域名
		port := c.Mc.Port         //数据库端口
		Dbname := c.Mc.Db         //数据库名
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", username, password, host, port, Dbname)
		var err error
		// 连接Mysql数据库
		_db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		//if err != nil {
		//	panic("连接数据库失败, error=" + err.Error())
		//}
		if err != nil {
			zap.L().Error("Use slave err ", zap.Error(err))
			return
		}
	}
	// 设置 config._db == gorms._db 绕过不导出的设定
	gorms.SetDB(_db)
}
