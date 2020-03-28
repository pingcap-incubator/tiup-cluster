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

package ansible

import (
	"os"
	"path/filepath"

	"github.com/creasty/defaults"
	"github.com/pingcap-incubator/tiops/pkg/meta"
	"github.com/relex/aini"
)

// ImportAnsible imports a TiDB cluster deployed by TiDB-Ansible
func ImportAnsible(dir string) error {
	inventoryFile, err := os.Open(filepath.Join(dir, ansibleInventoryFile))
	if err != nil {
		return err
	}
	defer inventoryFile.Close()

	inventory, err := aini.Parse(inventoryFile)
	if err != nil {
		return err
	}

	clsName, clsMeta, err := parseInventory(dir, inventory)
	if err != nil {
		return err
	}

	// TODO: add output of imported cluster name and version
	// TODO: check cluster name with other clusters managed by us for conflicts
	// TODO: prompt user for a chance to set a new cluster name

	// TODO: get values from templates of roles to overwrite defaults
	defaults.Set(clsMeta)
	return meta.SaveClusterMeta(clsName, clsMeta)
}
