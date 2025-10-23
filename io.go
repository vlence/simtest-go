package simtest

import (
	"bytes"
	"io"
	"net"
	"sync"
)

// A thin wrapper around the I/O API that the application
// uses.
type IO interface {
        Listen(network, addr string) (net.Listener, error)

        Dial(network, addr string) (net.Conn, error)

        Open() (io.ReadWriteCloser, error)
}

type SimIO struct {
        pool sync.Pool
        files map[string]bytes.Buffer
        tcpListeners map[string]net.Listener
}

type simNetAddr struct {
        network string
        address string
}

func (addr *simNetAddr) Network() string {
        return addr.network
}

func (addr *simNetAddr) String() string {
        return addr.address
}

func NewSimIO() {
}
