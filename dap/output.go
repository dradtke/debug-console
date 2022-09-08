package dap

import (
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/dradtke/debug-console/types"
)

// ???: Is it overengineered to support multiple connections to this?

type connWithEncoder struct {
	conn net.Conn
	enc *gob.Encoder
}

type OutputBroadcaster struct {
	mu              sync.Mutex
	conns           []connWithEncoder
	l               net.Listener
	firstConnSeenCh chan struct{}
	inited          bool
}

func (b *OutputBroadcaster) listen() {
	defer b.l.Close()
	for {
		c, err := b.l.Accept()
		if err != nil {
			return
		}
		b.mu.Lock()
		firstConnSeen := len(b.conns) == 0
		b.conns = append(b.conns, connWithEncoder{conn: c, enc: gob.NewEncoder(c)})
		b.mu.Unlock()

		if firstConnSeen {
			close(b.firstConnSeenCh)
		}
	}
}

func NewOutputBroadcaster() (*OutputBroadcaster, error) {
	l, err := net.Listen("tcp", "") // TODO: make these configurable
	if err != nil {
		return nil, fmt.Errorf("BroadcastOutput: %w", err)
	}
	b := &OutputBroadcaster{l: l, conns: make([]connWithEncoder, 0, 1), firstConnSeenCh: make(chan struct{})}
	go b.listen()
	return b, nil
}

func (b *OutputBroadcaster) Broadcast(output types.OutputEvent) {
	// Make sure we have had at least one connection
	<-b.firstConnSeenCh

	b.mu.Lock()
	defer b.mu.Unlock()

	log.Printf("Broadcasting output to %d connections", len(b.conns))

	for _, c := range b.conns {
		if err := c.enc.Encode(output); err != nil {
			log.Printf("error broadcasting output: %s", err)
		}
	}
}

func (b *OutputBroadcaster) Stop() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.l.Close()
	for _, c := range b.conns {
		c.conn.Close()
	}
}
