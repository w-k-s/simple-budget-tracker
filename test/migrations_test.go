package test

import (
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	db "github.com/w-k-s/simple-budget-tracker/internal/persistence"
)

type MigrationsTestSuite struct {
	suite.Suite
}

func TestMigrationsTestSuite(t *testing.T) {
	suite.Run(t, new(MigrationsTestSuite))
}

// -- SUITE

func (suite *MigrationsTestSuite) Test_WHEN_migrationsAreRunForASecondTime_THEN_noUpdateErrorIsIgnored() {
	// WHEN
	err := db.RunMigrations(TestDB, TestConfig.Database())

	// THEN
	assert.Nil(suite.T(), err)
}
