package meta

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/pingcap-incubator/tiops/pkg/executor"
	"github.com/pingcap-incubator/tiops/pkg/template/scripts"
)

// DrainerComponent represents Drainer component.
type DrainerComponent struct{ *Specification }

// Name implements Component interface.
func (c *DrainerComponent) Name() string {
	return ComponentDrainer
}

// Instances implements Component interface.
func (c *DrainerComponent) Instances() []Instance {
	ins := make([]Instance, 0, len(c.Drainers))
	for _, s := range c.Drainers {
		ins = append(ins, &DrainerInstance{instance{
			InstanceSpec: s,
			name:         c.Name(),
			host:         s.Host,
			port:         s.Port,
			sshp:         s.SSHPort,
			topo:         c.Specification,

			usedPorts: []int{
				s.Port,
			},
			usedDirs: []string{
				s.DeployDir,
				s.DataDir,
			},
			statusFn: func(_ ...string) string {
				url := fmt.Sprintf("http://%s:%d/status", s.Host, s.Port)
				return statusByURL(url)
			},
		}})
	}
	return ins
}

// DrainerInstance represent the Drainer instance.
type DrainerInstance struct {
	instance
}

// ScaleConfig deploy temporary config on scaling
func (i *DrainerInstance) ScaleConfig(e executor.TiOpsExecutor, b *Specification, user, cacheDir, deployDir string) error {
	s := i.instance.topo
	defer func() {
		i.instance.topo = s
	}()
	i.instance.topo = b

	return i.InitConfig(e, user, cacheDir, deployDir)
}

// InitConfig implements Instance interface.
func (i *DrainerInstance) InitConfig(e executor.TiOpsExecutor, user, cacheDir, deployDir string) error {
	if err := i.instance.InitConfig(e, user, cacheDir, deployDir); err != nil {
		return err
	}

	// transfer run script
	ends := []*scripts.PDScript{}
	for _, spec := range i.instance.topo.PDServers {
		ends = append(ends, scripts.NewPDScript(spec.Name, spec.Host, spec.DeployDir, spec.DataDir))
	}

	cfg := scripts.NewDrainerScript(
		i.GetHost()+":"+strconv.Itoa(i.GetPort()),
		i.GetHost(),
		deployDir,
		filepath.Join(deployDir, "data"),
	).AppendEndpoints(ends...)

	spec := i.InstanceSpec.(DrainerSpec)
	cfg.WithCommitTs(spec.CommitTS)

	fp := filepath.Join(cacheDir, fmt.Sprintf("run_drainer_%s_%d.sh", i.GetHost(), i.GetPort()))

	if err := cfg.ConfigToFile(fp); err != nil {
		return err
	}
	dst := filepath.Join(deployDir, "scripts", "run_drainer.sh")
	if err := e.Transfer(fp, dst, false); err != nil {
		return err
	}

	if _, _, err := e.Execute("chmod +x "+dst, false); err != nil {
		return err
	}

	return i.mergeServerConfig(e, i.topo.ServerConfigs.Drainer, spec.Config, cacheDir, deployDir)
}
