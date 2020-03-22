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

// AlertManagerScript represent the data to generate AlertManager start script
type AlertManagerScript struct {
	WebPort     uint64
	ClusterPort uint64
	DeployDir   string
	DataDir     string
	NumaNode    string
}

// NewAlertManagerScript returns a AlertManagerScript with given arguments
func NewAlertManagerScript(deployDir, dataDir string) *AlertManagerScript {
	return &AlertManagerScript{
		WebPort:     8888,
		ClusterPort: 9999,
		DeployDir:   deployDir,
		DataDir:     dataDir,
	}
}

// WithWebPort set WebPort field of AlertManagerScript
func (c *AlertManagerScript) WithWebPort(port uint64) *AlertManagerScript {
	c.WebPort = port
	return c
}

// WithClusterPort set WebPort field of AlertManagerScript
func (c *AlertManagerScript) WithClusterPort(port uint64) *AlertManagerScript {
	c.ClusterPort = port
	return c
}

// WithNumaNode set NumaNode field of AlertManagerScript
func (c *AlertManagerScript) WithNumaNode(numa string) *AlertManagerScript {
	c.NumaNode = numa
	return c
}

// Config read ${localdata.EnvNameComponentInstallDir}/templates/scripts/run_alertmanager.sh.tpl as template
// and generate the config by ConfigWithTemplate
func (c *AlertManagerScript) Config() (string, error) {
	fp := path.Join(os.Getenv(localdata.EnvNameComponentInstallDir), "templates", "scripts", "run_alertmanager.sh.tpl")
	tpl, err := ioutil.ReadFile(fp)
	if err != nil {
		return "", err
	}
	return c.ConfigWithTemplate(string(tpl))
}

// ConfigWithTemplate generate the AlertManager config content by tpl
func (c *AlertManagerScript) ConfigWithTemplate(tpl string) (string, error) {
	tmpl, err := template.New("AlertManager").Parse(tpl)
	if err != nil {
		return "", err
	}

	content := bytes.NewBufferString("")
	if err := tmpl.Execute(content, c); err != nil {
		return "", err
	}

	return content.String(), nil
}
