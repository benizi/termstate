// +build linux

package termstate

import "syscall"

const supported = true
const getIoctl = syscall.TCGETS
const setIoctl = syscall.TCSETS
