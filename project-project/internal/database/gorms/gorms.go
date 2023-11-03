package gorms

import (
	"context"
	"gorm.io/gorm"
)

// 全局DB连接
var _db *gorm.DB

//改为config 配置时 连接mysql （config）
//func init() {
//	if config.AppConf.Dc.Separation { // 开启读写分离
//		//master
//		master_username := config.AppConf.Dc.Master.Username //账号
//		master_password := config.AppConf.Dc.Master.Password //密码
//		master_host := config.AppConf.Dc.Master.Host         //数据库地址，可以是Ip或者域名
//		master_port := config.AppConf.Dc.Master.Port         //数据库端口
//		master_Dbname := config.AppConf.Dc.Master.Db         //数据库名
//		master_dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", master_username, master_password, master_host, master_port, master_Dbname)
//		var err error
//		// 连接Mysql数据库
//		_db, err = gorm.Open(mysql.Open(master_dsn), &gorm.Config{
//			Logger: logger.Default.LogMode(logger.Info),
//		})
//		if err != nil {
//			panic("连接数据库失败, error=" + err.Error())
//		}
//		// slave
//		replicas := []gorm.Dialector{}
//		for _, v := range config.AppConf.Dc.Slave {
//			username := v.Username //账号
//			password := v.Password //密码
//			host := v.Host         //数据库地址，可以是Ip或者域名
//			port := v.Port         //数据库端口
//			Dbname := v.Db         //数据库名
//			dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", username, password, host, port, Dbname)
//			cfg := mysql.Config{
//				DSN: dsn,
//			}
//			replicas = append(replicas, mysql.New(cfg))
//		}
//		//读写分离
//		_db.Use(dbresolver.Register(dbresolver.Config{
//			Sources:  []gorm.Dialector{mysql.Open(master_dsn)}, // 或者 mysql.New(mysql.config{DSN:master_dsn})
//			Replicas: replicas,
//			Policy:   dbresolver.RandomPolicy{}, //策略 随机
//		}).SetMaxOpenConns(200).SetMaxIdleConns(10), // 连接池最大连接数200，空闲连接10
//		)
//	} else { //未启用读写分离
//		//配置MySQL连接参数
//		username := config.AppConf.Mc.Username //账号
//		password := config.AppConf.Mc.Password //密码
//		host := config.AppConf.Mc.Host         //数据库地址，可以是Ip或者域名
//		port := config.AppConf.Mc.Port         //数据库端口
//		Dbname := config.AppConf.Mc.Db         //数据库名
//		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", username, password, host, port, Dbname)
//		var err error
//		// 连接Mysql数据库
//		_db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
//			Logger: logger.Default.LogMode(logger.Info),
//		})
//		if err != nil {
//			panic("连接数据库失败, error=" + err.Error())
//		}
//	}
//}

func GetDB() *gorm.DB {
	return _db
}

func SetDB(db *gorm.DB) {
	_db = db
}

type GormConn struct {
	db *gorm.DB // 数据库连接 _DB
	tx *gorm.DB // 事务专用连接
}

func New() *GormConn {
	return &GormConn{db: GetDB(), tx: GetDB()}
}

func NewTran() *GormConn {
	return &GormConn{db: GetDB(), tx: GetDB()}
}

func (g *GormConn) Session(ctx context.Context) *gorm.DB {
	return g.db.Session(&gorm.Session{Context: ctx})
}

func (g *GormConn) Begin() {
	// 这一步很重要，事务的连接需要每次都新建，不能直接连用
	//g.tx = g.tx.Begin() // error invalid transaction
	g.tx = GetDB().Begin()
}

func (g *GormConn) Rollback() {
	g.tx.Rollback()
}
func (g *GormConn) Commit() {
	g.tx.Commit()
}

func (g *GormConn) Tx(ctx context.Context) *gorm.DB {
	return g.tx.WithContext(ctx)
}
