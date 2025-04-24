package interfaces

import (
	"github.com/go-pg/pg/v10/orm"
	"io"
)

type DB interface {
	io.Closer
	orm.DB
}
