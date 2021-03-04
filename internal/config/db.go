package config

// DBConfig provides configuration options for the database connection
type DBConfig struct {
	Type            string `yaml:"type"`
	Address         string `yaml:"address"`
	Port            int    `yaml:"port"`
	User            string `yaml:"user"`
	Password        string `yaml:"password"`
	Database        string `yaml:"database"`
	ApplicationName string `yaml:"applicationName"`
}
