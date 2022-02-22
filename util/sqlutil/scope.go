package sqlutil

import "gorm.io/gorm"

// Columns returns a gorm scope that will select the given columns.
func Columns(names ...string) func(*gorm.DB) *gorm.DB {
	if len(names) == 0 {
		return func(db *gorm.DB) *gorm.DB { return db }
	}

	return func(db *gorm.DB) *gorm.DB { return db.Select(names) }
}
