package services

import (
	"log"
	"time"
)

type UniqueIdService interface {
	GetId() (uint64, error)
	MustGetId() uint64
}

type uniqueIdService struct {
	
}

func NewUniqueIdService() (UniqueIdService, error) {
	return &uniqueIdService{}, nil
}

func (u uniqueIdService) GetId() (uint64, error) {
	// TODO: Look into something like snowflake or whatever.
	return uint64(time.Now().UTC().UnixNano()),nil
}

func (u uniqueIdService) MustGetId() uint64{
	id, err := u.GetId()
	if err != nil{
		log.Fatalf("Failed to get unique id. Reason: %s", err)
	}
	return id
}