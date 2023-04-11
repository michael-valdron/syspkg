package apt

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"

	"github.com/bluet/syspkg/internal"
)

func ParseInstallOutput(output string, opts *internal.Options) []internal.PackageInfo {
	// remove the last empty line
	output = strings.TrimSuffix(output, "\n")
	lines := strings.Split(string(output), "\n")
	// lines := strings.Split(output, "\n")
	var packages []internal.PackageInfo

	for _, line := range lines {
		if opts.Verbose {
			log.Printf("apt: %s", line)
		}
		if strings.HasPrefix(line, "Setting up") {
			parts := strings.Fields(line)
			var name, arch string
			if strings.Contains(parts[2], ":") {
				name = strings.Split(parts[2], ":")[0]
				arch = strings.Split(parts[2], ":")[1]
			} else {
				name = parts[2]
			}

			// if name is empty, it might be not what we want
			if name == "" {
				continue
			}

			packageInfo := internal.PackageInfo{
				Name:           name,
				Arch:           arch,
				Version:        strings.Trim(parts[3], "()"),
				NewVersion:     strings.Trim(parts[3], "()"),
				Status:         internal.PackageStatusInstalled,
				PackageManager: pm,
			}
			packages = append(packages, packageInfo)
		}
	}

	return packages
}

func ParseDeletedOutput(output string, opts *internal.Options) []internal.PackageInfo {
	var packages []internal.PackageInfo
	// remove the last empty line
	output = strings.TrimSuffix(output, "\n")
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if opts.Verbose {
			log.Printf("apt: %s", line)
		}
		if strings.HasPrefix(line, "Removing") {
			parts := strings.Fields(line)
			if opts.Verbose {
				log.Printf("apt: parts: %s", parts)
			}
			var name, arch string
			if strings.Contains(parts[1], ":") {
				name = strings.Split(parts[1], ":")[0]
				arch = strings.Split(parts[1], ":")[1]
			} else {
				name = parts[1]
			}

			// if name is empty, it might be not what we want
			if name == "" {
				continue
			}

			packageInfo := internal.PackageInfo{
				Name:           name,
				Version:        strings.Trim(parts[2], "()"),
				NewVersion:     "",
				Category:       "",
				Arch:           arch,
				Status:         internal.PackageStatusAvailable,
				PackageManager: pm,
			}
			packages = append(packages, packageInfo)
		}
	}

	return packages
}

func ParseFindOutput(output string, opts *internal.Options) []internal.PackageInfo {
	// Sorting...
	// Full Text Search...
	// zutty/jammy 0.11.2.20220109.192032+dfsg1-1 amd64
	//   Efficient full-featured X11 terminal emulator
	//
	// zvbi/jammy 0.2.35-19 amd64
	//   Vertical Blanking Interval (VBI) utilities

	output = strings.TrimPrefix(output, "Sorting...\nFull Text Search...\n")

	// remove the last empty line
	output = strings.TrimSuffix(output, "\n")

	// split output by empty lines
	lines := strings.Split(output, "\n\n")

	var packages []internal.PackageInfo
	var packagesDict = make(map[string]internal.PackageInfo)

	for _, line := range lines {
		if regexp.MustCompile(`^[\w\d-]+/[\w\d-,]+`).MatchString(line) {
			parts := strings.Fields(line)

			// debug a package with version "1.4p5-50build1"
			if strings.Contains(parts[1], "1.4p5-50build1") || strings.Contains(parts[1], "1.2.6-1") {
				fmt.Printf("apt: debug line: %s\n", line)
			}

			// if name is empty, it might be not what we want
			if parts[0] == "" {
				continue
			}

			packageInfo := internal.PackageInfo{
				Name:           strings.Split(parts[0], "/")[0],
				Version:        parts[1],
				NewVersion:     parts[1],
				Category:       strings.Split(parts[0], "/")[1],
				Arch:           parts[2],
				PackageManager: pm,
			}

			packagesDict[packageInfo.Name] = packageInfo
		}
	}

	if len(packagesDict) == 0 {
		return packages
	}

	packages, err := getPackageStatus(&packagesDict)
	if err != nil {
		log.Printf("apt: getPackageStatus error: %s\n", err)
	}

	return packages
}

func ParseListInstalledOutput(output string, opts *internal.Options) []internal.PackageInfo {
	var packages []internal.PackageInfo
	// remove the last empty line
	output = strings.TrimSuffix(output, "\n")
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if len(line) > 0 {
			parts := strings.Fields(line)

			// if name is empty, it might be not what we want
			if parts[0] == "" {
				continue
			}
			var name, arch string
			if strings.Contains(parts[0], ":") {
				name = strings.Split(parts[0], ":")[0]
				arch = strings.Split(parts[0], ":")[1]
			} else {
				name = parts[0]
			}

			packageInfo := internal.PackageInfo{
				Name:           name,
				Version:        parts[1],
				Status:         internal.PackageStatusInstalled,
				Arch:           arch,
				PackageManager: pm,
			}
			packages = append(packages, packageInfo)
		}
	}

	return packages
}

func ParseListUpgradableOutput(output string, opts *internal.Options) []internal.PackageInfo {
	// Listing...
	// cloudflared/unknown 2023.4.0 amd64 [upgradable from: 2023.3.1]
	// libllvm15/jammy-updates 1:15.0.7-0ubuntu0.22.04.1 amd64 [upgradable from: 1:15.0.6-3~ubuntu0.22.04.2]
	// libllvm15/jammy-updates 1:15.0.7-0ubuntu0.22.04.1 i386 [upgradable from: 1:15.0.6-3~ubuntu0.22.04.2]

	var packages []internal.PackageInfo
	// remove the last empty line
	output = strings.TrimSuffix(output, "\n")
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		// skip if line starts with "Listing..."
		if strings.HasPrefix(line, "Listing...") {
			continue
		}

		if len(line) > 0 {
			parts := strings.Fields(line)
			// log.Printf("apt: parts: %+v", parts)

			name := strings.Split(parts[0], "/")[0]
			category := strings.Split(parts[0], "/")[1]
			newVersion := parts[1]
			arch := parts[2]
			version := parts[5]
			version = strings.TrimSuffix(version, "]")

			packageInfo := internal.PackageInfo{
				Name:           name,
				Version:        version,
				NewVersion:     newVersion,
				Category:       category,
				Arch:           arch,
				Status:         internal.PackageStatusUpgradable,
				PackageManager: pm,
			}
			packages = append(packages, packageInfo)
		}
	}

	return packages
}

func getPackageStatus(packages *map[string]internal.PackageInfo) ([]internal.PackageInfo, error) {
	var packageNames []string
	var packagesList []internal.PackageInfo

	if len(*packages) == 0 {
		return packagesList, nil
	}

	for name := range *packages {
		packageNames = append(packageNames, name)
	}

	args := []string{"-W", "--showformat", "${binary:Package} ${Status} ${Version}\n"}
	args = append(args, packageNames...)
	cmd := exec.Command("dpkg-query", args...)
	cmd.Env = ENV_NonInteractive

	// dpkg-query might exit with status 1, which is not an error when some packages are not found
	out, err := cmd.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() != 1 && !strings.Contains(string(out), "no packages found matching") {
				return nil, fmt.Errorf("command failed with output: %s", string(out))
			}
		}
	}

	packagesList, err = ParseDpkgQueryOutput(string(out), packages)
	if err != nil {
		return nil, fmt.Errorf("failed to parse dpkg-query output: %+v", err)
	}

	// for all the packages that are not found, set their status to unknown, if any
	for _, pkg := range *packages {
		fmt.Printf("apt: package not found by dpkg-query: %s", pkg.Name)
		pkg.Status = internal.PackageStatusUnknown
		packagesList = append(packagesList, pkg)
	}

	return packagesList, nil
}

func ParseDpkgQueryOutput(output string, packages *map[string]internal.PackageInfo) ([]internal.PackageInfo, error) {
	var packagesList []internal.PackageInfo

	// remove the last empty line
	output = strings.TrimSuffix(output, "\n")
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if len(line) > 0 {
			parts := strings.Fields(line)
			name := parts[0]

			if strings.HasPrefix(name, "dpkg-query:") {
				name = parts[len(parts)-1]

				if strings.Contains(name, ":") {
					name = strings.Split(name, ":")[0]
				}
			}

			// if name is empty, it might be not what we want
			if name == "" {
				continue
			}

			version := parts[len(parts)-1]

			if !regexp.MustCompile(`^\d`).MatchString(version) {
				version = ""
			}

			pkg := (*packages)[name]
			delete(*packages, name)

			if strings.HasPrefix(line, "dpkg-query: ") {
				pkg.Status = internal.PackageStatusUnknown
				pkg.Version = ""
			} else if parts[len(parts)-2] == "installed" {
				pkg.Status = internal.PackageStatusInstalled
				pkg.Version = version
			} else {
				pkg.Status = internal.PackageStatusAvailable
				pkg.Version = version
			}

			packagesList = append(packagesList, pkg)
		} else {
			fmt.Printf("apt: line is empty\n")
		}
	}

	return packagesList, nil
}

func ParsePackageInfoOutput(output string, opts *internal.Options) internal.PackageInfo {
	// remove the last empty line
	output = strings.TrimSuffix(output, "\n")
	lines := strings.Split(string(output), "\n")

	var pkg internal.PackageInfo

	for _, line := range lines {
		if len(line) > 0 {
			parts := strings.SplitN(line, ":", 2)

			if len(parts) != 2 {
				continue
			}

			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			switch key {
			case "Package":
				pkg.Name = value
			case "Version":
				pkg.Version = value
			case "Architecture":
				pkg.Arch = value
			case "Section":
				pkg.Category = value
			}
		}
	}

	pkg.PackageManager = "apt"

	return pkg
}