package main

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	appName           = "nodearmor"
	configFileName    = "settings"
	defaultBackendURL = "https://api.nodearmor.net/"
)

// Config : global configuration store
var Config = viper.New()

// LoadConfig : initializes defaults and loads configuration from file if present
func LoadConfig() {
	// Config settings
	Config.SetConfigType("json")
	Config.SetConfigName(configFileName)

	// Config dirs
	usr, err := user.Current()
	configDir := filepath.Join(usr.HomeDir, fmt.Sprintf(".%s", appName))
	_ = os.Mkdir(configDir, os.ModeDir)
	Config.AddConfigPath(configDir)

	// Set defaults
	Config.SetDefault("BackendURL", defaultBackendURL)

	err = Config.ReadInConfig()
	if err != nil {
		log.Print("configuration file not found. Creating default.")

		os.OpenFile(filepath.Join(configDir, fmt.Sprintf("%s.json", configFileName)), os.O_RDONLY|os.O_CREATE, 0666)
		err = Config.WriteConfig()
		if err != nil {
			log.Printf("error creating default config file: %s", err)
		}
	}
}

// WriteConfig : overwrites config file with current values
func WriteConfig() {
	err := Config.WriteConfig()
	if err != nil {
		log.Printf("error writing config file: %s", err)
		return
	}

	log.Print("config file written")
}
