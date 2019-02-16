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
	"errors"
	"fmt"
	"os"
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
			return errors.New("use requires exactly 1 argument")
		}
		return cobra.OnlyValidArgs(cmd, args)
	},
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("FIXME: use called, but not yet implemented.")
		m := UseMeta{}
		userHome, err := homedir.Dir()
		if err != nil {
			fmt.Printf("Unable to determine user home directory; error: %s", err)
			panic(err)
		}
		m.UserHome = userHome
		m.HvmHome = fmt.Sprintf("%s/.hvm", m.UserHome)
		m.LogFile = fmt.Sprintf("%s/.hvm/hvm.log", m.UserHome)
		m.BinaryDesiredVersion = binaryVersion
		m.BinaryName = strings.Join(args, " ")
		if _, err := os.Stat(m.HvmHome); os.IsNotExist(err) {
			os.Mkdir(m.HvmHome, 0755)
		}
		f, err := os.OpenFile(m.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("Failed to open log file with error: %s", err)
			panic(err)
		}
		defer f.Close()
		w := bufio.NewWriter(f)
		logger := hclog.New(&hclog.LoggerOptions{Name: "hvm", Level: hclog.LevelFromString("INFO"), Output: w})
		logger.Info("use", "run", "start with binary", m.BinaryName, "desired version", m.BinaryDesiredVersion)
		err = useBinary(&m)
		if err != nil {
			fmt.Printf("Cannot use binary: %s\n", err)
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
}

func useBinary(m *UseMeta) error {
	f, err := os.OpenFile(m.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Failed to open log file with error: %s", err)
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	logger := hclog.New(&hclog.LoggerOptions{Name: "hvm", Level: hclog.LevelFromString("INFO"), Output: w})
	logger.Debug("use", "f-use-binary", "start", "with-binary", m.BinaryName)
	if m.BinaryName == "" {
		m.BinaryName = "none"
		logger.Error("use", "unknown-binary-candidate", "🌀 GURU DEDICATION EMISSINGVERSION")
		err = errors.New("use: unknown binary version; please use --version <version>")
		return err
	}
	if m.BinaryDesiredVersion == "" {
		logger.Debug("install", "f-use-binary", "blank-version", "binary", m.BinaryName)
		err = errors.New("use: unknown binary version; please use --version <version>")
		return err
	}
	logger.Info("use", "use binary candidate", "final", "binary", m.BinaryName, "desired-version", m.BinaryDesiredVersion)
	fmt.Println("Using binary ", m.BinaryName)
	fmt.Println("Using version ", m.BinaryDesiredVersion)
	return nil
}
