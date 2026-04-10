package ophim1

import "time"

type Config struct {
	Enabled  bool          `env:"OPHIM1_ENABLED" envDefault:"true"`
	BaseURL  string        `env:"OPHIM1_BASE_URL" envDefault:"https://ophim1.com"`
	APIKey   string        `env:"OPHIM1_API_KEY"`	
	Timeout  time.Duration `env:"OPHIM1_TIMEOUT" envDefault:"15s"`
	Priority int           `env:"OPHIM1_PRIORITY" envDefault:"5"`
}

func DefaultConfig() Config {
	return Config{Enabled: true, BaseURL: "https://ophim1.com", Timeout: 15 * time.Second, Priority: 5}
}
