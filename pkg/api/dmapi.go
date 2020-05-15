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

package api

import (
	"crypto/tls"
	"time"

	"github.com/pingcap-incubator/tiup-cluster/pkg/utils"
)

// DMMasterClient is an HTTP client of the dm-master server
type DMMasterClient struct {
	addrs      []string
	tlsEnabled bool
	httpClient *utils.HTTPClient
}

// NewDMMasterClient returns a new PDClient
func NewDMMasterClient(addrs []string, timeout time.Duration, tlsConfig *tls.Config) *DMMasterClient {
	enableTLS := false
	if tlsConfig != nil {
		enableTLS = true
	}

	return &DMMasterClient{
		addrs:      addrs,
		tlsEnabled: enableTLS,
		httpClient: utils.NewHTTPClient(timeout, tlsConfig),
	}
}

// EvictDMMasterLeader evicts the dm master leader
func (dm *DMMasterClient) GetLeader() (string, error) {
	return "", nil
}

// EvictDMMasterLeader evicts the dm master leader
func (dm *DMMasterClient) EvictDMMasterLeader(retryOpt *utils.RetryOption) error {
	return nil
}
