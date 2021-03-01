package stream

import (
	"io"

	"k8s.io/client-go/tools/remotecommand"
)

// standard implementation in /k8s.io/kubectl/pkg/util/term/resize.go
// here is an incomplete implementation
// sizeQueue implements remotecommand.TerminalSizeQueue
type sizeQueue struct {
	// resizeChan receives a Size each time the user's terminal is resized.
	resizeChan chan remotecommand.TerminalSize
	doneChan   chan struct{}
}

// Next returns the new terminal size after the terminal has been resized. It returns nil when
// session stoped.
func (s *sizeQueue) Next() *remotecommand.TerminalSize {
	select {
	case size := <-s.resizeChan:
		return &size
	case <-s.doneChan:
		return nil
	}
}

// IOStreams provides the standard names for iostreams.  This is useful for embedding and for unit testing.
// Inconsistent and different names make it hard to read and review code
type IOStreams struct {
	// In think, os.Stdin
	In io.Reader
	// Out think, os.Stdout
	Out io.Writer
	// ErrOut think, os.Stderr
	ErrOut io.Writer
}

// TerminalSession implements PtyHandler, holds information pertaining to the current streaming session:
// input/output streams, if the client is requesting a TTY, and a terminal size queue to
// support terminal resizing.
type TerminalSession struct {
	IOStreams
	sizeQueue
	tty bool
}

// Done done, must call Done() before connection close, or Next() would not exits.
func (t *TerminalSession) Done() {
	close(t.doneChan)
}

// Tty ...
func (t *TerminalSession) Tty() bool {
	return t.tty
}

// Stdin ...
func (t *TerminalSession) Stdin() io.Reader {
	return t.IOStreams.In
}

// Stdout ...
func (t *TerminalSession) Stdout() io.Writer {
	return t.IOStreams.Out
}

// Stderr ...
func (t *TerminalSession) Stderr() io.Writer {
	return t.IOStreams.ErrOut
}

// NewTerminalSession create TerminalSession
func NewTerminalSession(stream IOStreams) *TerminalSession {
	return &TerminalSession{
		IOStreams: stream,
		tty:       false,
		sizeQueue: sizeQueue{
			resizeChan: make(chan remotecommand.TerminalSize),
			doneChan:   make(chan struct{}),
		},
	}
}
