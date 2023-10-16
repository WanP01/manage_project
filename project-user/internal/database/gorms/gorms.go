package gorms

import (
	"context"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"project-user/config"
)

// 全局DB连接
var _db *gorm.DB

func init() {
	//配置MySQL连接参数
	username := config.AppConf.Mc.Username //账号
	password := config.AppConf.Mc.Password //密码
	host := config.AppConf.Mc.Host         //数据库地址，可以是Ip或者域名
	port := config.AppConf.Mc.Port         //数据库端口
	Dbname := config.AppConf.Mc.Db         //数据库名
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", username, password, host, port, Dbname)
	var err error
	// 连接Mysql数据库
	_db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic("连接数据库失败, error=" + err.Error())
	}
}

func GetDB() *gorm.DB {
	return _db
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
	// g.tx = g.tx.Begin() // error transaction invalid
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
