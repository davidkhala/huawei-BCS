//go:binary-only-package

package logging

import (
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)