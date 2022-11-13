package types

import "sync/atomic"

var seq int64

func NextSeq() int64 {
	return atomic.AddInt64(&seq, 1)
}
