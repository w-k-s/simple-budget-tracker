package config

// GptConfig represents the configuration for the GPT server.
type GptConfig struct {
	apiKey string
}

// NewGptConfig creates a new GptConfig with the provided apiKey.
func NewGptConfig(apiKey string) *GptConfig {
	return &GptConfig{
		apiKey: apiKey,
	}
}

func (g GptConfig) ApiKey() string {
	return g.apiKey
}

func (g GptConfig) IsEnabled() bool{
	return len(g.apiKey) > 0
}

// GptConfigBuilder is a builder for GptConfig.
type GptConfigBuilder struct {
	apiKey string
}

// NewGptConfigBuilder creates a new GptConfigBuilder.
func NewGptConfigBuilder() *GptConfigBuilder {
	return &GptConfigBuilder{}
}

// SetApiKey sets the apiKey for the GptConfigBuilder.
func (b *GptConfigBuilder) SetApiKey(apiKey string) *GptConfigBuilder {
	b.apiKey = apiKey
	return b
}

// Build creates a new GptConfig using the current configuration of GptConfigBuilder.
func (b *GptConfigBuilder) Build() *GptConfig {
	return &GptConfig{
		apiKey: b.apiKey,
	}
}
