package priorate

import (
	"math"
	"time"

	"golang.org/x/time/rate"
)

type Level uint8

const (
	undef Level = iota
	Low
	High
)

type PriorityFunc func(map[Level]float64)

func Priority(p Level, rate float64) PriorityFunc {
	return func(m map[Level]float64) {
		if rate < 0.0 || 1.0 < rate {
			panic("rate must be specified with 0.01 <= x < 1.0")
		}

		m[p] = rate
	}
}

type Limiter struct {
	limit   int
	rate    map[Level]float64
	limiter map[Level]*rate.Limiter
}

func (lim *Limiter) Limit() int {
	return lim.limit
}

func (lim *Limiter) Reserve(lv Level) *Reservation {
	return lim.ReserveN(lv, time.Now(), lim.limit)
}

func (lim *Limiter) ReserveN(lv Level, now time.Time, n int) *Reservation {
	reserves := make([]*rate.Reservation, 0, len(lim.rate))
	remainN := n
	if r, ok := lim.rate[lv]; ok {
		rateN := int(float64(n) * r)
		if rateN < 1 {
			rateN = 1
		}
		remainN -= rateN
		reserves = append(reserves, lim.limiter[lv].ReserveN(now, rateN))
	}

	if remainN < 1 {
		return &Reservation{
			reserves: reserves,
		}
	}

	for otherLv, r := range lim.rate {
		if lv == otherLv || undef == otherLv {
			continue
		}

		rateN := int(float64(remainN) * r)
		if rateN < 1 {
			continue
		}
		remainN -= rateN
		reserves = append(reserves, lim.limiter[otherLv].ReserveN(now, rateN))
	}

	if 0 < remainN {
		if _, ok := lim.rate[lv]; ok {
			// remain all specified level
			reserves = append(reserves, lim.limiter[lv].ReserveN(now, remainN))
		} else {
			reserves = append(reserves, lim.limiter[undef].ReserveN(now, remainN))
		}
	}

	return &Reservation{reserves}
}

func InfLimiter() *Limiter {
	priorityRate := map[Level]float64{
		undef: 100,
	}
	priorityLimiter := map[Level]*rate.Limiter{
		undef: rate.NewLimiter(rate.Inf, math.MaxInt),
	}
	return &Limiter{
		limit:   math.MaxInt,
		rate:    priorityRate,
		limiter: priorityLimiter,
	}
}

func NewLimiter(limit int, funcs ...PriorityFunc) *Limiter {
	priorityRate := make(map[Level]float64, 8)
	for _, fn := range funcs {
		fn(priorityRate)
	}
	if len(priorityRate) < 1 {
		priorityRate[High] = 0.5
		priorityRate[Low] = 0.5
	}

	totalRate := 0.0
	for _, r := range priorityRate {
		totalRate += r
	}
	if 1.0 < totalRate {
		panic("total rate must not exceed 1.0")
	}

	if totalRate < 1.0 {
		undefRate := 1.0 - totalRate
		priorityRate[undef] = undefRate
	}

	priorityLimiter := make(map[Level]*rate.Limiter, len(priorityRate)+1)
	for lv, r := range priorityRate {
		adj := int(float64(limit) * r)
		priorityLimiter[lv] = rate.NewLimiter(rate.Limit(float64(adj)), adj)
	}

	return &Limiter{
		limit:   limit,
		rate:    priorityRate,
		limiter: priorityLimiter,
	}
}

type Reservation struct {
	reserves []*rate.Reservation
}

func (r *Reservation) Cancel() {
	r.CancelAt(time.Now())
}

func (r *Reservation) CancelAt(now time.Time) {
	for _, reserve := range r.reserves {
		reserve.CancelAt(now)
	}
}

func (r *Reservation) Delay() time.Duration {
	return r.DelayFrom(time.Now())
}

func (r *Reservation) DelayFrom(now time.Time) time.Duration {
	totalDelay := time.Duration(0)
	for _, reserve := range r.reserves {
		totalDelay += reserve.Delay()
	}
	return totalDelay
}

func (r *Reservation) OK() bool {
	for _, reserve := range r.reserves {
		if reserve.OK() != true {
			return false
		}
	}
	return true
}
