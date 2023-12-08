package file

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"time"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	i "smtp2communicator/internal/common"
)

// createDirectory create a directory if it doesn't exist
//
// This function checks if a directory exists and creates it if not.

// Parameters:
//
// - path (string): path to directory to be created
//
// Returns:
// - err (error): error if any or nil
func createDirectory(log *zap.SugaredLogger, path string) error {
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		log.Debugf("directory to store mails doesn't exist: %s", path)
		if err := os.Mkdir(path, 0o744); err != nil {
			log.Errorf("can't create directory '%s': %v", path, err)
			return err
		}
	}
	return nil
}

// saveEmailToFile saves email to YAML file
//
// This function saves each email to separate file in YAML format.
// This is fulfiling the File channel configuration
//
// Parameters:
//
// - conf (FileChannel): configuration specific to the File channel
// - msg (message): struct representing received email
//
// Returns:
// - err (error): error if any or nil
func SaveEmailToFile(log *zap.SugaredLogger, conf i.FileChannel, msg i.Message) error {
	if !conf.Enabled {
		log.Debug("File channel disabled, returning")
		return nil
	}

	createDirectory(log, conf.DirPath)

	// Create a unique filename based on the current timestamp
	filename := fmt.Sprintf("%s/received_email_%d.yaml", conf.DirPath, time.Now().Unix())

	msgMarshalled, err := yaml.Marshal(msg)
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile(filename, msgMarshalled, 0o644); err != nil {
		log.Errorf("Can't save to file: %v", err)
		return err
	}
	log.Infof("Email saved to %s\n", filename)

	return nil
}
