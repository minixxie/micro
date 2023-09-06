package models

import (
	"database/sql"
)

type FirstModel interface {
	CreateRecord(name string) (uint32, error)
}

type FirstModelImpl struct {
	MainDB *sql.DB
}

var _ FirstModel = (*FirstModelImpl)(nil) // make sure it implements the interface
