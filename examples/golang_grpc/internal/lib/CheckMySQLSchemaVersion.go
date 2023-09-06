package lib

import (
	"database/sql"
)

func CheckMySQLSchemaVersion(db *sql.DB, ver uint) (uint, bool) {
	row := db.QueryRow(`
		SELECT version, dirty FROM schema_migrations LIMIT 1
	`)
	var version int32
	var dirty int32
	err := row.Scan(&version, &dirty)
	if err != nil {
		return 0, false
	}
	if uint(version) == ver && dirty == 0 {
		return uint(version), true
	} else {
		return uint(version), false
	}
}
