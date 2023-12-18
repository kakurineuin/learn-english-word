package config

import (
	"os"
)

func EnvMongoDBURI() string {
	return os.Getenv("MONGODB_URI")
}
