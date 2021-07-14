package config

import "os"

const GH_TOKEN = "GH_TOKEN"

// EnvironmentAuthentication returns the the value for the enviromenment variable `GH_TOKEN`
func EnvironmentAuthentication() string {
	if token := os.Getenv(GH_TOKEN); token != "" {
		return token
	}

	return os.Getenv(GH_TOKEN)
}
