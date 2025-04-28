package runcfg

import "os"

// GetFirstEnv retrieves environment variable values from a list of environment
// variable names. It checks each environment variable in order and returns the
// first non-empty value found. If no value is found, returns an empty string.
func GetFirstEnv(keys ...string) string {
	for _, key := range keys {
		if val := os.Getenv(key); val != "" {
			return val
		}
	}

	return ""
}
