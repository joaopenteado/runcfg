package runcfg

// LoadOption is a function that configures the loading process.
type LoadOption func(*loadOptions)

// loadOptions configures how we load the configuration.
type loadOptions struct {
	requiredMetadata MetadataField
	defaultPort      uint16
}

func defaultOptions() loadOptions {
	return loadOptions{
		requiredMetadata: MetadataAll,
		defaultPort:      8080,
	}
}

// WithMetadata specifies which metadata fields to load.
// By default, all metadata fields are loaded.
func WithMetadata(fields MetadataField) LoadOption {
	return func(o *loadOptions) {
		o.requiredMetadata = fields
	}
}

// WithDefaultPort specifies the default port to use if the PORT environment
// variable is not set. By default, Port will be set to 8080 if the environment
// variable is not set.
func WithDefaultPort(port uint16) LoadOption {
	return func(o *loadOptions) {
		o.defaultPort = port
	}
}
