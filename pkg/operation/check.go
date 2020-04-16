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
	// checks that are enabled by default, use flag to disable one
	//DisableSysTime bool
	DisableNTP       bool
	DisableOSVersion bool
	DisableSwap      bool
	DisableLimits    bool
	DisableSysctl    bool

	// checks that are disabled by default
	EnableCPU bool
	EnableMem bool
	//EnableDisk bool

	// pre-defined goups of checks
	//GroupMinimal bool // a minimal set of checks
}

// CheckSystemInfo performs checks with basic system info
func CheckSystemInfo(opt *CheckOptions, rawData []byte) error {
	var insightInfo insight.InsightInfo
	if err := json.Unmarshal(rawData, &insightInfo); err != nil {
		return err
	}

	// check basic system info
	if err := checkSysInfo(opt, &insightInfo.SysInfo); err != nil {
		return err
	}

	// check NTP sync status
	if !opt.DisableNTP {
		err := checkNTP(&insightInfo.NTP)
		if err != nil {
			return err
		}
	}

	return nil
}

func checkSysInfo(opt *CheckOptions, sysInfo *sysinfo.SysInfo) error {
	if err := checkNodeInfo(opt, &sysInfo.Node); err != nil {
		return err
	}

	if !opt.DisableOSVersion {
		if err := checkOSInfo(opt, &sysInfo.OS); err != nil {
			return err
		}
	}

	// check cpu core counts
	if opt.EnableCPU {
		err := checkCPU(&sysInfo.CPU)
		if err != nil {
			return err
		}
	}

	// check memory size
	err := checkMem(opt, &sysInfo.Memory)
	if err != nil {
		return err
	}

	return nil
}

func checkNodeInfo(opt *CheckOptions, nodeInfo *sysinfo.Node) error {
	return nil
}

func checkOSInfo(opt *CheckOptions, osInfo *sysinfo.OS) error {
	// check OS vendor
	switch osInfo.Vendor {
	case "centos", "redhat":
		// check version
		if ver, _ := strconv.Atoi(osInfo.Version); ver < 7 {
			return fmt.Errorf("%s %s not supported, use version 7 or higher",
				osInfo.Name, osInfo.Release)
		}
	case "debian", "ubuntu":
		// check version
	default:
		return fmt.Errorf("os vendor %s not supported", osInfo.Vendor)
	}

	// TODO: check OS architecture

	return nil
}

func checkNTP(ntpInfo *insight.TimeStat) error {
	if ntpInfo.Status == "none" {
		log.Infof("The NTPd daemon may be not installed, skip.")
		return nil
	}

	// check if time offset greater than +- 500ms
	if math.Abs(ntpInfo.Offset) >= 500 {
		return fmt.Errorf("time offet %fms too high", ntpInfo.Offset)
	}

	return nil
}

func checkCPU(cpuInfo *sysinfo.CPU) error {
	if cpuInfo.Threads < 16 {
		return fmt.Errorf("CPU thread count %d too low, needs 16 or more", cpuInfo.Threads)
	}

	// check for CPU frequency governor
	if cpuInfo.Governor != "" && cpuInfo.Governor != "performance" {
		return fmt.Errorf("CPU frequency governor is %s, should use performance", cpuInfo.Governor)
	}

	return nil
}

func checkMem(opt *CheckOptions, memInfo *sysinfo.Memory) error {
	if !opt.DisableSwap && memInfo.Swap > 0 {
		return fmt.Errorf("swap is enabled, please disable for best performance")
	}

	// 32GB
	if opt.EnableMem && memInfo.Size < 1024*32 {
		return fmt.Errorf("memory size %dMB too low, needs 32GB or more", memInfo.Size)
	}

	return nil
}

// CheckSysLimits checks limits in /etc/security/limits.conf
func CheckSysLimits(opt *CheckOptions, user string, l []byte) error {
	if opt.DisableLimits {
		return nil
	}

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
		return fmt.Errorf("soft limit of nofile for user %s is not set or too low", user)
	}
	if nofileHard < 1000000 {
		return fmt.Errorf("hard limit of nofile for user %s is not set or too low", user)
	}
	if stackSoft < 10240 {
		return fmt.Errorf("soft limit of stack for user %s is not set or too low", user)
	}

	return nil
}

// CheckKernelParameters checks kernel parameter values
func CheckKernelParameters(opt *CheckOptions, p []byte) error {
	if opt.DisableSysctl {
		return nil
	}

	for _, line := range strings.Split(string(p), "\n") {
		line = strings.TrimSpace(line)
		fields := strings.Fields(line)

		switch fields[0] {
		case "fs.file-max":
			val, _ := strconv.Atoi(fields[2])
			if !opt.DisableLimits && val < 1000000 {
				return fmt.Errorf("fs.file-max = %d, should be greater than 1000000", val)
			}
		case "net.core.somaxconn":
			val, _ := strconv.Atoi(fields[2])
			if !opt.DisableLimits && val < 32768 {
				return fmt.Errorf("net.core.somaxconn = %d, should be greater than 32768", val)
			}
		case "net.ipv4.tcp_tw_recycle":
			val, _ := strconv.Atoi(fields[2])
			if val != 0 {
				return fmt.Errorf("net.ipv4.tcp_tw_recycle = %d, should be 0", val)
			}
		case "net.ipv4.tcp_syncookies":
			val, _ := strconv.Atoi(fields[2])
			if val != 0 {
				return fmt.Errorf("net.ipv4.tcp_syncookies = %d, should be 0", val)
			}
		case "vm.overcommit_memory":
			val, _ := strconv.Atoi(fields[2])
			if opt.EnableMem && val != 0 && val != 1 {
				return fmt.Errorf("vm.overcommit_memory = %d, should be 0 or 1", val)
			}
		case "vm.swappiness":
			val, _ := strconv.Atoi(fields[2])
			if !opt.DisableSwap && val != 0 {
				return fmt.Errorf("vm.swappiness = %d, should be 0", val)
			}
		}
	}
	return nil
}
