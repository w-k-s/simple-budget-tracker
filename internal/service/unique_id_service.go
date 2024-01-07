package service

import (
	"fmt"
	"log"

	dao "github.com/w-k-s/simple-budget-tracker/internal/persistence"
	svc "github.com/w-k-s/simple-budget-tracker/pkg/services"
)

type defaultUniqueIdService struct {
	uniqueIdDao  *dao.UniqueIdDao
	uniqueIdSalt string
}

func NewUniqueIdService(
	uniqueIdDao *dao.UniqueIdDao,
	uniqueIdSalt string,
) svc.UniqueIdService {

	if uniqueIdDao == nil {
		log.Fatalf("defaultUniqueIdService requires uniqueIdDao")
	}

	if len(uniqueIdSalt) == 0 {
		log.Fatalf("Salt can not be empty string")
	}

	return &defaultUniqueIdService{
		uniqueIdDao:  uniqueIdDao,
		uniqueIdSalt: uniqueIdSalt,
	}
}

func (d defaultUniqueIdService) GetId(entity svc.Entity) (uint64, error) {

	tx, err := d.uniqueIdDao.BeginTx()
	if err != nil {
		return 0, fmt.Errorf("Error getting unique id. Can't create transaction. %w", err)
	}

	uid, err := d.uniqueIdDao.GetId(tx, d.tableNameForEntity(entity), d.uniqueIdSalt)
	if err != nil {
		return 0, fmt.Errorf("Error getting unique id. DB Error. %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return 0, fmt.Errorf("Error getting unique id. Can't commit. %w", err)
	}

	return uid, nil
}

func (d defaultUniqueIdService) MustGetId(entity svc.Entity) uint64 {
	uid, err := d.GetId(svc.Entity(d.tableNameForEntity(entity)))
	if err != nil {
		log.Fatal(err)
	}
	return uid
}

func (d defaultUniqueIdService) tableNameForEntity(entity svc.Entity) string {
	var result string
	switch entity {
	case svc.EntityAccount:
		result = "budget.account"
	case svc.EntityBudget:
		result = "budget.budget"
	case svc.EntityCategory:
		result = "budget.category"
	case svc.EntityRecord:
		result = "budget.record"
	case svc.EntityUser:
		result = "budget.user"
	default:
		log.Fatalf("There is no table name mapping for entity %q. Unique Id can not be created", entity)
	}
	return result
}
