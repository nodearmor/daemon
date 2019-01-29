package nodearmord

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	appName              = "nodearmor"
	configFileName       = "settings"
	defaultControllerURL = "wss://api.nodearmor.net/"
)

// Config : global configuration store
var config = viper.New()

// LoadConfig : initializes defaults and loads configuration from file if present
func LoadConfig() {
	// Config settings
	config.SetConfigType("json")
	config.SetConfigName(configFileName)

	// Config dirs
	usr, err := user.Current()
	configDir := filepath.Join(usr.HomeDir, fmt.Sprintf(".%s", appName))
	_ = os.Mkdir(configDir, os.ModeDir)
	config.AddConfigPath(configDir)

	// Set defaults
	config.SetDefault("ControllerURL", defaultControllerURL)

	err = config.ReadInConfig()
	if err != nil {
		log.Print("configuration file not found. Creating default.")

		os.OpenFile(filepath.Join(configDir, fmt.Sprintf("%s.json", configFileName)), os.O_RDONLY|os.O_CREATE, 0666)
		err = config.WriteConfig()
		if err != nil {
			log.Printf("error creating default config file: %s", err)
		}
	}
}

// WriteConfig : overwrites config file with current values
func WriteConfig() {
	err := config.WriteConfig()
	if err != nil {
		log.Printf("error writing config file: %s", err)
		return
	}

	log.Print("config file written")
}
