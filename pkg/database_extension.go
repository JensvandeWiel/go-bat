package pkg

import (
	"github.com/jmoiron/sqlx"
	"reflect"
)

type DatabaseExtension struct {
	db *sqlx.DB
}

func (d *DatabaseExtension) Register(app *Bat) error {
	app.Logger.Info("Registering DatabaseExtension")
	return nil
}

func (d *DatabaseExtension) Requirements() []reflect.Type {
	return []reflect.Type{}
}

func (d *DatabaseExtension) GetDB() *sqlx.DB {
	return d.db
}

func NewDatabaseExtension(db *sqlx.DB) *DatabaseExtension {
	return &DatabaseExtension{db: db}
}
