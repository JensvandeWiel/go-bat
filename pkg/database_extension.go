package pkg

import "github.com/jmoiron/sqlx"

type DatabaseExtension struct {
	db *sqlx.DB
}

func (d *DatabaseExtension) GetDB() *sqlx.DB {
	return d.db
}

func NewDatabaseExtension(db *sqlx.DB) *DatabaseExtension {
	return &DatabaseExtension{db: db}
}
