package persistence

import (
	"database/sql"
	"log"

	"github.com/w-k-s/simple-budget-tracker/pkg"
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
		return 0, pkg.NewSystemError(pkg.ErrDatabaseState, "Failed to get unique id", err)
	}
	return uid, err
}
