package misc

import (
	"context"
	"fmt"
	"os/exec"

	"smtp2communicator/pkg/logger"
	"smtp2communicator/pkg/service"
)

// SystemdService handles systemd un/install flags
//
// This function is taking care of un/installing this tool as a Systemd service.
//
// Parameters:
//
// - ctx (context.Context): context,
// - systemdInstallFlag (*bool): flag if to install as Systemd service,
// - systemdUninstallFlag (*bool): flag is to uninstall Systemd serves,
// - configurationFileFlag (*string): flag pointing to configuration file,
// - configurationInstallPath (string): configuration install path,
// - thisBinaryPath (string): path to this tool,
// - binInstallPath (string): binary install path,
// - projectName (string): name of project, used name service,
//
// Returns:
//
// - exit (bool): always returns true indicating we need to exit afterwards
func SystemdService(
	ctx context.Context,
	systemdInstallFlag *bool,
	systemdUninstallFlag *bool,
	configurationFileFlag *string,
	configurationInstallPath string,
	thisBinaryPath string,
	binInstallPath string,
	projectName string,
) (exit bool) {
	// get logger
	log := logger.LoggerFromContext(ctx)

	// check if systemd installed by looking up systemd executable
	if systemdPath, err := exec.LookPath("systemd"); err != nil {
		log.Info("It doesn't look like Systemd is present on this system. Exiting.")
		return true
	} else {
		log.Debugf("Systemd executable found here: %s", systemdPath)
	}

	// Systemd
	if !*systemdInstallFlag || !*systemdUninstallFlag {
		s := service.New(ctx)
		configurationFileName := fmt.Sprintf("%s%s.yaml", configurationInstallPath, projectName)
		executableInstallationPath := fmt.Sprintf("%s%s", binInstallPath, projectName)
		s.Name = projectName
		s.Description = "This is my smtp stub that forwards all emails to a communicator"
		s.ExecStart = fmt.Sprintf("%s -configuration %s -verbosity debug", executableInstallationPath, configurationFileName)
		s.AddFileToCopy(*configurationFileFlag, configurationFileName, 0o600)
		s.AddFileToCopy(thisBinaryPath, executableInstallationPath, 0o755)

		if *systemdInstallFlag {
			s.InstallEnableStart()
		}
		if *systemdUninstallFlag {
			s.StopDisableUninstall()
		}
	} else {
		log.Error("The 'systemdInstall' and 'systemdUninstall' flags can't be used together!")
	}
	return true
}
