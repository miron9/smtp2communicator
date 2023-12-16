package common

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"smtp2communicator/pkg/logger"
	"smtp2communicator/pkg/utils"

	"gopkg.in/yaml.v3"
)

type FileChannel struct {
	Enabled bool
	DirPath string `yaml:"dirPath"`
}
type TelegramChannel struct {
	Enabled bool
	UserId  int64  `yaml:"userId"`
	BotKey  string `yaml:"botKey"`
}
type SlackChannel struct {
	Enabled bool
	UserId  string `yaml:"userId"`
	BotKey  string `yaml:"botKey"`
}
type TeamsChannel struct {
	Enabled bool
}
type WhatsupChannel struct {
	Enabled bool
}
type Channels struct {
	File     FileChannel
	Telegram TelegramChannel
	Slack    SlackChannel
	Teams    TeamsChannel
	Whatsup  WhatsupChannel
}
type Configuration struct {
	Port     int `yaml:"tcpPort"`
	Channels Channels
}

// GetConfiguration loads and returns configuration object
//
// This function loads configuration from specified location.
//
// Parameters:
//
// - ctx (context.Context): context
// - configurationPath (string): path to configuration yaml
//
// Returns:
// - err (error): error if any or nil
func (c *Configuration) GetConfiguration(ctx context.Context, configurationPath string) (err error) {
	log := logger.LoggerFromContext(ctx)

	if len(configurationPath) == 0 {
		return errors.New("No configuration file specified")
	}

	log.Debugf("loading configuration: %s", configurationPath)

	fileHandle, err := os.ReadFile(configurationPath)
	if err != nil {
		return
	}

	// configurationData := Configuration{}
	err = yaml.Unmarshal(fileHandle, c)
	if err != nil {
		log.Debugf("can't unmarshall configuration: %s", err)
		return
	}

	return
}

// FindConfigurationFile returns path to configuration file
//
// This function will check a list of location for configuration file.
// The locations are: current working dir, location of the binary,
// home dir of the user executinh this and finally /etc.
//
// Parameters:
//
// - ctx (context.Context): context
// - myPath (string): path to this executable
// - configurationFileName (string): name of the configuration file to load
//
// Returns:
// - location (string): path to first existing configuration file
// - err (error): error if any or nil
func FindConfigurationFile(ctx context.Context, myPath string, configurationFileName string) (location string, err error) {
	log := logger.LoggerFromContext(ctx)

	locations := []string{}

	cwd, err := os.Getwd()
	if err != nil {
		log.Errorf("Can't get current working directory: %v", err)
		return
	}

	// order of configuration locations sto search, starting with closest to
	// user, then executable itself, home dir and finally /etc
	locations = append(locations, cwd)
	locations = append(locations, filepath.Dir(myPath))
	locations = append(locations, "~")
	locations = append(locations, "/etc/")

	var failedLocations []string

	for _, location = range locations {
		// log.Debugf("trying configuration file at: %s", location)
		location, err = utils.ExpandTilde(location)
		if err != nil {
			log.Error(err)
			return
		}

		location = filepath.Join(location, configurationFileName)

		_, err = os.Lstat(location)
		if err == nil {
			return
		}
		failedLocations = append(failedLocations, location)
	}

	return "", errors.New(fmt.Sprintf("%v", failedLocations))
}
