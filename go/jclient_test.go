package main

import (
	"sync"
	"testing"
)

func TestJclient(t *testing.T) {
	var wg sync.WaitGroup
	total := 0
	for i := 0; i < 10000; i++ {
		wg.Add(1)
		total++
		go func(wg *sync.WaitGroup) {
			JsonClient()
			wg.Done()
		}(&wg)
	}
	wg.Wait()
}
