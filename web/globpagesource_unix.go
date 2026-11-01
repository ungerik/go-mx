//go:build unix

package web

import (
	"os"
	"syscall"
)

// openPageFileFlags opens a page file for reading without waiting for a
// writer: an open of a named pipe blocks until one connects, and the pipe
// would only be recognized and skipped once the open returned. A read of a
// regular file is not affected by O_NONBLOCK.
const openPageFileFlags = os.O_RDONLY | syscall.O_NONBLOCK
