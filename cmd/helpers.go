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
//
// helpers are random functions which are often shared and intermingled amongst the
// commands and so just sort of hang out here for now...

package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-version"
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

// CheckActiveVersion tries to locate binary tools in the system path and get their version using OS calls
// 'consul version' has a slightly different output style from the others, and must be handled differently
func CheckActiveVersion(binary string) (string, error) {
	activeVersion := ""
	userHome, err := homedir.Dir()
	if err != nil {
		return activeVersion, fmt.Errorf("Cannot determine user home directory with error: %v", err)
	}
	m := HelpersMeta{}
	m.UserHome = userHome
	m.HvmHome = fmt.Sprintf("%s/.hvm", m.UserHome)
	m.LogFile = fmt.Sprintf("%s/hvm.log", m.HvmHome)
	m.BinaryArch = runtime.GOARCH
	m.BinaryOS = runtime.GOOS
	m.BinaryName = binary
	f, err := os.OpenFile(m.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return "", fmt.Errorf("Cannot open log file %s with error: %v", m.LogFile, err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	logger := hclog.New(&hclog.LoggerOptions{Name: "hvm", Level: hclog.LevelFromString("INFO"), Output: w})
	binPath, err := exec.LookPath(binary)
	if err != nil {
		logger.Error("helper", "cannot detect binary on PATH", binary, "error", err.Error())
		return "", fmt.Errorf("Cannot detect binary on PATH with error: %v", err)
	}
	var version []byte
	if binary == Consul {
		version, err = exec.Command("/bin/sh", "-c", fmt.Sprintf("%s version | head -n 1 | awk '{print $2}' | cut -d 'v' -f2", binPath)).Output()
		if err != nil {
			logger.Error("helper", "cannot execute binary", binary, "error", err.Error())
			return "", fmt.Errorf("Cannot execute binary with error: %v", err)
		}
		return string(version), nil
	} else {
		version, err = exec.Command("/bin/sh", "-c", fmt.Sprintf("%s version | awk '{print $2}' | cut -d 'v' -f2", binPath)).Output()
		if err != nil {
			logger.Error("helper", "cannot execute binary", binary, "error", err.Error())
			return "", fmt.Errorf("Cannot execute binary with error: %v", err)
		}
		return string(version), nil
	}
}

// FetchData grabs bits of HTML data over HTTP for some reason...
func FetchData(URL string) ([]byte, error) {
	userHome, err := homedir.Dir()
	if err != nil {
		return nil, fmt.Errorf("Cannot determine user home directory with error: %v", err)
	}
	m := HelpersMeta{}
	m.UserHome = userHome
	m.HvmHome = fmt.Sprintf("%s/.hvm", m.UserHome)
	m.LogFile = fmt.Sprintf("%s/hvm.log", m.HvmHome)
	f, err := os.OpenFile(m.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("Cannot open log file %s with error: %v", m.LogFile, err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	logger := hclog.New(&hclog.LoggerOptions{Name: "hvm", Level: hclog.LevelFromString("INFO"), Output: w})
	response, err := http.Get(URL)
	if err != nil {
		logger.Error("helper", "Cannot fetch data with error", err.Error())
		return nil, fmt.Errorf("cannot fetch data with error: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		err = errors.New(response.Status)
		logger.Error("helper", "Cannot fetch data with error", err.Error())
		return nil, fmt.Errorf("cannot failed to fetch data with error: %v", err)
	}
	var fetchData bytes.Buffer
	_, err = io.Copy(&fetchData, response.Body)
	if err != nil {
		logger.Error("helper", "cannot fetch data with error", err.Error())
		return nil, fmt.Errorf("Cannot fetch data with bytes buffer with error: %v", err)
	}
	return fetchData.Bytes(), nil
}

// GetLatestVersion returns the latest available binary version from releases.hashicorp.com
func GetLatestVersion(binary string) (string, error) {
	userHome, err := homedir.Dir()
	if err != nil {
		return "", fmt.Errorf("Cannot determine user home directory with error: %v", err)
	}
	m := HelpersMeta{}
	m.UserHome = userHome
	m.HvmHome = fmt.Sprintf("%s/.hvm", m.UserHome)
	m.LogFile = fmt.Sprintf("%s/hvm.log", m.HvmHome)
	f, err := os.OpenFile(m.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return "", fmt.Errorf("Cannot open log file %s with error: %v", m.LogFile, err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	logger := hclog.New(&hclog.LoggerOptions{Name: "hvm", Level: hclog.LevelFromString("INFO"), Output: w})
	logger.Debug("helper", "f-get-latest-version", binary)
	switch binary {
	// Some binary latest versions cannot be queried through the Checkpoint API.
	// Those binaries must unfortunately be queried using an HTML scraping approach instead.
	case Vault:
		logger.Debug("helper", "f-get-latest-version-html-scrape-url-base", VaultReleaseURLBase)
		logger.Debug("helper", "f-get-latest-version-html-scrape-binary-name", binary)
		var found bool
		resp, err := http.Get(VaultReleaseURLBase)
		if err != nil {
			return "", fmt.Errorf("Cannot get Vault release URL with error: %v", err)
		}
		defer resp.Body.Close()
		z := html.NewTokenizer(bufio.NewReader(resp.Body))
		for found == false {
			tt := z.Next()
			switch tt {
			case html.ErrorToken:
				return "", err
			case html.StartTagToken:
				t := z.Token()
				switch t.Data {
				case "a":
					z.Next()
					t = z.Token()
					if t.Data != "../" {
						latestVersion := strings.TrimPrefix(t.Data, "vault_")
						m.BinaryLatestVersion = latestVersion
						found = true
						break
					}
				default:
					continue
				}
			}
		}
	case Consul, Nomad, Packer, Vagrant, Terraform:
		logger.Debug("helper", "f-get-latest-version-checkpoint-url-base", CheckpointURLBase)
		logger.Debug("helper", "f-get-latest-version-checkpoint-binary-name", binary)
		checkpointDataURL := fmt.Sprintf("%s/v1/check/%s", CheckpointURLBase, binary)
		logger.Debug("helper", "f-get-latest-version-checkpoint-data-url", checkpointDataURL)
		checkPointClient := http.Client{Timeout: time.Second * 2}
		req, err := http.NewRequest(http.MethodGet, checkpointDataURL, nil)
		if err != nil {
			logger.Error("helper", "f-get-latest-version", "request-error", err.Error())
			return "", err
		}
		req.Header.Set("User-Agent", "hvm-oss-http-client")
		res, err := checkPointClient.Do(req)
		if err != nil {
			logger.Error("helper", "f-get-latest-version", "get-error", err.Error())
			return "", err
		}
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			logger.Error("helper", "f-get-latest-version", "read-body-error", err.Error())
			return "", err
		}
		err = json.Unmarshal(body, &m)
		if err != nil {
			logger.Error("helper", "f-get-latest-version", "json-unmarshall-error", err.Error())
			return "", fmt.Errorf("cannot unmarshal JSON with error: %v", err)
		}
		// Ensure that we get something like a valid version back from the API
		// and not a maintenance page or similar...
		checkpointLatestVersion, err := version.NewVersion(m.BinaryLatestVersion)
		if err != nil {
			logger.Error("helper", "issue", "cannot determine comparison version", "error", err.Error())
			return "", err
		}
		constraints, err := version.NewConstraint(">= 0.0.1")
		if err != nil {
			logger.Error("helper", "f-get-latest-version", "issue", "cannot determine comparison constraints", "error", err.Error())
			return "", err
		}
		if constraints.Check(checkpointLatestVersion) {
			logger.Debug("helper", "f-get-latest-version", "chcked-version", "version", checkpointLatestVersion, "constraints", constraints)
		} else {
			// Eh oh, something is wrong!
			logger.Error("helper", "f-get-latest-version", "issue", "unexpected-checkpoint-api-value", m.BinaryLatestVersion)
			return "", fmt.Errorf("problem determining latest binary version")
		}
		return m.BinaryLatestVersion, nil
	default:
		if m.BinaryName != Vault {
			logger.Warn("helper", "binary", m.BinaryName, "unsupported-binary", "Binary not in CheckPoint API or otherwise not supported.")
			return "", fmt.Errorf("Binary currently unsupported")
		}
	}
	return m.BinaryLatestVersion, nil
}

// IsInstalledVersion determines if specified binary version is already installed by hvm
func IsInstalledVersion(binary string, checkVersion string) (bool, error) {
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
	m.BinaryName = binary
	if _, err := os.Stat(m.HvmHome); os.IsNotExist(err) {
		err = os.Mkdir(m.HvmHome, 0755)
		if err != nil {
			return false, fmt.Errorf("failed to create directory %s with error: %v", m.HvmHome, err)
		}
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

// ValidateVersion accepts a binary name and version number then validates it against all versions
// from releases.hashicorp.com returning true if the proposed version number matches a version listed
// there or false if not found or an error occurs
func ValidateVersion(binary string, binaryVersion string) (bool, error) {
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
	m.BinaryCheckVersion = binaryVersion
	m.BinaryOS = runtime.GOOS
	m.BinaryName = binary
	if _, err := os.Stat(m.HvmHome); os.IsNotExist(err) {
		err = os.Mkdir(m.HvmHome, 0755)
		if err != nil {
			return false, fmt.Errorf("failed to create directory %s with error: %v", m.HvmHome, err)
		}
	}
	f, err := os.OpenFile(m.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return validVersion, fmt.Errorf("Failed to open log file with error: %v", err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	logger := hclog.New(&hclog.LoggerOptions{Name: "hvm", Level: hclog.LevelFromString("INFO"), Output: w})
	logger.Info("helper", "validateversion", m.BinaryName, "check version", m.BinaryCheckVersion)
	binaryVersions := []string{}
	var foundVersions bool
	resp, err := http.Get(fmt.Sprintf("%s/%s", ReleaseURLBase, m.BinaryName))
	if err != nil {
		logger.Error("helper", "failed to open validateversion url with error", err.Error())
		return validVersion, fmt.Errorf("failed to get url with error: %v", err)
	}
	defer resp.Body.Close()
	z := html.NewTokenizer(bufio.NewReader(resp.Body))
	for foundVersions == false {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			return false, nil
		case html.StartTagToken:
			t := z.Token()
			switch t.Data {
			case "a":
				z.Next()
				t = z.Token()
				version := strings.TrimPrefix(t.Data, fmt.Sprintf("%s_", binary))
				// strip "../" from inclusion into the slice
				if version == "../" {
					continue
				}
				binaryVersions = append(binaryVersions, version)
				if version == "0.1.0" {
					// we are at the bottom of the versions list now
					foundVersions = true
					break
				}
			}
		default:
			continue
		}
	}
	// we have relatively small slices, so...
	logger.Info("helper", "Versions", binaryVersions)
	for _, n := range binaryVersions {
		if binaryVersion == n {
			validVersion = true
			return validVersion, nil
		}
	}
	return validVersion, nil
}
