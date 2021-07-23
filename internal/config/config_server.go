package config

import (
	"fmt"
	"time"
)

type ServerConfig struct {
	port           int
	readTimeout    time.Duration
	writeTimeout   time.Duration
	maxHeaderBytes int
}

func makeServerConfig(b *serverConfigBuilder) ServerConfig {
	return ServerConfig{
		b.port,
		b.readTimeout,
		b.writeTimeout,
		b.maxHeaderBytes,
	}
}

func (s ServerConfig) Port() int {
	return s.port
}

func (s ServerConfig) MaxHeaderBytes() int {
	if s.maxHeaderBytes <= 0 {
		return 1 << 20 // 1MB
	}
	return s.maxHeaderBytes
}

func (s ServerConfig) ReadTimeout() time.Duration {
	if s.readTimeout == 0 {
		return 10 * time.Second
	}
	return s.readTimeout
}

func (s ServerConfig) WriteTimeout() time.Duration {
	if s.writeTimeout == 0 {
		return 10 * time.Second
	}
	return s.writeTimeout
}

func (s ServerConfig) ListenAddress() string {
	return fmt.Sprintf(":%d", s.port)
}

type serverConfigBuilder struct {
	port           int
	readTimeout    time.Duration
	writeTimeout   time.Duration
	maxHeaderBytes int
}

func NewServerConfigBuilder() *serverConfigBuilder {
	return &serverConfigBuilder{
		port:         8080,
		readTimeout:  time.Duration(0),
		writeTimeout: time.Duration(0),
	}
}

func (b *serverConfigBuilder) SetPort(port int) *serverConfigBuilder {
	b.port = port
	return b
}

func (b *serverConfigBuilder) SetReadTimeout(timeout time.Duration) *serverConfigBuilder {
	b.readTimeout = timeout
	return b
}

func (b *serverConfigBuilder) SetWriteTimeout(timeout time.Duration) *serverConfigBuilder {
	b.readTimeout = timeout
	return b
}

func (b *serverConfigBuilder) SetMaxHeaderBytes(maxHeaderBytes int) *serverConfigBuilder {
	b.maxHeaderBytes = maxHeaderBytes
	return b
}

func (b *serverConfigBuilder) Build() ServerConfig {
	return makeServerConfig(b)
}
