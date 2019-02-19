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
	"os/exec"
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

	// Consul binary name
	Consul string = "consul"

	// ConsulTemplate binary name
	ConsulTemplate string = "consul-template"

	// EnvConsul binary name
	EnvConsul string = "envconsul"

	// Nomad binary name
	Nomad string = "nomad"

	// Packer binary name
	Packer string = "packer"

	// Sentinel binary name
	Sentinel string = "sentinel"
	// Terraform binary name
	Terraform string = "terraform"

	// Vagrant binary name
	Vagrant string = "vagrant"

	// Vault binary name
	Vault string = "vault"
)

// HelpersMeta contains data for use by the helper functions
type HelpersMeta struct {
	BinaryArch          string
	BinaryName          string
	BinaryOS            string
	BinaryCheckVersion  string
	BinaryLatestVersion string `json:"current_version"`
	LogFile             string
	UserHome            string
	HvmHome             string
}

// CheckHashiVersion attempts to locate binary tools and get their current versions using OS calls
// Consul has slightly different version output style so it must be handled differently
func CheckHashiVersion(checkBinary string) (string, error) {
	installedVersion := ""
	m := HelpersMeta{}
	userHome, err := homedir.Dir()
	if err != nil {
		return installedVersion, fmt.Errorf("Unable to determine user home directory; error: %v", err)
	}
	m.UserHome = userHome
	m.HvmHome = fmt.Sprintf("%s/.hvm", m.UserHome)
	m.LogFile = fmt.Sprintf("%s/hvm.log", m.HvmHome)
	m.BinaryArch = runtime.GOARCH
	m.BinaryOS = runtime.GOOS
	m.BinaryName = checkBinary
	f, err := os.OpenFile(m.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to open log file %s with error: %v", m.LogFile, err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	logger := hclog.New(&hclog.LoggerOptions{Name: "hvm", Level: hclog.LevelFromString("INFO"), Output: w})
	path, err := exec.LookPath(checkBinary)
	if err != nil {
		logger.Error("info", "error detecting binary on PATH", checkBinary, "error", err.Error())
		return "", err
	}
	var version []byte
	if checkBinary == Consul {
		version, err = exec.Command("/bin/sh", "-c", fmt.Sprintf("%s version | head -n 1 | awk '{print $2}'", path)).Output()
		if err != nil {
			logger.Error("info", "error executing binary", checkBinary, "error", err.Error())
			return "", err
		}
		return string(version), nil
	} else if checkBinary == Nomad || checkBinary == Vault {
		version, err = exec.Command("/bin/sh", "-c", fmt.Sprintf("%s version | awk '{print $2}'", path)).Output()
		if err != nil {
			logger.Error("info", "error executing binary", checkBinary, "error", err.Error())
			return "", err
		}
		return string(version), nil
	}
	return "", err
}

// GetAllVersions grabs a list of all valid versions of an open source HashiTool from the releases website
func GetAllVersions(checkBinary string) (string, error) {
	allVersions := "wow"
	return allVersions, nil
}

// IsInstalledVersion determines if a given version is already installed
func IsInstalledVersion(checkBinary string, checkVersion string) (bool, error) {
	installedVersion := false
	m := HelpersMeta{}
	userHome, err := homedir.Dir()
	if err != nil {
		return installedVersion, fmt.Errorf("Unable to determine user home directory; error: %v", err)
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
		return installedVersion, fmt.Errorf("Failed to open log file with error: %v", err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	logger := hclog.New(&hclog.LoggerOptions{Name: "hvm", Level: hclog.LevelFromString("INFO"), Output: w})
	logger.Debug("helper", "is-installed-version", m.BinaryName, "check version", m.BinaryCheckVersion)
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
func IsValidVersion(checkBinary string, checkVersion string) (bool, error) {
	validVersion := false
	m := HelpersMeta{}
	userHome, err := homedir.Dir()
	if err != nil {
		return validVersion, fmt.Errorf("Unable to determine user home directory; error: %v", err)
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
		return validVersion, fmt.Errorf("Failed to open log file with error: %v", err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	logger := hclog.New(&hclog.LoggerOptions{Name: "hvm", Level: hclog.LevelFromString("INFO"), Output: w})
	logger.Debug("helper", "isvalidversion", m.BinaryName, "check version", m.BinaryCheckVersion)
	return validVersion, nil
}
