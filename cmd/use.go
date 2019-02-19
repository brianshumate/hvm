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
	// "errors"
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
Use the specified binary version; this is not
accomplished with symbolic links and instead relies
on copying the specified binary to the hvm binary
directory which should be in the PATH.`,
	Example: `
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
	// Using a custom args function here as workaround for GH-745
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("use requires exactly 1 argument")
		}
		return cobra.OnlyValidArgs(cmd, args)
	},
	Run: func(cmd *cobra.Command, args []string) {
		m := UseMeta{}
		userHome, err := homedir.Dir()
		if err != nil {
			fmt.Println(fmt.Sprintf("failed to access home directory with error: %v", err))
			os.Exit(1)
        }
		m.UserHome = userHome
		m.HvmHome = fmt.Sprintf("%s/.hvm", m.UserHome)
		m.LogFile = fmt.Sprintf("%s/.hvm/hvm.log", m.UserHome)
		m.BinaryArch = runtime.GOARCH
		m.BinaryDesiredVersion = binaryVersion
		m.BinaryOS = runtime.GOOS
		m.BinaryName = strings.Join(args, " ")
		if _, err := os.Stat(m.HvmHome); os.IsNotExist(err) {
			os.Mkdir(m.HvmHome, 0755)
		}
		f, err := os.OpenFile(m.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println(fmt.Sprintf("failed to open log file with error: %v", err))
			os.Exit(1)
		}
		defer f.Close()
		w := bufio.NewWriter(f)
		logger := hclog.New(&hclog.LoggerOptions{Name: "hvm", Level: hclog.LevelFromString("INFO"), Output: w})
		logger.Info("use", "run", "start with binary", m.BinaryName, "desired version", m.BinaryDesiredVersion)

        // Validate binary attributes with helper functions

        // Is binary already installed?
        var installedV bool
		installedV, err = IsInstalledVersion(m.BinaryName, m.BinaryDesiredVersion)
		if err != nil {
			fmt.Println(fmt.Sprintf("cannot use binary: %s with error: %v", m.BinaryName, err))
			os.Exit(1)
		}
        if installedV == true {
        	if m.BinaryDesiredVersion == "" {
        		fmt.Println(fmt.Sprintf("latest %s version installed", m.BinaryName))
        		os.Exit(1)
        		} else {
        			fmt.Println(fmt.Sprintf("%s version %s appears to be already installed", m.BinaryName, m.BinaryDesiredVersion))
        			os.Exit(1)
        		}
        }
		err = useBinary(&m)
		if err != nil {
			fmt.Println(fmt.Sprintf("cannot use binary: %s with error: %v", m.BinaryName, err))
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
	f, err := os.OpenFile(m.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Errorf("failed to open log file with error: %v", err)
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	logger := hclog.New(&hclog.LoggerOptions{Name: "hvm", Level: hclog.LevelFromString("INFO"), Output: w})
	logger.Debug("use", "f-use-binary", "start", "with-binary", m.BinaryName)
	if m.BinaryName == "" {
		m.BinaryName = "none"
		logger.Error("use", "unknown-binary-candidate", "GURU DEDICATION EMISSINGVERSION")
		return fmt.Errorf("use: unknown binary; please specify binary name as first argument")
	}
	if m.BinaryDesiredVersion == "" {
		logger.Debug("install", "f-use-binary", "blank-version", "binary", m.BinaryName)
		return fmt.Errorf("use: unknown binary version; please specify version with '--version' flag")
	}
	logger.Info("use", "use binary candidate", "final", "binary", m.BinaryName, "desired-version", m.BinaryDesiredVersion)
	srcPath := fmt.Sprintf("%s/%s/%s/%s", m.HvmHome, m.BinaryName, m.BinaryDesiredVersion, m.BinaryName)
    destPath := fmt.Sprintf("%s/bin/%s", m.UserHome, m.BinaryName)

    // Handle the binary symbolic link with jazz-like hands...
    if _, err := os.Lstat(destPath); err == nil {
    	if err := os.Remove(destPath); err != nil {
        	return fmt.Errorf("failed to unlink %s with error: %+v", destPath, err)
    	}
    }
      // XXX: yarrr
      // else if os.IsNotExist(err) {
      //     return fmt.Errorf("failed to resolve symbolic link: %+v", err)
      // }
    os.Symlink(srcPath, destPath)
    if err != nil {
      	logger.Error("install", "f-use-binary", "symlink", "error", err)
      	return err
    }
    fmt.Println(fmt.Sprintf("Now using %s (%s/%s) version %s", m.BinaryName, m.BinaryOS, m.BinaryArch, m.BinaryDesiredVersion))
	return nil
}
