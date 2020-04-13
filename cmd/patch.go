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

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/joomcode/errorx"
	"github.com/pingcap-incubator/tiup-cluster/pkg/bindversion"
	"github.com/pingcap-incubator/tiup-cluster/pkg/clusterutil"
	"github.com/pingcap-incubator/tiup-cluster/pkg/meta"
	operator "github.com/pingcap-incubator/tiup-cluster/pkg/operation"
	"github.com/pingcap-incubator/tiup-cluster/pkg/task"
	"github.com/pingcap-incubator/tiup-cluster/pkg/utils"
	tiupmeta "github.com/pingcap-incubator/tiup/pkg/meta"
	"github.com/pingcap-incubator/tiup/pkg/repository"
	"github.com/pingcap-incubator/tiup/pkg/set"
	tiuputils "github.com/pingcap-incubator/tiup/pkg/utils"
	"github.com/pingcap/errors"
	"github.com/spf13/cobra"
)

func newPatchCmd() *cobra.Command {
	var (
		overwrite bool
		options   operator.Options
	)
	cmd := &cobra.Command{
		Use:   "patch <cluster-name> <package-path>",
		Short: "Replace remote package with local one",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return cmd.Help()
			}
			if len(options.Nodes) == 0 && len(options.Roles) == 0 {
				return errors.New("the flag -R or -N must be specified at least one")
			}
			return patch(args[0], args[1], options, overwrite)
		},
	}

	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "Use this package in the futrue scale-out operations")
	cmd.Flags().StringSliceVarP(&options.Nodes, "node", "N", nil, "Specify the nodes")
	cmd.Flags().StringSliceVarP(&options.Roles, "role", "R", nil, "Specify the role")

	return cmd
}

func patch(clusterName, packagePath string, options operator.Options, overwrite bool) error {
	if tiuputils.IsNotExist(meta.ClusterPath(clusterName, meta.MetaFileName)) {
		return errors.Errorf("cannot patch non-exists cluster %s", clusterName)
	}

	if exist, err := utils.FileExist(packagePath); err != nil {
		return err
	} else if !exist {
		return errors.New("specified package not exists")
	}

	metadata, err := meta.ClusterMetadata(clusterName)
	if err != nil {
		return err
	}

	insts, err := instancesToPatch(metadata, options)
	if err != nil {
		return err
	}
	if err := checkPackage(clusterName, insts[0].Role(), packagePath); err != nil {
		return err
	}

	var replacePackageTasks []task.Task
	for _, inst := range insts {
		deployDir := clusterutil.Abs(metadata.User, inst.DeployDir())
		tb := task.NewBuilder()
		tb.BackupComponent(inst.ComponentName(), metadata.Version, inst.GetHost(), deployDir).
			InstallPackage(packagePath, inst.GetHost(), deployDir)
		replacePackageTasks = append(replacePackageTasks, tb.Build())
	}

	t := task.NewBuilder().
		SSHKeySet(
			meta.ClusterPath(clusterName, "ssh", "id_rsa"),
			meta.ClusterPath(clusterName, "ssh", "id_rsa.pub")).
		ClusterSSH(metadata.Topology, metadata.User, sshTimeout).
		Parallel(replacePackageTasks...).
		ClusterOperate(metadata.Topology, operator.UpgradeOperation, options).
		Build()

	if err := t.Execute(task.NewContext()); err != nil {
		if errorx.Cast(err) != nil {
			// FIXME: Map possible task errors and give suggestions.
			return err
		}
		return errors.Trace(err)
	}

	if overwrite {
		if err := overwritePatch(clusterName, insts[0].Role(), packagePath); err != nil {
			return err
		}
	}

	return nil
}

func instancesToPatch(metadata *meta.ClusterMeta, options operator.Options) ([]meta.Instance, error) {
	roleFilter := set.NewStringSet(options.Roles...)
	nodeFilter := set.NewStringSet(options.Nodes...)
	components := metadata.Topology.ComponentsByStartOrder()
	components = operator.FilterComponent(components, roleFilter)

	instances := []meta.Instance{}
	for _, com := range components {
		instances = append(instances, operator.FilterInstance(com.Instances(), nodeFilter)...)
	}

	if len(instances) == 0 {
		return nil, errors.New("no instance found")
	}

	return instances, nil
}

func checkPackage(clusterName, role, packagePath string) error {
	metadata, err := meta.ClusterMetadata(clusterName)
	if err != nil {
		return err
	}
	manifest, err := tiupmeta.Repository().ComponentVersions(role)
	if err != nil {
		return err
	}
	var versionInfo *repository.VersionInfo
	ver := bindversion.ComponentVersion(role, metadata.Version)
	for _, vi := range manifest.Versions {
		if vi.Version == ver {
			versionInfo = new(repository.VersionInfo)
			*versionInfo = vi
		}
	}
	if versionInfo == nil {
		return fmt.Errorf("cannot found version %v in %s manifest", ver, role)
	}

	checksum, err := utils.Checksum(packagePath)
	if err != nil {
		return err
	}
	cacheDir := meta.ClusterPath(clusterName, "cache", role+"-"+checksum[:7])
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return err
	}
	if err := exec.Command("tar", "-xvf", packagePath, "-C", cacheDir).Run(); err != nil {
		return err
	}

	if exists, err := utils.FileExist(path.Join(cacheDir, versionInfo.Entry)); err != nil {
		return err
	} else if !exists {
		return fmt.Errorf("entry %s not found in package %s", versionInfo.Entry, packagePath)
	}

	return nil
}

func overwritePatch(clusterName, role, packagePath string) error {
	if err := os.MkdirAll(meta.ClusterPath(clusterName, meta.PatchDirName), 0755); err != nil {
		return err
	}
	checksum, err := utils.Checksum(packagePath)
	if err != nil {
		return err
	}
	tg := meta.ClusterPath(clusterName, meta.PatchDirName, role+"-"+checksum[:7]+".tar.gz")
	if err := utils.CopyFile(packagePath, tg); err != nil {
		return err
	}
	if err := os.Symlink(tg, meta.ClusterPath(clusterName, meta.PatchDirName, role+".tar.gz")); err != nil {
		return err
	}
	return nil
}
