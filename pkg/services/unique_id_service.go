package services

type Entity string

const (
	EntityAccount  Entity = "account"
	EntityUser     Entity = "user"
	EntityRecord   Entity = "record"
	EntityCategory Entity = "categoty"
	EntityBudget   Entity = "budget"
)

type UniqueIdService interface {
	GetId(tableName Entity) (uint64, error)
	MustGetId(tableName Entity) uint64
}
