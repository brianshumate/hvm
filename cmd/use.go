// Copyright © 2019 Brian Shumate <brian@brianshumate.com>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice,
//    this list of conditions and the following disclaimer.
//
// 2. Redistributions in binary form must reproduce the above copyright notice,
//    this list of conditions and the following disclaimer in the documentation
//    and/or other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.

package cmd

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

// UseMeta contains data for using a binary version
type UseMeta struct {
	BinaryArch           string
	BinaryName           string
	BinaryOS             string
	BinaryDesiredVersion string
	LogFile              string
	UserHome             string
	HvmHome              string
}

// useCmd represents the use command
var useCmd = &cobra.Command{
	Use:   "use (<binary>) [--version <version>]",
	Short: "Use a specific binary version",
	Long: `
Use a supported binary binary at specified version.
The --version flag is required.

hvm can use the following binaries:

* consul
* consul-template (WIP)
* envconsul (WIP)
* nomad
* packer
* sentinel (WIP)
* terraform
* vagrant
* vault`,
	Example: `
  hvm use --help

  hvm use vault --version 1.0.2`,
	ValidArgs: []string{"consul",
		"consul-template",
		"envconsul",
		"nomad",
		"packer",
		"sentinel",
		"terraform",
		"vagrant",
		"vault"},
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		m := UseMeta{}
		userHome, err := homedir.Dir()
		if err != nil {
			fmt.Println(fmt.Sprintf("Cannot access home directory with error: %v", err))
			os.Exit(1)
		}
		m.UserHome = userHome
		m.HvmHome = fmt.Sprintf("%s/.hvm", m.UserHome)
		m.LogFile = fmt.Sprintf("%s/hvm.log", m.HvmHome)
		m.BinaryArch = runtime.GOARCH
		m.BinaryDesiredVersion = binaryVersion
		m.BinaryOS = runtime.GOOS
		m.BinaryName = strings.Join(args, " ")
		b := m.BinaryName
		v := m.BinaryDesiredVersion
		if _, err := os.Stat(m.HvmHome); os.IsNotExist(err) {
			err = os.Mkdir(m.HvmHome, 0755)
			if err != nil {
			fmt.Println(fmt.Sprintf("Cannot create directory %s with error: %v", m.HvmHome, err))
			os.Exit(1)
			}
		}
		f, err := os.OpenFile(m.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println(fmt.Sprintf("Cannot open log file %s with error: %v", m.LogFile, err))
			os.Exit(1)
		}
		defer f.Close()
		w := bufio.NewWriter(f)
		logger := hclog.New(&hclog.LoggerOptions{Name: "hvm", Level: hclog.LevelFromString("INFO"), Output: w})
		logger.Info("use", "run", "start with binary", b, "desired version", v)

		err = useBinary(&m)
		if err != nil {
			fmt.Println(fmt.Sprintf("Cannot use binary %s with error: %v", b, err))
			os.Exit(1)
		}
	},
}

// Initialize the command
func init() {
	rootCmd.AddCommand(useCmd)
	useCmd.PersistentFlags().StringVar(&binaryVersion,
		"version",
		"",
		"use binary version")
	useCmd.MarkFlagRequired("version")
}

func useBinary(m *UseMeta) error {
	b := m.BinaryName
	v := m.BinaryDesiredVersion
	f, err := os.OpenFile(m.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("Cannot open log file with error: %v", err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	logger := hclog.New(&hclog.LoggerOptions{Name: "hvm", Level: hclog.LevelFromString("INFO"), Output: w})
	logger.Debug("use", "f-use-binary", b)
	if m.BinaryName == "" {
		m.BinaryName = "none"
		logger.Error("use", "unknown-binary-candidate", "GURU DEDICATION EMISSINGVERSION")
		return fmt.Errorf("Unknown binary; please specify binary name as first argument")
	}
	if m.BinaryDesiredVersion == "" {
		logger.Debug("use", "f-use-binary", b)
		return fmt.Errorf("Unknown binary version; please specify version with '--version' flag")
	}
	logger.Info("use", "binary", b, "desired-version", v)

	// Is desired binary version valid?
	vv, err := ValidVersion(b, v)
	if err != nil {
		fmt.Println(fmt.Sprintf("Cannot determine if %s version %s is valid: %v", b, v, err))
		os.Exit(1)
	} else {
		if vv == false {
			fmt.Println(fmt.Sprintf("%s is not a version of %s hvm can use", v, b))
			os.Exit(1)
		}
	}

	// Is desired binary already installed?
	var installedVersion bool
	installedVersion, err = InstalledVersion(b, v)
	if err != nil {
		fmt.Println(fmt.Sprintf("Cannot determine if %s version %s is installed: %v", b, v, err))
		os.Exit(1)
	}
	if installedVersion == true {
		logger.Debug("use", "binary", b, "version", v, "installed", "true")
	} else {
		fmt.Println(fmt.Sprintf("%s version %s is not installed; install it with: hvm install %s --version %s", b, v, b, v))
		os.Exit(1)
	}
	srcPath := fmt.Sprintf("%s/%s/%s/%s", m.HvmHome, b, v, b)
	destPath := fmt.Sprintf("%s/bin/%s", m.UserHome, b)
	// Handle the binary symbolic link with jazz-like hands...
	if fi, err := os.Lstat(destPath); err == nil {
		if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
			if err = os.Remove(destPath); err != nil {
				return fmt.Errorf("failed to unlink %s with error: %+v", destPath, err)
			}
		} else {
			return fmt.Errorf("Path %s exists and is not a symbolic link created by hvm.\nhvm needs your help to resolve this problem; please inspect and move %s, thanks.", destPath, destPath)
		}
	}
	// XXX: yarrr
	// else if os.IsNotExist(err) {
	//     return fmt.Errorf("failed to resolve symbolic link: %+v", err)
	// }
	err = os.Symlink(srcPath, destPath)
	if err != nil {
		logger.Error("install", "f-use-binary", "symlink", "error", err)
		return err
	}
	fmt.Println(fmt.Sprintf("Using %s (%s/%s) version %s", b, m.BinaryOS, m.BinaryArch, v))
	return nil
}
