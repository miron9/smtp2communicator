package misc

import (
	"context"
	"fmt"
	"os"
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
// - n/a
func SystemdService(
	ctx context.Context,
	systemdInstallFlag *bool,
	systemdUninstallFlag *bool,
	configurationFileFlag *string,
	configurationInstallPath string,
	thisBinaryPath string,
	binInstallPath string,
	projectName string,
) {
	//get logger
	log := logger.LoggerFromContext(ctx)

	// Systemd
	if !*systemdInstallFlag || !*systemdUninstallFlag {
		s := service.New(ctx)
		configurationFileName := fmt.Sprintf("%s%s.yaml", configurationInstallPath, projectName)
		executableInstallationPath := fmt.Sprintf("%s%s", binInstallPath, projectName)
		s.Name = projectName
		s.Description = "This is my smtp stub that forwards all emails to Telegram"
		s.ExecStart = fmt.Sprintf("%s -configuration %s -verbosity debug", executableInstallationPath, configurationFileName)
		s.AddFileToCopy(*configurationFileFlag, configurationFileName, 0o600)
		s.AddFileToCopy(thisBinaryPath, executableInstallationPath, 0o755)

		if *systemdInstallFlag {
			s.InstallEnableStart()
			os.Exit(0)
		}
		if *systemdUninstallFlag {
			s.StopDisableUninstall()
			os.Exit(0)
		}
	} else {
		log.Error("The 'systemdInstall' and 'systemdUninstall' flags can't be used together!")
		os.Exit(1)
	}
}
