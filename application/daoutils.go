package application

import (
	"github.com/lib/pq"
)

func isDuplicateKeyError(err error) (string, bool) {
	if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
		return pqErr.Detail, true
	}
	return "", false
}
