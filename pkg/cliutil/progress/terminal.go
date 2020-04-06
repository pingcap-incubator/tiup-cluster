package progress

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/atomic"
	"golang.org/x/sys/unix"
)

var (
	termSizeWidth = atomic.Int32{}
)

func updateTerminalSize() error {
	ws, err := unix.IoctlGetWinsize(syscall.Stdout, unix.TIOCGWINSZ)
	if err != nil {
		return err
	}
	termSizeWidth.Store(int32(ws.Col))
	return nil
}

func moveCursorUp(w io.Writer, n int) {
	_, _ = fmt.Fprintf(w, "\033[%dA", n)
}

func moveCursorDown(w io.Writer, n int) {
	_, _ = fmt.Fprintf(w, "\033[%dB", n)
}

func moveCursorToLineStart(w io.Writer) {
	_, _ = fmt.Fprintf(w, "\r")
}

func clearLine(w io.Writer) {
	_, _ = fmt.Fprintf(w, "\033[2K")
}

func init() {
	_ = updateTerminalSize()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGWINCH)

	go func() {
		for {
			if _, ok := <-sigCh; !ok {
				return
			}
			_ = updateTerminalSize()
		}
	}()
}
