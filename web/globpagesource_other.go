//go:build !unix

package web

import "os"

// openPageFileFlags opens a page file for reading. Outside of Unix a file
// matched below a directory is no named pipe that an open would wait on:
// the named pipes of Windows live in their own namespace, `\\.\pipe\`.
const openPageFileFlags = os.O_RDONLY
