package main

import (
	"fmt"
	"runtime"
	"syscall"
	"time"

	"github.com/benizi/termstate"
)

func dump(state termstate.State, err error) {
	if err == nil {
		fmt.Printf("%#+v\n", state)
	} else {
		fmt.Printf("%s\n", err)
	}
}

func main() {
	if !termstate.IsSupported() {
		fmt.Printf("No support on %s/%s\n", runtime.GOOS, runtime.GOARCH)
		return
	}

	initial, err := termstate.Get()
	dump(initial, err)

	if err == nil {
		oldstate, err := initial.Cbreak().Set()
		dump(initial, err)
		if initial != oldstate {
			fmt.Println("Changed")
		}
	}
	initial.Set()

	defer termstate.DeferredReset(
		termstate.State.Cbreak,
		termstate.State.EchoOff,
	)()

	getchar := make(chan byte, 1)
	go func() {
		for {
			var buf [1]byte
			n, err := syscall.Read(0, buf[:])
			if n == 0 || err != nil {
				fmt.Println("Error reading from stdin", err)
				break
			}
			getchar <- buf[0]
		}
	}()

	fmt.Println("Reading characters")

	timeout := time.After(5 * time.Second)
	for {
		select {
		case <-timeout:
			fmt.Println("DONE")
			return
		case c := <-getchar:
			fmt.Printf("Read byte: [%d]\n", c)
		}
	}
}
