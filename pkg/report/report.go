package report

import (
	"bytes"
	"os"

	"github.com/gogo/protobuf/proto"
	"github.com/pingcap-incubator/tiup-cluster/pkg/telemetry"
	"github.com/pingcap-incubator/tiup/pkg/localdata"
	tiuptele "github.com/pingcap-incubator/tiup/pkg/telemetry"
)

// Enable return true if we enable telemetry.
func Enable() bool {
	s := os.Getenv(localdata.EnvNameTelemetryStatus)
	status := tiuptele.Status(s)
	return status == tiuptele.EnableStatus
}

// UUID return telemetry uuid.
func UUID() string {
	return os.Getenv(localdata.EnvNameTelemetryUUID)
}

// NodeInfoFromText get telemetry.NodeInfo from the text.
func NodeInfoFromText(text string) (info *telemetry.NodeInfo, err error) {
	info = new(telemetry.NodeInfo)
	err = proto.UnmarshalText(text, info)
	if err != nil {
		return nil, err
	}

	return
}

// NodeInfoToText get telemetry.NodeInfo in text.
func NodeInfoToText(info *telemetry.NodeInfo) (text string, err error) {
	buf := new(bytes.Buffer)
	err = proto.MarshalText(buf, info)
	if err != nil {
		return
	}
	text = buf.String()

	return
}
