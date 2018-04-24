package main

import (
	"github.com/golang/glog"
	"github.com/spf13/viper"
)

// Config go-hone configuration structure
type Config struct {
	APIEnabled  bool
	APISecret   string
	APIKey      string
	APIEndpoint string
	APIVerify   bool
}

// NewConfig Creates a new configuration struct, return a *Config
func NewConfig() *Config {
	c := new(Config)

	// Configuration file management
	// name of config file (without extension)
	viper.SetConfigName("config")
	viper.AddConfigPath("conf/")
	viper.SetConfigType("toml")
	// Find and read the config file
	err := viper.ReadInConfig()
	if err != nil {
		// Handle errors reading the config file
		glog.Fatalf("Fatal error config file: %s", err.Error())
	}

	c.loadConfig()

	return c
}

func (c *Config) loadConfig() {
	//TODO need some error verification here
	glog.V(2).Infof("Loading OpenAPI Config...")
	o := viper.Sub("openapi")
	c.APISecret = o.GetString("secret")
	c.APIKey = o.GetString("key")
	c.APIEndpoint = o.GetString("endpoint")
	c.APIVerify = o.GetBool("verify")
	c.APIEnabled = o.GetBool("enabled")
}
