package config

type AuditLog struct {
	// `console_output` controls audit log console output.
	ConsoleOutput AuditLogConsole `yaml:"console_output" json:"console_output,omitempty" koanf:"console_output" split_words:"true" jsonschema:"title=console_output"`
	// `mask` determines whether sensitive information (usernames, emails) should be masked in the audit log output.
	//
	// This configuration applies to logs written to the console as well as persisted logs.
	Mask bool `yaml:"mask" json:"mask,omitempty" koanf:"mask" jsonschema:"default=true"`
	// `storage` controls audit log retention.
	Storage AuditLogStorage `yaml:"storage" json:"storage,omitempty" koanf:"storage"`
}

type AuditLogStorage struct {
	// `enabled` controls whether audit log should be retained (i.e. persisted).
	Enabled bool `yaml:"enabled" json:"enabled,omitempty" koanf:"enabled" jsonschema:"default=false"`
}

type AuditLogConsole struct {
	// `enabled` controls whether audit log output on the console is enabled or disabled.
	Enabled bool `yaml:"enabled" json:"enabled,omitempty" koanf:"enabled" jsonschema:"default=true"`
	// `output` determines the output stream audit logs are sent to.
	OutputStream OutputStream `yaml:"output" json:"output,omitempty" koanf:"output" split_words:"true" jsonschema:"default=stdout,enum=stdout,enum=stderr"`
}

type OutputStream string

var (
	OutputStreamStdOut OutputStream = "stdout"
	OutputStreamStdErr OutputStream = "stderr"
)
