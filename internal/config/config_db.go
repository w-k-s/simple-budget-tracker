package config

import (
	"fmt"
)

type DBConfig struct {
	username     string
	password     string
	host         string
	port         int
	name         string
	sslMode      string
	migrationDir string
}

func (d DBConfig) Username() string {
	return d.username
}

func (d DBConfig) Password() string {
	return d.password
}

func (d DBConfig) Host() string {
	return d.host
}

func (d DBConfig) Port() int {
	return d.port
}

func (d DBConfig) Name() string {
	return d.name
}

func (d DBConfig) Schema() string {
	return "budget"
}

func (d DBConfig) SslMode() string {
	return d.sslMode
}

func (d DBConfig) DriverName() string {
	return "postgres"
}

func (d DBConfig) MigrationDirectory() string {
	if len(d.migrationDir) == 0 {
		return DefaultMigrationsDirectoryPath()
	}
	return d.migrationDir
}

func (d DBConfig) ConnectionString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host(),
		d.Port(),
		d.Username(),
		d.Password(),
		d.Name(),
		d.SslMode(),
	)
}

type dbConfigBuilder struct {
	username     string
	password     string
	host         string
	port         int
	name         string
	sslMode      string
	migrationDir string
}

func NewDBConfigBuilder() *dbConfigBuilder {
	return &dbConfigBuilder{
		sslMode:      "disable",
		migrationDir: DefaultMigrationsDirectoryPath(),
	}
}

func (b *dbConfigBuilder) SetUsername(username string) *dbConfigBuilder {
	b.username = username
	return b
}

func (b *dbConfigBuilder) SetPassword(password string) *dbConfigBuilder {
	b.password = password
	return b
}

func (b *dbConfigBuilder) SetHost(host string) *dbConfigBuilder {
	b.host = host
	return b
}

func (b *dbConfigBuilder) SetPort(port int) *dbConfigBuilder {
	b.port = port
	return b
}

func (b *dbConfigBuilder) SetName(name string) *dbConfigBuilder {
	b.name = name
	return b
}

func (b *dbConfigBuilder) SetSSLMode(sslMode string) *dbConfigBuilder {
	b.sslMode = sslMode
	return b
}

func (b *dbConfigBuilder) SetMigrationDirectory(migrationDir string) *dbConfigBuilder {
	b.migrationDir = migrationDir
	return b
}

func (b *dbConfigBuilder) Build() DBConfig {
	return DBConfig{
		b.username,
		b.password,
		b.host,
		b.port,
		b.name,
		b.sslMode,
		b.migrationDir,
	}
}