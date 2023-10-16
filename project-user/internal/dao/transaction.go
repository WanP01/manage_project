package dao

import (
	"project-user/internal/database"
	"project-user/internal/database/gorms"
)

type TransactinonDao struct {
	conn database.DbConn
}

func (td *TransactinonDao) Action(f func(conn database.DbConn) error) error {
	td.conn.Begin()
	err := f(td.conn)
	if err != nil {
		td.conn.Rollback()
		return err
	}
	td.conn.Commit()
	return nil
}
func NewTransactionDao() *TransactinonDao {
	return &TransactinonDao{
		conn: gorms.NewTran(),
	}
}
