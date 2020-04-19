// Copyright 2020 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package operator

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	//"github.com/pingcap-incubator/tiup-cluster/pkg/api"
	"github.com/pingcap-incubator/tiup-cluster/pkg/log"
	//"github.com/pingcap-incubator/tiup-cluster/pkg/meta"
	//"github.com/pingcap-incubator/tiup-cluster/pkg/utils"
	//"github.com/pingcap-incubator/tiup/pkg/set"
	//"github.com/pingcap/errors"
	"github.com/AstroProfundis/sysinfo"
	"github.com/pingcap/tidb-insight/collector/insight"
)

// CheckOptions control the list of checks to be performed
type CheckOptions struct {
	// checks that are disabled by default
	EnableCPU bool
	EnableMem bool
	//EnableDisk bool

	// pre-defined goups of checks
	//GroupMinimal bool // a minimal set of checks
}

// Names of checks
var (
	CheckTypeGeneral     = "general" // errors that don't fit any specific check
	CheckTypeNTP         = "ntp"
	CheckTypeOSVer       = "os-version"
	CheckTypeSwap        = "swap"
	CheckTypeSysctl      = "sysctl"
	CheckTypeCPUThreads  = "cpu-cores"
	CheckTypeCPUGovernor = "cpu-governor"
	CheckTypeMem         = "memory"
	CheckTypeLimits      = "limits"
	//CheckTypeFio    = "fio"
)

// CheckResult is the result of a check
type CheckResult struct {
	Name string // Name of the check
	Err  error  // An embedded error
	Warn bool   // The check didn't pass, but not a big problem
}

// Error implements the error interface
func (c CheckResult) Error() string {
	return fmt.Sprintf("check failed for %s: %s", c.Name, c.Err)
}

// Unwrap implements the Wrapper interface
func (c CheckResult) Unwrap() error {
	return c.Err
}

// IsWarning checks if the result is a warning error
func (c CheckResult) IsWarning() bool {
	return c.Warn
}

// Passed checks if the result is a success
func (c CheckResult) Passed() bool {
	return c.Err == nil
}

// CheckSystemInfo performs checks with basic system info
func CheckSystemInfo(opt *CheckOptions, rawData []byte) []CheckResult {
	results := make([]CheckResult, 0)
	var insightInfo insight.InsightInfo
	if err := json.Unmarshal(rawData, &insightInfo); err != nil {
		return append(results, CheckResult{
			Name: CheckTypeGeneral,
			Err:  err,
		})
	}

	// check basic system info
	results = append(results, checkSysInfo(opt, &insightInfo.SysInfo)...)

	// check NTP sync status
	results = append(results, checkNTP(&insightInfo.NTP))

	return results
}

func checkSysInfo(opt *CheckOptions, sysInfo *sysinfo.SysInfo) []CheckResult {
	results := make([]CheckResult, 0)

	results = append(results, checkOSInfo(opt, &sysInfo.OS))

	// check cpu core counts
	if opt.EnableCPU {
		results = append(results, checkCPU(&sysInfo.CPU)...)
	}

	// check memory size
	results = append(results, checkMem(opt, &sysInfo.Memory)...)

	return results
}

func checkOSInfo(opt *CheckOptions, osInfo *sysinfo.OS) CheckResult {
	result := CheckResult{
		Name: CheckTypeOSVer,
	}

	// check OS vendor
	switch osInfo.Vendor {
	case "centos", "redhat":
		// check version
		if ver, _ := strconv.Atoi(osInfo.Version); ver < 7 {
			result.Err = fmt.Errorf("%s %s not supported, use version 7 or higher",
				osInfo.Name, osInfo.Release)
			return result
		}
	case "debian", "ubuntu":
		// check version
	default:
		result.Err = fmt.Errorf("os vendor %s not supported", osInfo.Vendor)
		return result
	}

	// TODO: check OS architecture

	return result
}

func checkNTP(ntpInfo *insight.TimeStat) CheckResult {
	result := CheckResult{
		Name: CheckTypeNTP,
	}

	if ntpInfo.Status == "none" {
		log.Infof("The NTPd daemon may be not installed, skip.")
		return result
	}

	// check if time offset greater than +- 500ms
	if math.Abs(ntpInfo.Offset) >= 500 {
		result.Err = fmt.Errorf("time offet %fms too high", ntpInfo.Offset)
	}

	return result
}

func checkCPU(cpuInfo *sysinfo.CPU) []CheckResult {
	results := make([]CheckResult, 0)
	if cpuInfo.Threads < 16 {
		results = append(results, CheckResult{
			Name: CheckTypeCPUThreads,
			Err:  fmt.Errorf("CPU thread count %d too low, needs 16 or more", cpuInfo.Threads),
		})
	}

	// check for CPU frequency governor
	if cpuInfo.Governor != "" && cpuInfo.Governor != "performance" {
		results = append(results, CheckResult{
			Name: CheckTypeCPUGovernor,
			Err:  fmt.Errorf("CPU frequency governor is %s, should use performance", cpuInfo.Governor),
		})
	}

	return results
}

func checkMem(opt *CheckOptions, memInfo *sysinfo.Memory) []CheckResult {
	results := make([]CheckResult, 0)
	if memInfo.Swap > 0 {
		results = append(results, CheckResult{
			Name: CheckTypeSwap,
			Err:  fmt.Errorf("swap is enabled, please disable for best performance"),
		})
	}

	// 32GB
	if opt.EnableMem && memInfo.Size < 1024*32 {
		results = append(results, CheckResult{
			Name: CheckTypeMem,
			Err:  fmt.Errorf("memory size %dMB too low, needs 32GB or more", memInfo.Size),
		})
	}

	return results
}

// CheckSysLimits checks limits in /etc/security/limits.conf
func CheckSysLimits(opt *CheckOptions, user string, l []byte) []CheckResult {
	results := make([]CheckResult, 0)

	var (
		stackSoft  int
		nofileSoft int
		nofileHard int
	)

	for _, line := range strings.Split(string(l), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 3 || fields[0] != user {
			continue
		}

		switch fields[2] {
		case "nofile":
			if fields[1] == "soft" {
				nofileSoft, _ = strconv.Atoi(fields[3])
			} else {
				nofileHard, _ = strconv.Atoi(fields[3])
			}
		case "stack":
			if fields[1] == "soft" {
				stackSoft, _ = strconv.Atoi(fields[3])
			}
		}
	}

	if nofileSoft < 1000000 {
		results = append(results, CheckResult{
			Name: CheckTypeLimits,
			Err:  fmt.Errorf("soft limit of nofile for user %s is not set or too low", user),
		})
	}
	if nofileHard < 1000000 {
		results = append(results, CheckResult{
			Name: CheckTypeLimits,
			Err:  fmt.Errorf("hard limit of nofile for user %s is not set or too low", user),
		})
	}
	if stackSoft < 10240 {
		results = append(results, CheckResult{
			Name: CheckTypeLimits,
			Err:  fmt.Errorf("soft limit of stack for user %s is not set or too low", user),
		})
	}

	// all pass
	if len(results) < 1 {
		results = append(results, CheckResult{
			Name: CheckTypeLimits,
		})
	}

	return results
}

// CheckKernelParameters checks kernel parameter values
func CheckKernelParameters(opt *CheckOptions, p []byte) []CheckResult {
	results := make([]CheckResult, 0)

	for _, line := range strings.Split(string(p), "\n") {
		line = strings.TrimSpace(line)
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		switch fields[0] {
		case "fs.file-max":
			val, _ := strconv.Atoi(fields[2])
			if val < 1000000 {
				results = append(results, CheckResult{
					Name: CheckTypeSysctl,
					Err:  fmt.Errorf("fs.file-max = %d, should be greater than 1000000", val),
				})
			}
		case "net.core.somaxconn":
			val, _ := strconv.Atoi(fields[2])
			if val < 32768 {
				results = append(results, CheckResult{
					Name: CheckTypeSysctl,
					Err:  fmt.Errorf("net.core.somaxconn = %d, should be greater than 32768", val),
				})
			}
		case "net.ipv4.tcp_tw_recycle":
			val, _ := strconv.Atoi(fields[2])
			if val != 0 {
				results = append(results, CheckResult{
					Name: CheckTypeSysctl,
					Err:  fmt.Errorf("net.ipv4.tcp_tw_recycle = %d, should be 0", val),
				})
			}
		case "net.ipv4.tcp_syncookies":
			val, _ := strconv.Atoi(fields[2])
			if val != 0 {
				results = append(results, CheckResult{
					Name: CheckTypeSysctl,
					Err:  fmt.Errorf("net.ipv4.tcp_syncookies = %d, should be 0", val),
				})
			}
		case "vm.overcommit_memory":
			val, _ := strconv.Atoi(fields[2])
			if opt.EnableMem && val != 0 && val != 1 {
				results = append(results, CheckResult{
					Name: CheckTypeSysctl,
					Err:  fmt.Errorf("vm.overcommit_memory = %d, should be 0 or 1", val),
				})
			}
		case "vm.swappiness":
			val, _ := strconv.Atoi(fields[2])
			if val != 0 {
				results = append(results, CheckResult{
					Name: CheckTypeSysctl,
					Err:  fmt.Errorf("vm.swappiness = %d, should be 0", val),
				})
			}
		}
	}

	// all pass
	if len(results) < 1 {
		results = append(results, CheckResult{
			Name: CheckTypeSysctl,
		})
	}

	return results
}