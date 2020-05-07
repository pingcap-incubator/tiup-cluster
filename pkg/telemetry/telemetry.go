package telemetry

import (
	"bytes"
	"context"
	"net/http"

	"github.com/pingcap/errors"
)

var defaultURL = "https://TODO/v1/telemetry"

// Telemetry control telemetry.
type Telemetry struct {
	url string
	cli *http.Client
}

// NewTelemetry return a new Telemetry instance.
func NewTelemetry() *Telemetry {
	cli := new(http.Client)

	return &Telemetry{
		url: defaultURL,
		cli: cli,
	}
}

// Report report the msg right away.
func (t *Telemetry) Report(ctx context.Context, msg *Report) error {
	dst, err := msg.Marshal()
	if err != nil {
		return errors.AddStack(err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", t.url, bytes.NewReader(dst))
	if err != nil {
		return errors.AddStack(err)
	}

	req.Header.Add("Content-Type", "application/x-protobuf")
	resp, err := t.cli.Do(req)
	if err != nil {
		return errors.AddStack(err)
	}

	code := resp.StatusCode
	if code < 200 || code >= 300 {
		return errors.Errorf("StatusCode: %d, Status: %s", resp.StatusCode, resp.Status)
	}

	return nil
}
