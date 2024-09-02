package config

type LoggerConfig struct {
	// `log_health_and_metrics` determines whether requests of the `/health` and `/metrics` endpoints are logged.
	LogHealthAndMetrics bool `yaml:"log_health_and_metrics,omitempty" json:"log_health_and_metrics" koanf:"log_health_and_metrics" jsonschema:"default=true"`
}
