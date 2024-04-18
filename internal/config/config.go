package config

import (
	"bytes"
	"context"
	"encoding/json"
	"html/template"
	"io/ioutil"
	"os"

	api "github.com/gdcorp-domains/fulfillment-go-api"
)

// Config implements JsonConfig and loads configuration values mostly fromm environment variables.
type Config struct {
	api.Config
}

// Load populates the config from the passed location
func (config *Config) Load(ctx context.Context, configLocation string) (err error) {
	var fileContents []byte
	if fileContents, err = ioutil.ReadFile(configLocation); err != nil {
		return err
	}

	envConfig := struct {
		Env    string
		Region string
		EnvDNS string
	}{
		Env:    os.Getenv("ENV"),
		Region: os.Getenv("AWS_REGION"),
		EnvDNS: os.Getenv("ENV"),
	}

	if envConfig.Env == "dev-private" {
		envConfig.EnvDNS = "dp"
	}

	t, err := template.New("").Parse(string(fileContents))
	if err != nil {
		return err
	}

	buf := bytes.Buffer{}
	if err := t.Execute(&buf, envConfig); err != nil {
		return err
	}

	if err := json.Unmarshal(buf.Bytes(), config); err != nil {
		return err
	}

	return nil
}
