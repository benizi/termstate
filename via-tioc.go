// +build darwin dragonfly freebsd netbsd openbsd

package termstate

import "syscall"

const supported = true
const getIoctl = syscall.TIOCGETA
const setIoctl = syscall.TIOCSETA
