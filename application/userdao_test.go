package application

import (
	"context"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/w-k-s/simple-budget-tracker/core"
	"github.com/w-k-s/simple-budget-tracker/migrations"
)

const (
	POSTGRES_USER = "test"
	POSTGRES_PASSWORD = "test"
	POSTGRES_DB = "simple_budget_tracker"
	driverName   = "postgres"
)

type UserDaoTestSuite struct {
	suite.Suite
	containerCtx context.Context
	postgresC    tc.Container
	userDao      core.UserDao
}

func TestUserDaoTestSuite(t *testing.T) {
	suite.Run(t, new(UserDaoTestSuite))
}

// -- SETUP

func (suite *UserDaoTestSuite) SetupTest() {
	suite.containerCtx = context.Background()
	req := tc.ContainerRequest{
		Image:        "postgres:11.6-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER": POSTGRES_USER,
			"POSTGRES_PASSWORD": POSTGRES_PASSWORD,
			"POSTGRES_DB": POSTGRES_DB,
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections"),
	}
	postgresC, err := tc.GenericContainer(suite.containerCtx, tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		panic(err)
	}

	suite.postgresC = postgresC
	containerHost, _ := postgresC.Host(suite.containerCtx)
	containerPort, _ := postgresC.MappedPort(suite.containerCtx, "5432")
	dataSourceName :=  fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",containerHost, containerPort.Int(), POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB)

	migrations.MustRunMigrations(driverName, dataSourceName, os.Getenv("TEST_MIGRATIONS_DIRECTORY"))
	suite.userDao = MustOpenUserDao(driverName, dataSourceName)
}

// -- TEARDOWN

func (suite *UserDaoTestSuite) TearDownTest() {
	if container := suite.postgresC; container != nil{
		container.Terminate(suite.containerCtx)
	}
}

// -- SUITE

func (suite *UserDaoTestSuite) TestExample() {
	assert.Equal(suite.T(), 5, 5)
}
