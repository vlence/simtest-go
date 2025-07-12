package dst

import (
	"time"
)

// Sleepers can put goroutines to sleep.
type Sleeper interface {
        // Sleep puts the current goroutine to sleep for the specified duration.
        Sleep(time.Duration)
}
