package config

import (
	"fmt"
	"os"

	"github.com/dcwangmit01/goapi/util"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// globals
var (
	Viper *viper.Viper

	appName    = os.Args[0]
	configName = "config"
	configType = "yaml"
	configDir  = ""
	configPath = ""
)

func init() {
	Viper = viper.New()

	// Explicity set configPath as the config file location
	//   /home/<user>/.goapi/config.yaml
	homeDir, ok := os.LookupEnv("HOME")
	if !ok {
		panic("Unable to locate $HOME directory")
	}
	configDir = fmt.Sprintf("%v/.%v", homeDir, appName)
	configPath = fmt.Sprintf("%v/.%v/%v.%v", homeDir, appName, configName, configType)

	Viper.AddConfigPath(configDir)
	Viper.SetConfigFile(configPath)
	Viper.SetConfigType("yaml") // the configuration is a yaml file

	// Always set defaults, whether or not the config file exists
	Viper.SetDefault("configpath", configPath)
	Viper.SetDefault("appname", appName)
	Viper.SetDefault("host", "localhost")
	Viper.SetDefault("insecure", true) // TODO: Break this out into a CLI flag and change this to false.
	Viper.SetDefault("port", 10080)
	Viper.SetDefault("token", "")

	// Attempt to read the config, if it exists
	err := Viper.ReadInConfig()
	if err == nil {
		// And watch it for changes (does not seem to work on sure if
		// this works on vagrant vbox mount)
		Viper.WatchConfig()
		Viper.OnConfigChange(func(e fsnotify.Event) {
			fmt.Println("Config file changed: ", e.Name)
		})
	} else {
		fmt.Fprintf(os.Stderr, "No configuration file loaded - using defaults\n")
	}
}

func GetHost() string {
	return Viper.GetString("host")
}

func GetInsecure() bool {
	return Viper.GetBool("insecure")
}

func GetPort() int {
	return Viper.GetInt("port")
}

func GetAppName() string {
	return appName
}

// for the time being, viper does not support saving the config back to disk.
//   This is about to arrive in: https://github.com/spf13/viper/pull/287
func SaveConfig() error {
	return util.StructToYamlFile(Viper.AllSettings(), configPath)
}
