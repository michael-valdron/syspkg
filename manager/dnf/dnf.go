// Package dnf provides an implementation of the syspkg manager interface for the dnf/yum package manager.
// It provides a Go (golang) API interface for interacting with the DNF package manager.
// This package is a wrapper around the dnf command line tool.
//
// Dandified YUM (DNF) is the default package manager on RPM-based systems such as Fedora.
// DNF maintains compatibility with YUM, the previous default package manager for RPM-based systems, and has a strict API for extension and plugins.
//
// For more information about dnf, visit:
// - https://github.com/rpm-software-management/dnf
// - https://yum.baseurl.org
// This package is part of the syspkg library.
package dnf

import "os/exec"

var pm string = "dnf"

// Constants used for dnf commands
const (
	ArgsAssumeYes    string = "-y"
	ArgsAssumeNo     string = "--assumeno"
	ArgsQuiet        string = "-q"
	ArgsNoAutoRemove string = "--no-autoremove"
)

// PackageManager implements the manager.PackageManager interface for the dnf package manager.
type PackageManager struct{}

// IsAvailable checks if the dnf package manager is available on the system.
func (d *PackageManager) IsAvailable() bool {
	_, err := exec.LookPath(pm)
	return err == nil
}
