package telemetry

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/pingcap/check"
)

type scrubSuite struct{}

var _ = check.Suite(&scrubSuite{})

func (s *scrubSuite) testScrubYaml(c *check.C, generate bool) {
	files, err := ioutil.ReadDir("./testdata")
	c.Assert(err, check.IsNil)

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), "yaml") {
			continue
		}

		c.Log("file: ", f.Name())

		data, err := ioutil.ReadFile(filepath.Join("./testdata", f.Name()))
		c.Assert(err, check.IsNil)

		hashs := make(map[string]struct{})
		hashs["host"] = struct{}{}

		scrubed, err := ScrubYaml(data, hashs)
		c.Assert(err, check.IsNil)

		outName := filepath.Join("./testdata", f.Name()+".out")
		if generate {
			err = ioutil.WriteFile(outName, scrubed, 0644)
			c.Assert(err, check.IsNil)
		} else {
			out, err := ioutil.ReadFile(outName)
			c.Assert(err, check.IsNil)
			c.Assert(scrubed, check.BytesEquals, out)
		}
	}
}

func (s *scrubSuite) TestScrubYaml(c *check.C) {
	s.testScrubYaml(c, false)
}
