package report

import (
	"context"
	"testing"

	"github.com/pingcap-incubator/tiup-cluster/pkg/telemetry"
	"github.com/pingcap/check"
)

type reportSuite struct{}

var _ = check.Suite(&reportSuite{})

func TestT(t *testing.T) { check.TestingT(t) }

func (s *reportSuite) TestNodeInfo(c *check.C) {
	info := new(telemetry.NodeInfo)
	err := telemetry.FillNodeInfo(context.Background(), info)
	c.Assert(err, check.IsNil)

	text, err := NodeInfoToText(info)
	c.Assert(err, check.IsNil)

	info2, err := NodeInfoFromText(text)
	c.Assert(err, check.IsNil)
	c.Assert(info2, check.DeepEquals, info)
}
