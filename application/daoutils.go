package application

import (
	"context"
	"fmt"

	"github.com/lib/pq"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	testContainerPostgresUser     = "test"
	testContainerPostgresPassword = "test"
	testContainerPostgresDB       = "simple_budget_tracker"
	TestContainerDriverName       = "postgres"
)

func isDuplicateKeyError(err error) (string, bool) {
	if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
		return pqErr.Detail, true
	}
	return "", false
}

func requestPostgresTestContainer() (*context.Context, tc.Container, string, error) {
	containerCtx := context.Background()
	req := tc.ContainerRequest{
		Image:        "postgres:11.6-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     testContainerPostgresUser,
			"POSTGRES_PASSWORD": testContainerPostgresPassword,
			"POSTGRES_DB":       testContainerPostgresDB,
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections"),
	}
	postgresC, err := tc.GenericContainer(containerCtx, tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, nil, "", err
	}

	containerHost, _ := postgresC.Host(containerCtx)
	containerPort, _ := postgresC.MappedPort(containerCtx, "5432")
	dataSourceName := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", containerHost, containerPort.Int(), testContainerPostgresUser, testContainerPostgresPassword, testContainerPostgresDB)

	return &containerCtx, postgresC, dataSourceName, nil
}
