package executor

import (
	"fmt"
	"io"
	"strings"

	"github.com/fatih/color"
)

var log io.Writer

// SetLogger set the logger.
func SetLogger(logger io.Writer) {
	log = logger
}

func logExe(dst string, cmd string, stdout []byte, stderr []byte, done bool, err error) {
	if log == nil {
		return
	}

	if err != nil {
		fmt.Fprintf(log, "error: %s\n", color.RedString(err.Error()))
	}

	cmd = strings.TrimLeft(cmd, pathENV)

	fmt.Fprintf(log, "[%s] cmd:\n %s\n, stdout:\n %s\n, stderr\n: %s\n done\n: %v\n\n", dst, cmd, stdout, stderr, done)
}
