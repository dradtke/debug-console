package util

import (
	"errors"
	"net/rpc"
	"time"
)

func TryDial(network, addr string) (*rpc.Client, error) {
	for try := 0; try < 6; try++ {
		c, err := rpc.Dial(network, addr)
		if err == nil {
			return c, nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return nil, errors.New("timeout")
}
