package spec

import "time"

func safeReceive(ch chan struct{}) {
	select {
	case <-time.After(10 * time.Second):
		panic("nothing received")
	case <-ch:
	}
}
