package application

import (
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
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
	err := RunMigrations(testContainerDriverName, testContainerDataSourceName, DefaultMigrationsDirectoryPath())

	// THEN
	assert.Nil(suite.T(), err)
}