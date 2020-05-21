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
	"fmt"
	"strings"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/pingcap-incubator/tiup-cluster/pkg/utils"
	dmpb "github.com/pingcap/dm/dm/pb"
	"github.com/pingcap/errors"
	"go.uber.org/zap"
)

var (
	dmMembersURI = "apis/v1alpha1/members"
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

// GetURL builds the the client URL of DMClient
func (dm *DMMasterClient) GetURL(addr string) string {
	httpPrefix := "http"
	if dm.tlsEnabled {
		httpPrefix = "https"
	}
	return fmt.Sprintf("%s://%s", httpPrefix, addr)
}

func (dm *DMMasterClient) getEndpoints(cmd string) (endpoints []string) {
	for _, addr := range dm.addrs {
		endpoint := fmt.Sprintf("%s/%s", dm.GetURL(addr), cmd)
		endpoints = append(endpoints, endpoint)
	}

	return
}

func (dm *DMMasterClient) getMember(endpoints []string) (*dmpb.ListMemberResponse, error) {
	memberResp := &dmpb.ListMemberResponse{}
	err := tryURLs(endpoints, func(endpoint string) error {
		body, err := dm.httpClient.Get(endpoint)
		if err != nil {
			return err
		}

		err = jsonpb.Unmarshal(strings.NewReader(string(body)), memberResp)

		if err != nil {
			return err
		}

		if !memberResp.Result {
			return errors.New("dm-master get members failed: " + memberResp.Msg)
		}

		return nil
	})
	return memberResp, err
}

// GetMaster returns the dm master leader
// returns isFound, isActive, isLeader, error
func (dm *DMMasterClient) GetMaster(name string) (isFound bool, isActive bool, isLeader bool, err error) {
	query := "?leader=true&master=true&names=" + name
	endpoints := dm.getEndpoints(dmMembersURI + query)
	memberResp, err := dm.getMember(endpoints)

	if err != nil {
		zap.L().Error("get dm master status failed", zap.Error(err))
		return false, false, false, errors.AddStack(err)
	}

	for _, member := range memberResp.GetMembers() {
		if leader := member.GetLeader(); leader != nil {
			if leader.GetName() == name {
				isFound = true
				isLeader = true
			}
		} else if masters := member.GetMaster(); masters != nil {
			for _, master := range masters.GetMasters() {
				if master.GetName() == name {
					isFound = true
					isActive = master.GetAlive()
				}
			}
		}
	}
	return
}

// GetMaster returns the dm master leader
// returns (worker stage, error). If worker stage is "", that means this worker is in cluster
func (dm *DMMasterClient) GetWorker(name string) (string, error) {
	query := "?worker=true&names=" + name
	endpoints := dm.getEndpoints(dmMembersURI + query)
	memberResp, err := dm.getMember(endpoints)

	if err != nil {
		zap.L().Error("get dm worker status failed", zap.Error(err))
		return "", err
	}

	stage := ""
	for _, member := range memberResp.Members {
		if workers := member.GetWorker(); workers != nil {
			for _, worker := range workers.GetWorkers() {
				if worker.GetName() == name {
					stage = worker.GetStage()
				}
			}
		}
	}
	if len(stage) > 0 {
		stage = strings.ToUpper(stage[0:1]) + stage[1:]
	}

	return stage, nil
}

// GetLeader gets leader of dm cluster
func (dm *DMMasterClient) GetLeader() (string, error) {
	query := "?leader=true"
	endpoints := dm.getEndpoints(dmMembersURI + query)
	memberResp, err := dm.getMember(endpoints)

	if err != nil {
		return "", errors.AddStack(err)
	}

	leaderName := ""
	for _, member := range memberResp.Members {
		if leader := member.GetLeader(); leader != nil {
			leaderName = leader.GetName()
		}
	}
	return leaderName, nil
}

// EvictDMMasterLeader evicts the dm master leader
func (dm *DMMasterClient) EvictDMMasterLeader(retryOpt *utils.RetryOption) error {
	return nil
}

func (dm *DMMasterClient) GetRegisteredMastersWorkers() ([]string, []string, error) {
	return []string{}, []string{}, nil
}

// OfflineWorker offlines the dm worker
func (dm *DMMasterClient) OfflineWorker(name string) error {
	return nil
}

// OfflineMaster offlines the dm master
func (dm *DMMasterClient) OfflineMaster(name string) error {
	return nil
}
