package persistence

import (
	"database/sql"
	"fmt"
	"log"
)

type UniqueIdDao struct {
	*RootDao
}

func MustOpenDefaultUniqueIdDao(db *sql.DB) *UniqueIdDao {
	return &UniqueIdDao{
		&RootDao{db},
	}
}

func (u UniqueIdDao) GetId(tx *sql.Tx, tableName string, salt string) (uint64, error) {
	var uid uint64
	err := tx.QueryRow("SELECT timestamp_id(?, ?)", tableName, salt).Scan(&uid)
	if err != nil {
		log.Printf("Failed to get unique id for table name %q. Reason; %q", tableName, err)
		return 0, fmt.Errorf("Failed to get unique id. Reason: %w", err)
	}
	return uid, err
}
