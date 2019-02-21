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
	"runtime"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/mitchellh/go-homedir"
	"github.com/ryanuber/columnize"
	"github.com/spf13/cobra"
)

type InfoMeta struct {
	CurrentConsulVersion    string
	CurrentNomadVersion     string
	CurrentPackerVersion    string
	CurrentTerraformVersion string
	CurrentVagrantVersion   string
	CurrentVaultVersion     string
	HostArch                string
	HostName                string
	HostOS                  string
	HvmHome                 string
	LogFile                 string
	UserHome                string
}

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Host information and current versions",
	Long: `Hashi Version Manager (hvm) is mostly a tongue in cheek personal project,
but is also quite real; it is not associated with HashiCorp in any official
capacity whatsoever, but allows you to manage multiple installations of their
popular CLI tools on supported platforms.`,
    Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return errors.New("arguments to info are not allowed")
		}
		return cobra.OnlyValidArgs(cmd, args)
	},
	Run: func(cmd *cobra.Command, args []string) {
		m := InfoMeta{}
		userHome, err := homedir.Dir()
		if err != nil {
			fmt.Println(fmt.Sprintf("failed to access home directory with error: %v", err))
			os.Exit(1)
		}
		m.UserHome = userHome
		m.HvmHome = fmt.Sprintf("%s/.hvm", m.UserHome)
		m.LogFile = fmt.Sprintf("%s/.hvm/hvm.log", m.UserHome)
		m.HostArch = runtime.GOARCH
		m.HostOS = runtime.GOOS
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
		hostName, err := os.Hostname()
		if err != nil {
			logger.Error("info", "cannot determine hostname with error", err.Error())
		}
		m.HostName = hostName

		consulV, err := CheckActiveVersion(Consul)
		if err != nil {
			logger.Error("info", "cannot determine active Consul version with error", err.Error())
			m.CurrentConsulVersion = ""
		}
		m.CurrentConsulVersion = consulV

		nomadV, err := CheckActiveVersion(Nomad)
		if err != nil {
			logger.Error("info", "cannot determine active Nomad version with error", err.Error())
			m.CurrentNomadVersion = ""
		}
		m.CurrentNomadVersion = nomadV

		packerV, err := CheckActiveVersion(Packer)
		if err != nil {
			logger.Error("info", "cannot determine active Packer version with error", err.Error())
			m.CurrentPackerVersion = ""
		}
		m.CurrentPackerVersion = packerV

		terraformV, err := CheckActiveVersion(Terraform)
		if err != nil {
			logger.Error("info", "cannot determine active Terraform version with error", err.Error())
			m.CurrentTerraformVersion = ""
		}
		m.CurrentTerraformVersion = terraformV

        // Problem with vagrant binary in Linux container on ChromeOS due to FUSE / AppImage issue
        // so if we learn that CROS is present, we skip checking Vagrant version since it cannot be determined
        if _, err := os.Stat("/dev/.cros_milestone"); os.IsNotExist(err) {
			vagrantV, err := CheckActiveVersion(Vagrant)
			if err != nil {
				logger.Error("info", "cannot determine active Vagrant version with error", err.Error())
				m.CurrentVagrantVersion = ""
			}
			m.CurrentVagrantVersion = vagrantV
		} else {
			m.CurrentVagrantVersion = ""
		}
		vaultV, err := CheckActiveVersion(Vault)
		if err != nil {
			logger.Error("info", "cannot determine active Vault version with error", err.Error())
			m.CurrentVaultVersion = ""
		}
		m.CurrentVaultVersion = vaultV
		infoData := map[string]string{"OS": m.HostOS, "Architecture": m.HostArch}
		t := time.Now()
		infoData["Date/Time"] = t.Format("Mon Jan _2 15:04:05 2006")
		if m.CurrentConsulVersion != "" {
			infoData["Consul version"] = m.CurrentConsulVersion
		}
		if m.CurrentNomadVersion != "" {
			infoData["Nomad version"] = m.CurrentNomadVersion
		}
		if m.CurrentPackerVersion != "" {
			infoData["Packer version"] = m.CurrentPackerVersion
		}
		if m.CurrentTerraformVersion != "" {
			infoData["Terraform version"] = m.CurrentTerraformVersion
		}
		if m.CurrentVagrantVersion != "" {
			infoData["Vagrant version"] = m.CurrentVagrantVersion
		}
		if m.CurrentVaultVersion != "" {
			infoData["Vault version"] = m.CurrentVaultVersion
		}
		columns := []string{}
		for k, v := range infoData {
			columns = append(columns, fmt.Sprintf("%s: | %s ", k, v))
		}
		data := columnize.SimpleFormat(columns)
		fmt.Println("Basic local system factoids:\n")
		fmt.Printf("%s\n", data)
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
