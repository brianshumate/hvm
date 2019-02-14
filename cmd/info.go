// Copyright © 2018 Brian Shumate <brian@brianshumate.com>
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
	"os/exec"
	"runtime"
	// "strings"
	"time"

	"github.com/ryanuber/columnize"
	"github.com/mitchellh/go-homedir"
	"github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"
)

type InfoMeta struct {
	CurrentConsulVersion	string
	CurrentVaultVersion		string
	HostArch				string
	HostName 				string
	HostOS					string
	HvmHome     		        string
	LogFile					string
	UserHome 				string
}

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Host information and current versions",
	Long: `Hashi Version Manager (hvm) is mostly a tongue in cheek personal project,
but is also quite real; it is not associated with HashiCorp in any official
capacity whatsoever, but allows you to manage multiple installations of their
popular CLI tools on supported platforms.`,
	Run: func(cmd *cobra.Command, args []string) {
		m := InfoMeta{}
		userHome, err := homedir.Dir()
		if err != nil {
			fmt.Printf("Unable to determine user home directory; error: %s", err)
			panic(err)
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
			fmt.Printf("Failed to open log file with error: %s", err)
			panic(err)
		}
		defer f.Close()
		w := bufio.NewWriter(f)
		logger := hclog.New(&hclog.LoggerOptions{Name: "hvm", Level: hclog.LevelFromString("INFO"), Output: w})
		hostName, err := os.Hostname()
		if err != nil {
			logger.Error("info", "Cannot determine hostname!")
		    panic(err)
		}
		m.HostName = hostName
		consulV, err := checkHashiVersion(&m, "consul")
		if err != nil {
 			logger.Error("info", "cannot determine version", "consul", "error", err.Error() )
		}
		m.CurrentConsulVersion = consulV

		vaultV, err := checkHashiVersion(&m, "vault")
		if err != nil {
 			logger.Error("info", "cannot determine version", "vault", "error", err.Error() )
		}
		m.CurrentVaultVersion = vaultV

		infoData := map[string]string{"OS": m.HostOS, "Architecture": m.HostArch}
		t := time.Now()
		infoData["Date/Time"] = t.Format("Mon Jan _2 15:04:05 2006")
        fmt.Sprintf("DEBUG: adding Consul version %s", m.CurrentConsulVersion)
		if m.CurrentConsulVersion != "ENOVERSION" {
			infoData["Consul version"] = m.CurrentConsulVersion
		}
		if m.CurrentVaultVersion != "ENOVERSION" {
			infoData["Vault version"] = m.CurrentVaultVersion
		}
		columns := []string{}
		for k, v := range infoData {
			columns = append(columns, fmt.Sprintf("%s: | %s ", k, v))
		}
		data := columnize.SimpleFormat(columns)
		fmt.Println("Basic system factoids:")
		fmt.Printf("%s\n", data)
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// infoCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// infoCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// checkHashiVersion attempts to locate the tools and get their versions -
// Consul has slightly different version output style so it must be handled differently
func checkHashiVersion(m *InfoMeta, name string) (string, error) {
	f, err := os.OpenFile(m.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Failed to open log file with error: %s", err)
		panic(err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	logger := hclog.New(&hclog.LoggerOptions{Name: "hvm", Level: hclog.LevelFromString("INFO"), Output: w})
    path, err := exec.LookPath(name)
	if err != nil {
		logger.Error("info", "error detecting binary on PATH", name, "error", err.Error())
		return "", err
	}
	if name == "consul" {
		version, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("%s version | head -n 1 | awk '{print $2}'", path)).Output()
		if err != nil {
			logger.Error("info", "error executing binary", name, "error", err.Error())
			return "", err
		}
		return string(version), nil
	} else if name == "nomad" || name == "vault" {
		version, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("%s version | awk '{print $2}'", path)).Output()
		if err != nil {
			logger.Error("info", "error executing binary", name, "error", err.Error())
			return "", err
		}
		return string(version), nil
	}
	return "", err
}
