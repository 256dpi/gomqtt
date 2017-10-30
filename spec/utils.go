package spec

import "time"

func safeReceive(ch chan struct{}) {
	select {
	case <-time.After(1 * time.Minute):
		panic("nothing received")
	case <-ch:
	}
}
