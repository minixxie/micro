package lib

import (
	"bufio"
	"gopkg.in/yaml.v2"
	"os"
)

func LoadConfig(config interface{}) error {
	var err error

	var env string
	env = os.Getenv("ENV")
	if env = os.Getenv("ENV"); env == "" {
		env = "local"
	}

	file, err := os.Open("/config/" + env + ".yml")
	if err != nil {
		return err
	}
	defer file.Close()

	stats, statsErr := file.Stat()
	if statsErr != nil {
		return statsErr
	}

	var size int64 = stats.Size()
	bytes := make([]byte, size)

	bufr := bufio.NewReader(file)
	_, err = bufr.Read(bytes)

	err = yaml.Unmarshal(bytes, config)
	if err != nil {
		return err
	}
	return nil
}
