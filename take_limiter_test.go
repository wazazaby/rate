package rate

import (
	"fmt"
	"time"
)

func ExampleTakeLimiter_Take() {
	limiter := NewTakeLimiter(100) // 100 ticks/s.
	prev := time.Now()
	for i := range 10 {
		tookAt := limiter.Take()
		fmt.Printf("Tick %d delayed by %s\n", i, tookAt.Sub(prev).Round(time.Millisecond))
		prev = tookAt
	}

	// Output:
	// Tick 0 delayed by 10ms
	// Tick 1 delayed by 10ms
	// Tick 2 delayed by 10ms
	// Tick 3 delayed by 10ms
	// Tick 4 delayed by 10ms
	// Tick 5 delayed by 10ms
	// Tick 6 delayed by 10ms
	// Tick 7 delayed by 10ms
	// Tick 8 delayed by 10ms
	// Tick 9 delayed by 10ms
}
