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
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"io/ioutil"
	"net/http"
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

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install (<binary>) [<version>]",
	Short: "Install a supported binary at the latest available or specified version",
	Long: `
Install a supported binary binary at specified version for the host architecture
and operating system; if version is omitted, the latest available version will be
installed.

hvm can install the following utilities:

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
  hvm install vault

  hvm install consul 1.4.2`,
	// Using a custom Args function here as workaround for GH-745
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args)  < 1 {
		  return errors.New("install requires exactly 1 argument")
		}
        if args[1] != "" {
        	fmt.Println("DEBUG: Got binary:", args[0], "version:", args[1])
        	binary := args[0]
        	version := args[1]
			v, err := ValidateVersion(binary, version)
            fmt.Println("DEBUG got valid:", v)
        	if err != nil {
        		fmt.Println("cannot validate binary or version")
        		os.Exit(1)
        	}
        	if v == false {
				// return fmt.Errorf("%s is not a version of %s that can be installed", args[1], args[0])
				fmt.Println(fmt.Printf("DEBUG: CLAIMS: %s is not a version of %s that can be installed", args[1], args[0]))
        	}
        }
		return cobra.OnlyValidArgs(cmd, args)
	},
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
		// m.BinaryDesiredVersion = binaryVersion
		if args[1] == "" {
            m.BinaryDesiredVersion = "latest"
			} else {
				m.BinaryDesiredVersion = args[1]
		    }
		m.BinaryOS = runtime.GOOS
		// m.BinaryName = strings.Join(args, " ")
		m.BinaryName = args[0]
		if _, err := os.Stat(m.HvmHome); os.IsNotExist(err) {
			os.Mkdir(m.HvmHome, 0755)
		}
		f, err := os.OpenFile(m.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println(fmt.Sprintf("failed to open log file %s with error: %v", m.LogFile, err))
			os.Exit(1)
		}
		defer f.Close()
		w := bufio.NewWriter(f)
		logger := hclog.New(&hclog.LoggerOptions{Name: "hvm", Level: hclog.LevelFromString("INFO"), Output: w})

		// Validate binary attributes with helper functions
		// Is desired binary version valid?
		// XXX: TODO: finish helper for valid version

		// Is desired binary already installed?
		var installedVersion bool
		installedVersion, err = IsInstalledVersion(m.BinaryName, m.BinaryDesiredVersion)
		if err != nil {
			fmt.Println(fmt.Sprintf("cannot use binary: %s with error: %v", m.BinaryName, err))
			os.Exit(1)
		}
		if installedVersion == true {
			if m.BinaryDesiredVersion == "" {
				fmt.Println(fmt.Sprintf("latest %s version installed", m.BinaryName))
				os.Exit(1)
			} else {
				fmt.Println(fmt.Sprintf("%s version %s appears to be already installed", m.BinaryName, m.BinaryDesiredVersion))
				os.Exit(1)
			}
		} else {
			logger.Info("install", "run", m.BinaryName, "desired version", m.BinaryDesiredVersion)
			err = installBinary(&m)
			if err != nil {
				fmt.Println(fmt.Sprintf("cannot install %s version %s with error: %v", m.BinaryName, m.BinaryDesiredVersion, err))
				os.Exit(1)
			}
		}

	},
}

// Initialize the command
func init() {
	rootCmd.AddCommand(installCmd)
	// installCmd.PersistentFlags().StringVar(&binaryVersion,
	//	"version",
	//	"",
	//	"install binary version")
}

// fetch HTML data over HTTP
func fetchData(m *InstallMeta, URL string) ([]byte, error) {
	f, err := os.OpenFile(m.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file %s with error: %v", m.LogFile, err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	logger := hclog.New(&hclog.LoggerOptions{Name: "hvm", Level: hclog.LevelFromString("INFO"), Output: w})
	response, err := http.Get(URL)
	if err != nil {
		logger.Error("install", "fetch data error", err.Error())
		return nil, fmt.Errorf("failed to fetch data with error: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		err = errors.New(response.Status)
		logger.Error("install", "fetch data error", err.Error())
		return nil, fmt.Errorf("failed to fetch data with error: %v", err)
	}
	var fetchData bytes.Buffer
	_, err = io.Copy(&fetchData, response.Body)
	if err != nil {
		logger.Error("install", "fetch data error", err.Error())
		return nil, fmt.Errorf("failed to fetch data with bytes buffer error: %v", err)
	}
	return fetchData.Bytes(), nil
}

// getLatestVersion also tooooo complicated now
// need to refactor stuff into helpers
func getLatestVersion(binary string, m *InstallMeta) (string, error) {
	f, err := os.OpenFile(m.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to open log file %s with error: %v", m.LogFile, err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	logger := hclog.New(&hclog.LoggerOptions{Name: "hvm", Level: hclog.LevelFromString("INFO"), Output: w})
	logger.Debug("install", "f-get-latest-version-switch-pre", m.BinaryName)
	switch binary {
	// Some binary latest versions cannot be queried through the Checkpoint API.
	// Those must unfortunately use an HTML scraping approach instead.
	case Vault:
		logger.Debug("install", "f-get-latest-version-html-scrape-url-base", VaultReleaseURLBase)
		logger.Debug("install", "f-get-latest-version-html-scrape-binary-name", binary)
		var found bool
		resp, err := http.Get(VaultReleaseURLBase)
		if err != nil {
			return "", err
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
		logger.Debug("install", "f-get-latest-version-checkpoint-url-base", CheckpointURLBase)
		logger.Debug("install", "f-get-latest-version-checkpoint-binary-name", binary)

		checkpointDataURL := fmt.Sprintf("%s/v1/check/%s", CheckpointURLBase, binary)
		logger.Debug("install", "f-get-latest-version-checkpoint-data-url", checkpointDataURL)

		checkPointClient := http.Client{Timeout: time.Second * 2}
		req, err := http.NewRequest(http.MethodGet, checkpointDataURL, nil)
		if err != nil {
			logger.Error("install", "f-get-latest-version", "request-error", err.Error())
			return "", err
		}

		req.Header.Set("User-Agent", "hvm-oss-http-client")
		res, err := checkPointClient.Do(req)
		if err != nil {
			logger.Error("install", "f-get-latest-version", "get-error", err.Error())
			return "", err
		}

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			logger.Error("install", "f-get-latest-version", "read-body-error", err.Error())
			return "", err
		}

		err = json.Unmarshal(body, &m)
		if err != nil {
			logger.Error("install", "f-get-latest-version", "json-unmarshall-error", err.Error())
			return "", fmt.Errorf("failed to unmarshal JSON with error: %v", err)
		}
		// Ensure that we get something like a valid version back from the API
		// and not for example, a maintenance page or similar...
		checkpointLatestVersion, err := version.NewVersion(m.BinaryLatestVersion)
		if err != nil {
			logger.Error("install", "issue", "Could not determine comparison version!", "error", err.Error())
			return "", err
		}
		constraints, err := version.NewConstraint(">= 0.0.1")
		if err != nil {
			logger.Error("install", "f-get-latest-version", "issue", "Could not determine comparison constraints!", "error", err.Error())
			return "", err
		}
		if constraints.Check(checkpointLatestVersion) {
			logger.Debug("install", "f-get-latest-version", "chcked-version", "version", checkpointLatestVersion, "constraints", constraints)
		} else {
			// Eh oh, something is wrong!
			logger.Error("install", "f-get-latest-version", "issue", "unexpected-checkpoint-api-value", m.BinaryLatestVersion)
			return "", fmt.Errorf("problem determining latest binary version")
		}
		return m.BinaryLatestVersion, nil
	default:
		if m.BinaryName != Vault {
			logger.Warn("install", "binary", m.BinaryName, "unsupported-binary", "Binary not in CheckPoint API or otherwise not supported.")
			return "", fmt.Errorf("Binary currently unsupported")
		}
	}
	return m.BinaryLatestVersion, nil
}

// installBinary has entirely too much going on in it right now!
// some of this needs to go into the helpers
func installBinary(m *InstallMeta) error {
	f, err := os.OpenFile(m.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file %s with error: %v", m.LogFile, err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	logger := hclog.New(&hclog.LoggerOptions{Name: "hvm", Level: hclog.LevelFromString("INFO"), Output: w})
	logger.Debug("install", "f-install-binary", "start", "with-binary", m.BinaryName)
	if m.BinaryName == "" {
		m.BinaryName = "none"
		logger.Error("install", "unknown-binary-candidate", "GURU DEDICATION")
		return fmt.Errorf("install: unknown binary candidate. GURU DEDICATION")
	}
	if m.BinaryDesiredVersion == "latest" {
		logger.Debug("install", "f-install-binary", "blank-version", "binary", m.BinaryName)
		latestBinaryVersion, err := getLatestVersion(m.BinaryName, m)
		if err != nil {
			logger.Error("install", "get-latest-version-fail", "error", err.Error())
			return err
		}
		logger.Debug("install", "get-latest-version", "inner", "got-version", latestBinaryVersion)
		m.BinaryDesiredVersion = latestBinaryVersion
	}
	logger.Info("install", "install binary candidate", "final", "binary", m.BinaryName, "desired-version", m.BinaryDesiredVersion)

	switch m.BinaryName {
	case Consul, Nomad, Packer, Terraform, Vagrant, Vault:
		targetPath := fmt.Sprintf("%s/.hvm/%s/%s", m.UserHome, m.BinaryName, m.BinaryDesiredVersion)
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
		binaryShaURL := fmt.Sprintf("%s/%s/%s/%s_%s_SHA256SUMS", ReleaseURLBase, m.BinaryName, m.BinaryDesiredVersion, m.BinaryName, m.BinaryDesiredVersion)
		logger.Debug("install", "sha256sums-file-url", binaryShaURL)
		binarySha, err := fetchData(m, binaryShaURL)
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
				if m.BinaryName == Nomad {
					logger.Debug("install", "stage", "scanner", "binary", Nomad)
					nomadLatestVersion, err := version.NewVersion(m.BinaryDesiredVersion)
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
					logger.Debug("install", "binary", m.BinaryName)
				}
			}
			fileSha[s[1]] = s[0]
		}
		if err := scanner.Err(); err != nil {
			logger.Error("install", "process-sha256sums-error", err.Error())
			return err
		}
		pkgFilename := fmt.Sprintf("%s_%s_%s_%s.zip",
			m.BinaryName,
			m.BinaryDesiredVersion,
			m.BinaryOS,
			m.BinaryArch)
		checkSha := fileSha[pkgFilename]
		fullURL := fmt.Sprintf("%s/%s/%s/%s?checksum=sha256:%s",
			ReleaseURLBase,
			m.BinaryName,
			m.BinaryDesiredVersion,
			pkgFilename,
			checkSha)
		installPath := fmt.Sprintf("%s/%s", targetPath, m.BinaryName)
		logger.Debug("install", "valid-binary", "true", "full-url", fullURL, "install-path", installPath)
		// Get binary archive using go-getter from a URL which takes the form of:
		// 'https://releases.hashicorp.com/<binary>/<version>/<binary>_<version>_<os>_<arch>.zip
		// go-getter validates the intended download against its published SHA256 summary before downloading, or fails if the there is mismatch / other issue which prevents comparison.
		// Shout out to Ye Olde School BSD spinner!
		hvmSpinnerSet := []string{"/", "|", "\\", "-", "|", "\\", "-"}
		s := spinner.New(hvmSpinnerSet, 174*time.Millisecond)
		s.Writer = os.Stderr
		s.Color("fgHiCyan")
		s.Suffix = " Installing..."
		s.FinalMSG = fmt.Sprintf("Installed %s (%s/%s) version %s\n", m.BinaryName, m.BinaryOS, m.BinaryArch, m.BinaryDesiredVersion)
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
		logger.Warn("install", "binary", m.BinaryName, "unsupported-binary", "not in CheckPoint API")
		return errors.New("Binary currently unsupported")
	}
}
