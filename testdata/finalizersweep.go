package main

// Test that the heap stays consistent while the sweep of an idle pressure GC
// is not finished yet: allocation, runtime.GC, goroutine exit and
// runtime.ReadMemStats can all run before the sweep is done.

import (
	"runtime"
	"time"
)

type node struct {
	next  *node
	value int
}

var (
	live      []*node
	liveBytes [][]byte
	finalized int
	garbage   any
)

func main() {
	for round := 0; round < 20; round++ {
		// Register more finalizers than the pressure GC threshold.
		for i := 0; i < 200; i++ {
			n := &node{value: i}
			runtime.SetFinalizer(n, func(*node) { finalized++ })
			if i%10 == 0 {
				n.next = &node{value: i + 1}
				live = append(live, n)
			}
		}

		// Large live objects with interior pointers, and garbage between them.
		b := make([]byte, 20000+round*10)
		for j := range b {
			b[j] = byte(j)
		}
		liveBytes = append(liveBytes, b[100:])
		garbage = make([]byte, 30000)

		// Let the scheduler go idle, so that it marks and sweeps in slices.
		time.Sleep(time.Duration(round%5) * 3 * time.Millisecond)

		switch round % 4 {
		case 0:
			for i := 0; i < 1000; i++ {
				garbage = &node{value: i}
			}
		case 1:
			done := make(chan struct{})
			go func() {
				garbage = make([]byte, 100)
				close(done)
			}()
			<-done
			runtime.Gosched()
		case 2:
			var stats runtime.MemStats
			runtime.ReadMemStats(&stats)
			if stats.HeapAlloc == 0 || stats.HeapAlloc > stats.HeapSys {
				panic("unexpected heap statistics")
			}
		case 3:
			runtime.GC()
		}

		check()
	}
	if finalized == 0 {
		panic("no finalizer ran")
	}
	println("ok")
}

func check() {
	for _, n := range live {
		if n.value%10 != 0 || n.next.value != n.value+1 {
			panic("live node was overwritten")
		}
	}
	for _, b := range liveBytes {
		for j, v := range b {
			if v != byte(j+100) {
				panic("live byte slice was overwritten")
			}
		}
	}
}
