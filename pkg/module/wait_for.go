package module

import (
	"bytes"
	"fmt"
	"time"

	"github.com/pingcap-incubator/tiops/pkg/executor"
	"github.com/pingcap-incubator/tiops/pkg/utils"
	"github.com/pingcap/errors"
)

// WaitForConfig is the configurations of WaitFor module.
type WaitForConfig struct {
	Port  int           // Port number to poll.
	Sleep time.Duration // Duration to sleep between checks, default 1 second.
	// Choices:
	// started
	// stopped
	// When checking a port started will ensure the port is open, stopped will check that it is closed
	State   string
	Timeout time.Duration // Maximum duration to wait for.
}

// WaitFor is the module used to wait for some condition.
type WaitFor struct {
	c WaitForConfig
}

// NewWaitFor create a WaitFor instance.
func NewWaitFor(c WaitForConfig) *WaitFor {
	if c.Sleep == 0 {
		c.Sleep = time.Second
	}
	if c.Timeout == 0 {
		c.Timeout = time.Second * 60
	}
	if c.State == "" {
		c.State = "started"
	}

	w := &WaitFor{
		c: c,
	}

	return w
}

// Execute the module return nil if successfully wait for the event.
func (w *WaitFor) Execute(e executor.TiOpsExecutor) (err error) {
	pattern := []byte(fmt.Sprintf(":%d ", w.c.Port))

	retryOpt := utils.RetryOption{
		Attempts: 60,
		Delay:    w.c.Sleep,
		Timeout:  w.c.Timeout,
	}
	if err := utils.Retry(func() error {
		// only listing TCP ports
		stdout, _, err := e.Execute("ss -ltn", false)
		if err == nil {
			switch w.c.State {
			case "started":
				if bytes.Contains(stdout, pattern) {
					return nil
				}
			case "stopped":
				if !bytes.Contains(stdout, pattern) {
					return nil
				}
			}
		}
		return err
	}, retryOpt); err != nil {
		return errors.Errorf("timed out waiting for port %d to be %s", w.c.Port, w.c.State)
	}
	return nil
}
