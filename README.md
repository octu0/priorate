# `priorate`

[![Apache License](https://img.shields.io/github/license/octu0/priorate)](https://github.com/octu0/priorate/blob/master/LICENSE)
[![GoDoc](https://godoc.org/github.com/octu0/priorate?status.svg)](https://godoc.org/github.com/octu0/priorate)
[![Go Report Card](https://goreportcard.com/badge/github.com/octu0/priorate)](https://goreportcard.com/report/github.com/octu0/priorate)
[![Releases](https://img.shields.io/github/v/release/octu0/priorate)](https://github.com/octu0/priorate/releases)

`priorate` provides ratelimit with priority using [golang.org/x/time/rate](https://pkg.go.dev/golang.org/x/time/rate).  
Priority can be defined as a ratio from 0.01 to 0.99, and ratelimit can be performed according to priority on a given limit.  
Fairly gets limit according to priority.

## Installation

```bash
go get github.com/octu0/priorate
```

## Example

Here's a quick example for using `priorate.NewLimiter`.

```go
import(
  "time"
  "fmt"

  "github.com/octu0/priorate"
)

func main() {
	limit := priorate.NewLimiter(100,
		priorate.Priority(High, 0.7),
		priorate.Priority(Low, 0.3),
	)
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

func printDelay(d time.Duration) {
	fmt.Println(d.Truncate(time.Millisecond))
}
```

## License

MIT, see LICENSE file for details.
