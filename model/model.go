package model

import "github.com/jinzhu/gorm"

type Model struct {
	Db *gorm.DB
}

func New(Db *gorm.DB) *Model {
	return &Model{Db}
}

// Close closes the databases associated to this model
func (m *Model) Close() {
	if m.Db != nil {
		m.Db.Close()
	}
}
