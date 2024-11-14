package log

// Config for log
type Config struct {
	// Environment defining the log format ("production" or "development").
	// In development mode, the logger:
	// - Enables development settings (e.g., DPanicLevel logs panic)
	// - Uses a console encoder and writes to standard error
	// - Disables sampling
	// - Automatically includes stack traces on WarnLevel and above
	// Check [here](https://pkg.go.dev/go.uber.org/zap@v1.24.0#NewDevelopmentConfig)
	Environment LogEnvironment `mapstructure:"Environment" jsonschema:"enum=production,enum=development"`
	// Level of log. As lower value more logs are going to be generated
	Level string `mapstructure:"Level" jsonschema:"enum=debug,enum=info,enum=warn,enum=error,enum=dpanic,enum=panic,enum=fatal"` //nolint:lll
	// Outputs
	Outputs []string `mapstructure:"Outputs"`
}
