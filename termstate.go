package termstate

import (
	"errors"
	"syscall"
	"unsafe"
)

// State holds information about a terminal.
type State syscall.Termios

var errNoSupport = errors.New("Unsupported platform")

func callIoctl(fd, ioctl int, state *State) (err error) {
	if !supported {
		return errNoSupport
	}
	_, _, errno := syscall.Syscall6(
		syscall.SYS_IOCTL,
		uintptr(fd),
		uintptr(ioctl),
		uintptr(unsafe.Pointer(state)),
		0, 0, 0)
	if errno != 0 {
		err = errno
	}
	return err
}

// IsSupported indicates whether termstate is supported on this platform.
func IsSupported() bool {
	return supported
}

// Get fetches the current terminal state for stdin.
func Get() (State, error) {
	return GetFD(0)
}

// GetFD fetches the current terminal state for the given file descriptor.
func GetFD(fd int) (state State, err error) {
	err = callIoctl(fd, getIoctl, &state)
	return state, err
}

type modifier func(State) State

// DeferredReset is a convenience function for setting modes on the current
// state, then resetting them later.
//
// Example usage:
//
//   defer termstate.DeferredReset(
//     termstate.State.Cbreak,
//     termstate.State.EchoOff,
//   )()
//
func DeferredReset(modifiers ...modifier) func() {
	// TODO(benizi): convert example to proper godoc

	initial, err := Get()
	state := initial
	if err == nil {
		for _, modifier := range modifiers {
			state = modifier(state)
		}
	}
	state.Set()

	return func() {
		if err == nil {
			initial.Set()
		}
	}
}

// Cbreak adds `cbreak` (also called `rare`) mode to the state.
func (state State) Cbreak() State {
	c := state
	// set noncanonical mode
	c.Lflag &^= syscall.ICANON
	// `read` should block until 1 byte is read
	c.Cc[syscall.VMIN] = 1
	// `read` should not timeout
	c.Cc[syscall.VTIME] = 0
	return c
}

const echos = syscall.ECHO | syscall.ECHONL

// Echo sets the `ECHO` and `ECHONL` properties.
func (state State) Echo(on bool) State {
	c := state
	if on {
		c.Lflag |= echos
	} else {
		c.Lflag &^= echos
	}
	return c
}

// EchoOn is for use in `DeferredReset` (equivalent to `Echo(true)`).
func (state State) EchoOn() State {
	return state.Echo(true)
}

// EchoOff is for use in `DeferredReset` (equivalent to `Echo(false)`).
func (state State) EchoOff() State {
	return state.Echo(false)
}

// Set attempts to set the terminal state for stdin.
func (state State) Set() (oldstate State, err error) {
	return state.SetFD(0)
}

// SetFD attempts to set the terminal to this `State`, returning the original
// state and or an error.
func (state State) SetFD(fd int) (oldstate State, err error) {
	err = callIoctl(fd, getIoctl, &oldstate)
	if err == nil {
		err = callIoctl(fd, setIoctl, &state)
	}
	return oldstate, err
}
