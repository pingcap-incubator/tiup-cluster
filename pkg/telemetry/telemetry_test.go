package telemetry

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pingcap/check"
)

func Test(t *testing.T) { check.TestingT(t) }

var _ = check.Suite(&TelemetrySuite{})

type TelemetrySuite struct {
}

func (s *TelemetrySuite) TestReport(c *check.C) {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dst, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(400)
			return
		}

		msg := new(Report)
		err = msg.Unmarshal(dst)
		if err != nil {
			w.WriteHeader(400)
			return
		}

		if msg.EventUUID == "" {
			w.WriteHeader(400)
			return
		}
	}))

	defer ts.Close()

	tele := NewTelemetry()
	tele.cli = ts.Client()
	tele.url = ts.URL

	msg := new(Report)

	err := tele.Report(context.Background(), msg)
	c.Assert(err, check.NotNil)

	msg.EventUUID = "dfdfdf"
	err = tele.Report(context.Background(), msg)
	c.Assert(err, check.IsNil)
}
