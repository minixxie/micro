package models

type Config struct {
	Dbs struct {
		Main struct {
			Uri                   string `yaml:"uri"`
			MaxIdleConns          int    `yaml:"maxIdleConns"`
			MaxOpenConns          int    `yaml:"maxOpenConns"`
			ExpectedSchemaVersion uint   `yaml:"expectedSchemaVersion"`
		} `yaml:"main"`
	} `yaml:"dbs"`
}
