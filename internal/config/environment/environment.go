package environment

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v9"
	"github.com/joho/godotenv"
)

// EnvConfig is a struct that holds all the environment variables
type EnvConfig struct {
	MinioEndpoint      string        `env:"MINIO_ENDPOINT"       envDefault:"localhost:9000"`
	MinioAccessKey     string        `env:"MINIO_ACCESS_KEY"`
	MinioAccessSecret  string        `env:"MINIO_ACCESS_SECRET"`
	MinioUseSSL        bool          `env:"MINIO_USE_SSL"        envDefault:"false"`
	MinioBucketName    string        `env:"MINIO_BUCKET_NAME"    envDefault:"minio-link"`
	MinioRegion        string        `env:"MINIO_REGION"         envDefault:"us-east-1"`
	MinioObjectLocking bool          `env:"MINIO_OBJECT_LOCKING" envDefault:"false"`
	MinioDefaultExpiry time.Duration `env:"MINIO_DEFAULT_EXPIRY" envDefault:"168h"`
	YourlsEndpoint     string        `env:"YOURLS_ENDPOINT"      envDefault:"http://localhost:8080"`
	YourlsSignatureKey string        `env:"YOURLS_SIGNATURE_KEY"`
}

// Enables printing of config without sensitive data
func (e *EnvConfig) String() string {
	return fmt.Sprintf(
		"minio endpoint: %s, minio use ssl: %t, minio bucket name: %s, minio region: %s, minio object locking: %t, yourls url: %s",
		e.MinioEndpoint,
		e.MinioUseSSL,
		e.MinioBucketName,
		e.MinioRegion,
		e.MinioObjectLocking,
		e.YourlsEndpoint,
	)
}

// Load loads the environment variables from environment
// or given files if specified
func Load(envFiles ...string) (*EnvConfig, error) {
	for _, envFile := range envFiles {
		if envFile != "" {
			if err := godotenv.Load(envFile); err != nil {
				return nil, fmt.Errorf("failed to load environment file: %s", err)
			}
		}
	}
	var cfg EnvConfig
	if err := env.ParseWithOptions(&cfg, options); err != nil {
		return nil, fmt.Errorf("failed to parse environment: %s", err)
	}
	return &cfg, nil
}

var (
	options = env.Options{
		Prefix:          "LINK_",
		RequiredIfNoDef: true,
	}
)
