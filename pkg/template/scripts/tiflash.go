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

package scripts

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"text/template"

	"github.com/pingcap-incubator/tiup/pkg/localdata"
)

// TiFlashScript represent the data to generate TiFlash config
type TiFlashScript struct {
	IP         string
	Port       uint64
	StatusPort uint64
	DeployDir  string
	DataDir    string
	NumaNode   string
	Endpoints  []*PDScript
}

// NewTiFlashScript returns a TiFlashScript with given arguments
func NewTiFlashScript(ip, deployDir, dataDir string) *TiFlashScript {
	return &TiFlashScript{
		IP:         ip,
		Port:       20160,
		StatusPort: 20180,
		DeployDir:  deployDir,
		DataDir:    dataDir,
	}
}

// WithPort set Port field of TiFlashScript
func (c *TiFlashScript) WithPort(port uint64) *TiFlashScript {
	c.Port = port
	return c
}

// WithStatusPort set StatusPort field of TiFlashScript
func (c *TiFlashScript) WithStatusPort(port uint64) *TiFlashScript {
	c.StatusPort = port
	return c
}

// WithNumaNode set NumaNode field of TiFlashScript
func (c *TiFlashScript) WithNumaNode(numa string) *TiFlashScript {
	c.NumaNode = numa
	return c
}

// AppendEndpoints add new PDScript to Endpoints field
func (c *TiFlashScript) AppendEndpoints(ends ...*PDScript) *TiFlashScript {
	c.Endpoints = append(c.Endpoints, ends...)
	return c
}

// Config read ${localdata.EnvNameComponentInstallDir}/templates/scripts/run_TiFlash.sh.tpl as template
// and generate the config by ConfigWithTemplate
func (c *TiFlashScript) Config() ([]byte, error) {
	fp := path.Join(os.Getenv(localdata.EnvNameComponentInstallDir), "templates", "scripts", "run_TiFlash.sh.tpl")
	tpl, err := ioutil.ReadFile(fp)
	if err != nil {
		return nil, err
	}
	return c.ConfigWithTemplate(string(tpl))
}

// ConfigToFile write config content to specific path
func (c *TiFlashScript) ConfigToFile(file string) error {
	config, err := c.Config()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(file, config, 0755)
}

// ConfigWithTemplate generate the TiFlash config content by tpl
func (c *TiFlashScript) ConfigWithTemplate(tpl string) ([]byte, error) {
	tmpl, err := template.New("TiFlash").Parse(tpl)
	if err != nil {
		return nil, err
	}

	content := bytes.NewBufferString("")
	if err := tmpl.Execute(content, c); err != nil {
		return nil, err
	}

	return content.Bytes(), nil
}
