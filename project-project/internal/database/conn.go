package database

// DbConn 事务相关连接
type DbConn interface {
	Begin()
	Rollback()
	Commit()
}
