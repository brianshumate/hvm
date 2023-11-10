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
	"sort"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/mitchellh/go-homedir"
	"github.com/ryanuber/columnize"
	"github.com/spf13/cobra"
)

type InfoMeta struct {
	CurrentConsulVersion 	string
	CurrentNomadVersion  	string
	CurrentPackerVersion	string
	CurrentTerraformVersion	string
	CurrentVagrantVersion	string
	CurrentVaultVersion  	string
	HostArch             	string
	HostName             	string
	HostOS               	string
	HvmHome              	string
	LogFile              	string
	UserHome             	string
}

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Host system information and current versions",
	Long: `Hashi Version Manager (hvm) is mostly a tongue in cheek personal
project, but is also quite real; it is not associated with HashiCorp in any
official capacity whatsoever, but allows you to manage multiple installations
of their popular CLI tools on supported platforms.`,
    Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
			m := InfoMeta{}
			userHome, err := homedir.Dir()
			if err != nil {
				fmt.Println(fmt.Sprintf("cannot access home directory with error: %v", err))
				os.Exit(1)
			}
			m.UserHome = userHome
			m.HvmHome = fmt.Sprintf("%s/.hvm", m.UserHome)
			m.LogFile = fmt.Sprintf("%s/hvm.log", m.HvmHome)
			m.HostArch = runtime.GOARCH
			m.HostOS = runtime.GOOS
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

            // System info
            hostName, err := os.Hostname()
			if err != nil {
				logger.Error("info", "Cannot determine hostname", "error:", err.Error())
			}
			m.HostName = hostName
			s := map[string]string{"OS": m.HostOS, "Architecture": m.HostArch}
			t := time.Now()
			s["Date/Time"] = t.Format("Mon Jan _2 15:04:05 2006")
			si := []string{}
			for k, v := range s {
				si = append(si, fmt.Sprintf("%s: | %s ", k, v))
			}
			// sort.Strings(si)
			systemData := columnize.SimpleFormat(si)

			// Version info
			v := map[string]string{}
			consulV, err := ActiveLocalVersion(Consul)
			if err != nil {
				logger.Error("info", "cannot determine version", "consul", "error", err.Error())
			}
			if consulV != "" {
				m.CurrentConsulVersion = consulV
				v["Consul"] = m.CurrentConsulVersion
            }
			nomadV, err := ActiveLocalVersion(Nomad)
			if err != nil {
				logger.Error("info", "cannot determine version", "nomad", "error", err.Error())
			}
			if nomadV != "" {
				m.CurrentNomadVersion = nomadV
				v["Nomad"] = m.CurrentNomadVersion
            }
			vaultV, err := ActiveLocalVersion(Vault)
			if err != nil {
				logger.Error("info", "cannot determine version", "vault", "error", err.Error())
			}
			if vaultV != "" {
				m.CurrentVaultVersion = vaultV
				v["Vault"] = m.CurrentVaultVersion
			}
            vi := []string{}
			for k, v := range v {
				vi = append(vi, fmt.Sprintf("%s: | %s ", k, v))
			}
			sort.Strings(vi)
			versionData := columnize.SimpleFormat(vi)

            // Display all
			fmt.Println("System Factoids")
			fmt.Println("")
			fmt.Println(systemData)
			fmt.Println("")
			fmt.Println("Installed Versions")
			fmt.Println("")
			fmt.Println(versionData)
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
