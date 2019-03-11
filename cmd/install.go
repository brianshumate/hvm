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
	"bytes"
	//"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/hashicorp/go-getter"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-version"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

// InstallMeta contains data for a binary installation candidate
type InstallMeta struct {
	BinaryArch           string
	BinaryName           string
	BinaryOS             string
	BinaryDesiredVersion string
	BinaryLatestVersion  string `json:"current_version"`
	LogFile              string
	UserHome             string
	HvmHome              string
}

var binaryVersion string

// installCmd downloads, extracts, and installs a binary into the hvm home path
var installCmd = &cobra.Command{
	Use:   "install (<binary>) [--version <version>]",
	Short: "Install binary at latest available or specified version",
	Long: `
Install a supported binary binary at specified version for the host detected
architecture and operating system; if the version flag is omitted, the latest
available version will be installed.

hvm can install the following binaries:

* consul
* consul-template (WIP)
* envconsul (WIP)
* nomad
* packer
* sentinel (WIP)
* terraform
* vagrant
* vault
`,
	Example: `
  hvm install --help

  hvm install vault

  hvm install nomad --version 0.8.5`,
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
		m := InstallMeta{}
		userHome, err := homedir.Dir()
		if err != nil {
			fmt.Println(fmt.Sprintf("failed to access home directory with error: %v", err))
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
			fmt.Println(fmt.Sprintf("Failed to create directory %s with error: %v", m.HvmHome, err))
			os.Exit(1)
		}
		}
		f, err := os.OpenFile(m.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println(fmt.Sprintf("Failed to open log file %s with error: %v", m.LogFile, err))
			os.Exit(1)
		}
		defer f.Close()
		w := bufio.NewWriter(f)
		logger := hclog.New(&hclog.LoggerOptions{Name: "hvm", Level: hclog.LevelFromString("INFO"), Output: w})
		// Validate binary attributes with helper functions

        // Is it a supported binary?
        s := []string{Consul, Nomad, Packer, Terraform, Vagrant, Vault}
        sb := false
        for _, v := range s {
    		if v == b {
        		sb = true
    		}
		}
		if sb != true {
			fmt.Println(fmt.Sprintf("Cannot install that.", b))
			os.Exit(1)
		}

		// Is desired binary version valid?
		if v != "" {
			vv, err := ValidateVersion(b, v)
			if err != nil {
				fmt.Println(fmt.Sprintf("Cannot determine if %s version %s is valid; error %v.", b, v, err))
				os.Exit(1)
			} else {
				if vv == false {
				fmt.Println(fmt.Sprintf("Cannot install %s version %s.", b, v))
				os.Exit(1)
				}
			}
		}


		// Is desired binary already installed?
		var installedVersion bool
		installedVersion, err = IsInstalledVersion(b, v)
		if err != nil {
			fmt.Println(fmt.Sprintf("Cannot install %s with error: %v.", b, err))
			os.Exit(1)
		}
		if installedVersion == true {
			if m.BinaryDesiredVersion == "" {
				fmt.Println(fmt.Sprintf("Latest %s version installed.", b))
				os.Exit(1)
			} else {
				fmt.Println(fmt.Sprintf("%s version %s appears to be already installed.", b, v))
				os.Exit(1)
			}
		} else {
			logger.Info("install", "run", b, "desired version", v)
			err = installBinary(&m)
			if err != nil {
				fmt.Println(fmt.Sprintf("Cannot install %s version %s with error: %v.", b, v, err))
				os.Exit(1)
			}
		}

	},
}

// Initialize the command
func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.PersistentFlags().StringVar(&binaryVersion,
		"version",
		"",
		"install binary version")
	installCmd.MarkFlagRequired("version")
}

// installBinary has entirely too much going on in it right now!
// some of this needs to possibly be refactored into helpers
func installBinary(m *InstallMeta) error {
	b := m.BinaryName
	v := m.BinaryDesiredVersion
	f, err := os.OpenFile(m.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("Failed to open log file %s with error: %v", m.LogFile, err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	logger := hclog.New(&hclog.LoggerOptions{Name: "hvm", Level: hclog.LevelFromString("INFO"), Output: w})
	logger.Debug("install", "f-install-binary", "start", "with-binary", b)
	if b == "" {
		b = "none"
		logger.Error("install", "unknown-binary", "GURU DEDICATION")
		return fmt.Errorf("install: unknown binary. GURU DEDICATION")
	}
	if v == "" {
		logger.Debug("install", "f-install-binary", "blank-version", "binary", b)
		latestBinaryVersion, err := GetLatestVersion(b)
		if err != nil {
			logger.Error("install", "get-latest-version-fail", "error", err.Error())
			return err
		}
		logger.Debug("install", "get-latest-version", "inner", "got-version", latestBinaryVersion)
		v = latestBinaryVersion
	}
	logger.Info("install", "install binary candidate", "final", "binary", b, "desired-version", v)

	switch b {
	case Consul, Nomad, Packer, Terraform, Vagrant, Vault:
		targetPath := fmt.Sprintf("%s/.hvm/%s/%s", m.UserHome, b, v)
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			if os.IsNotExist(err) {
				err := os.MkdirAll(targetPath, 0770)
				if err != nil {
					logger.Error("install", "directory-creation-error", err.Error())
					return fmt.Errorf("directory creation error: %v", err)
				}
			}
		}
		// Store <binary>_<version>_SHA256SUMS file obtained from
		// https://releases.hashicorp.com/<binary>/<version>/<binary>_<version>_SHA256SUMS
		// in map for comparison
		binaryShaURL := fmt.Sprintf("%s/%s/%s/%s_%s_SHA256SUMS", ReleaseURLBase, b, v, b, v)
		logger.Debug("install", "sha256sums-file-url", binaryShaURL)
		binarySha, err := FetchData(binaryShaURL)
		if err != nil {
			logger.Error("install", "download-sha256sums-error", err.Error())
			return err
		}
		shaStream := bytes.NewReader(binarySha)
		scanner := bufio.NewScanner(shaStream)
		fileSha := map[string]string{}
		for scanner.Scan() {
			s := strings.Fields(scanner.Text())
			if len(s) == 2 {
				if b == Nomad {
					logger.Debug("install", "stage", "scanner", "binary", Nomad)
					nomadLatestVersion, err := version.NewVersion(v)
					if err != nil {
						logger.Error("install", "issue", "Could not determine Nomad comparison version!", "error", err.Error())
						return err
					}
					constraints, err := version.NewConstraint(">= 0.7.0-beta1")
					if err != nil {
						logger.Error("install", "issue", "Could not determine Nomad version constraints", "error", err.Error())
						return err
					}
					// Handle the current Nomad SHA256SUMS style
					if constraints.Check(nomadLatestVersion) {
						logger.Debug("install", "newer-nomad-version", nomadLatestVersion, "constraints", constraints)
						fileSha[strings.Trim(s[1], "./")] = s[0]
					} else {
						// Handle older Nomad SHA256SUMS style
						fileSha[s[1]] = s[0]
					}
				} else {
					logger.Debug("install", "binary", b)
				}
			}
			fileSha[s[1]] = s[0]
		}
		if err := scanner.Err(); err != nil {
			logger.Error("install", "process-sha256sums-error", err.Error())
			return err
		}
		pkgFilename := fmt.Sprintf("%s_%s_%s_%s.zip", b, v, m.BinaryOS, m.BinaryArch)
		checkSha := fileSha[pkgFilename]
		fullURL := fmt.Sprintf("%s/%s/%s/%s?checksum=sha256:%s", ReleaseURLBase, b, v, pkgFilename, checkSha)
		installPath := fmt.Sprintf("%s/%s", targetPath, b)
		logger.Debug("install", "valid-binary", "true", "full-url", fullURL, "install-path", installPath)
		// Get binary archive using go-getter from a URL which takes the form of:
		// 'https://releases.hashicorp.com/<binary>/<version>/<binary>_<version>_<os>_<arch>.zip
		// go-getter validates the intended download against its published SHA256 summary before downloading, or fails if the there is mismatch / other issue which prevents comparison.
		// Shout out to Ye Olde School BSD spinner!
		hvmSpinnerSet := []string{"/", "|", "\\", "-", "|", "\\", "-"}
		s := spinner.New(hvmSpinnerSet, 174*time.Millisecond)
		s.Writer = os.Stderr
		err = s.Color("fgHiCyan")
		if err != nil {
			logger.Debug("install", "weird-error", err.Error())
		}
		s.Suffix = " Installing..."
		s.FinalMSG = fmt.Sprintf("Installed %s (%s/%s) version %s\n", b, m.BinaryOS, m.BinaryArch, v)
		s.Start()
		logger.Debug("install", "status", "go-getter", "download-url", fullURL)
		logger.Debug("install", "status", "go-getter", "install-path", installPath)
		if err := getter.GetFile(installPath, fullURL); err != nil {
			fmt.Printf("Download error with %q", err)
			// If the SHA don't match or we hit any issue, then we ain't dancing!
			logger.Error("install", "download-zip-error", err.Error())
			s.Stop()
			return err
		}
		s.Stop()
		return nil
	default:
		logger.Warn("install", "binary", b, "unsupported-binary", "not in CheckPoint API")
		return fmt.Errorf("Binary %s currently unsupported", b)
	}
}
