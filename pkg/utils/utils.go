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

package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/joomcode/errorx"
)

var (
	errNS = errorx.NewNamespace("utils")
)

// JoinInt joins a slice of int to string
func JoinInt(nums []int, delim string) string {
	result := ""
	for _, i := range nums {
		result += strconv.Itoa(i)
		result += delim
	}
	return strings.TrimSuffix(result, delim)
}

// RetryOption is options for Retry()
type RetryOption struct {
	Attempts int
	Delay    time.Duration
	Timeout  time.Duration
}

// default values for RetryOption
var (
	defaultAttempts = 10
	defaultDelay    = time.Millisecond * 500 // 500ms
	defaultTimeout  = time.Second * 10       // 10s
)

// Retry retries the func until it returns no error or reaches attempts limit or
// timed out, either one is earlier
func Retry(doFunc func() error, opts ...RetryOption) error {
	var cfg RetryOption
	if len(opts) > 0 {
		cfg = opts[0]
	} else {
		cfg = RetryOption{
			Attempts: defaultAttempts,
			Delay:    defaultDelay,
			Timeout:  defaultTimeout,
		}
	}

	// attempts must be greater than 0
	if cfg.Attempts <= 0 {
		cfg.Attempts = defaultAttempts
	}

	timeoutChan := time.After(cfg.Timeout)

	// call the function
	var attemptCount int
	for attemptCount = 0; attemptCount < cfg.Attempts; attemptCount++ {
		if err := doFunc(); err == nil {
			return nil
		}

		// check for timeout
		select {
		case <-timeoutChan:
			return fmt.Errorf("operation timed out after %s", cfg.Timeout)
		default:
			time.Sleep(cfg.Delay)
		}
	}

	return fmt.Errorf("operation exceeds the max retry attempts of %d", cfg.Attempts)
}
