package priorate

import (
	"fmt"
	"testing"
	"time"
)

func printDelay(d time.Duration) {
	// shows millisecond
	fmt.Println(d.Truncate(time.Millisecond))
}

func ExampleBasics() {
	limit := NewLimiter(100, Priority(High, 0.7), Priority(Low, 0.3))
	for i := 0; i < 10; i += 1 {
		if i < 5 {
			high := limit.ReserveN(High, time.Now(), 30)
			printDelay(high.Delay())
		} else {
			low := limit.ReserveN(Low, time.Now(), 30)
			printDelay(low.Delay())
		}
	}

	// Output:
	// 0s
	// 0s
	// 299ms
	// 1.099s
	// 1.899s
	// 1.199s
	// 1.966s
	// 3.233s
	// 4.499s
	// 5.766s
}

func ExampleSimple() {
	limit := NewLimiter(100, Priority(High, 0.7), Priority(Low, 0.3))
	for i := 0; i < 10; i += 1 {
		if i < 5 {
			high := limit.Reserve(High)
			printDelay(high.Delay())
		} else {
			low := limit.Reserve(Low)
			printDelay(low.Delay())
		}
	}

	// Output:
	// 299ms
	// 2.899s
	// 5.499s
	// 8.299s
	// 11.199s
	// 9.899s
	// 13.999s
	// 18.099s
	// 22.199s
	// 26.299s
}

func TestPriorityRate(t *testing.T) {
	t.Run("undefRate", func(tt *testing.T) {
		lim := NewLimiter(100, Priority(High, 0.1), Priority(Low, 0.2))
		if lim.rate[High] != 0.1 {
			tt.Errorf("specified High 0.1")
		}
		if lim.rate[Low] != 0.2 {
			tt.Errorf("specified Low 0.2")
		}
		if lim.rate[undef] != 0.7 {
			tt.Errorf("undefined 0.7")
		}
	})
	t.Run("single100", func(tt *testing.T) {
		lim := NewLimiter(100, Priority(Low, 1.0))
		if lim.rate[Low] != 1.0 {
			tt.Errorf("specifined Low 1.0")
		}
		if _, ok := lim.rate[undef]; ok {
			tt.Errorf("no define undef level")
		}
	})
	t.Run("noOption", func(tt *testing.T) {
		lim := NewLimiter(100)

		if lim.rate[High] != 0.5 {
			tt.Errorf("default High 0.5")
		}
		if lim.rate[Low] != 0.5 {
			tt.Errorf("default Low 0.5")
		}
		if _, ok := lim.rate[undef]; ok {
			tt.Errorf("no define undef level")
		}
	})
}

func TestReserveN(t *testing.T) {
	t.Run("single100", func(tt *testing.T) {
		lim := NewLimiter(100, Priority(Low, 1.0))
		r1 := lim.ReserveN(High, time.Now(), 100)
		r2 := lim.ReserveN(High, time.Now(), 100)

		if r1.Delay() != 0 {
			tt.Errorf("first token is 0")
		}
		if r2.Delay().Truncate(time.Millisecond) != (999 * time.Millisecond) {
			tt.Errorf("999ms < delay <1.0s: %s", r2.Delay())
		}
	})
	t.Run("noOption", func(tt *testing.T) {
		lim := NewLimiter(100)
		r1 := lim.ReserveN(High, time.Now(), 100)
		r2 := lim.ReserveN(Low, time.Now(), 100)
		if r1.Delay().Truncate(time.Millisecond) != (499 * time.Millisecond) {
			tt.Errorf("499ms < delay <500s: %s", r1.Delay())
		}
		if r2.Delay().Truncate(time.Millisecond) != (2499 * time.Millisecond) {
			tt.Errorf("2499ms < delay <2500s: %s", r2.Delay())
		}
	})
}

func TestInf(t *testing.T) {
	t.Run("delay", func(tt *testing.T) {
		lim := NewInf()
		for i := 0; i < 1000; i += 1 {
			var r *Reservation
			if (i % 2) == 0 {
				r = lim.Reserve(High)
			} else {
				r = lim.Reserve(Low)
			}

			if r.OK() != true {
				tt.Errorf("Inf always true")
			}
			if r.Delay() != 0 {
				tt.Errorf("Inf always 0 %v", r.Delay())
			}
		}
	})
}
