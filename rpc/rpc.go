package rpc

import (
	"errors"
	"net/rpc"
	"time"
)

func tryDial(network, addr string) (*rpc.Client, error) {
	for try := 0; try < 5; try++ {
		c, err := rpc.Dial(network, addr)
		if err == nil {
			return c, nil
		}
		time.Sleep(2 * time.Second)
	}
	return nil, errors.New("timeout")
}
