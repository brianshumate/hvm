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
	// "bytes"
	// "encoding/json"
	// "errors"
	"fmt"
	// "golang.org/x/net/html"
	// "io"
	// "io/ioutil"
	// "net/http"
	"os"
	"runtime"
	// "strings"
	// "time"

	"github.com/hashicorp/go-hclog"
	// "github.com/hashicorp/go-version"
	"github.com/mitchellh/go-homedir"
)

const (
	// CheckpointURLBase is the URL base for CheckPoint API
	CheckpointURLBase string = "https://checkpoint-api.hashicorp.com"
	// ReleaseURLBase is the URL base for the HashiCorp releases website
	ReleaseURLBase string = "https://releases.hashicorp.com"
	// VaultReleaseURLBase is the URL base for the Vault releases page
	VaultReleaseURLBase string = "https://releases.hashicorp.com/vault/"
)

// HelpersMeta contains data for use by the helper functions
type HelpersMeta struct {
	BinaryArch           string
	BinaryName           string
	BinaryOS             string
	BinaryCheckVersion string
	BinaryLatestVersion  string `json:"current_version"`
	LogFile              string
	UserHome             string
	HvmHome              string
}

// GetAllVersions grabs a list of all valid versions of an open source HashiTool from the releases website
func GetAllVersions (checkBinary string) (string, error) {
	allVersions := "wow"
	return allVersions, nil
}

// IsInstalledVersion determines if a given version is already installed
func IsInstalledVersion (checkBinary string, checkVersion string) (bool, error) {
		m := HelpersMeta{}
		userHome, err := homedir.Dir()
		if err != nil {
			fmt.Errorf("Unable to determine user home directory; error: %v", err)
		}
		m.UserHome = userHome
		m.HvmHome = fmt.Sprintf("%s/.hvm", m.UserHome)
		m.LogFile = fmt.Sprintf("%s/hvm.log", m.HvmHome)
		m.BinaryArch = runtime.GOARCH
		m.BinaryCheckVersion = checkVersion
		m.BinaryOS = runtime.GOOS
		m.BinaryName = checkBinary
		if _, err := os.Stat(m.HvmHome); os.IsNotExist(err) {
			os.Mkdir(m.HvmHome, 0755)
		}
		f, err := os.OpenFile(m.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Errorf("Failed to open log file with error: %v", err)
		}
		defer f.Close()
		w := bufio.NewWriter(f)
		logger := hclog.New(&hclog.LoggerOptions{Name: "hvm", Level: hclog.LevelFromString("INFO"), Output: w})
		logger.Debug("helper", "isinstalledversion", m.BinaryName, "check version", m.BinaryCheckVersion)
	installedVersion := false
	fullPath := fmt.Sprintf("%s/%s/%s", m.HvmHome, m.BinaryName, m.BinaryCheckVersion)
    // :phew:
    _, err = os.Stat(fullPath)
    if err != nil {
        if os.IsNotExist(err) {
            installedVersion = false
        }
    } else {
    	installedVersion = true
    }
    return installedVersion, nil
}

// IsValidVersion determines if 	a given version is contained in the list from GetAllVersions
func IsValidVersion (checkBinary string, checkVersion string) (bool, error) {
	m := HelpersMeta{}
		userHome, err := homedir.Dir()
		if err != nil {
			fmt.Errorf("Unable to determine user home directory; error: %v", err)
		}
		m.UserHome = userHome
		m.HvmHome = fmt.Sprintf("%s/.hvm", m.UserHome)
		m.LogFile = fmt.Sprintf("%s/hvm.log", m.HvmHome)
		m.BinaryArch = runtime.GOARCH
		m.BinaryCheckVersion = checkVersion
		m.BinaryOS = runtime.GOOS
		m.BinaryName = checkBinary
		if _, err := os.Stat(m.HvmHome); os.IsNotExist(err) {
			os.Mkdir(m.HvmHome, 0755)
		}
		f, err := os.OpenFile(m.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Errorf("Failed to open log file with error: %v", err)
		}
		defer f.Close()
		w := bufio.NewWriter(f)
		logger := hclog.New(&hclog.LoggerOptions{Name: "hvm", Level: hclog.LevelFromString("INFO"), Output: w})
		logger.Debug("helper", "isvalidversion", m.BinaryName, "check version", m.BinaryCheckVersion)
	validVersion := false

	return validVersion, nil
}
