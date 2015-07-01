package with

import (
	"errors"
	"math/rand"
	"time"
)

// ErrRetriesExceeded max number of retries exceeded.
var ErrRetriesExceeded = errors.New("retries exceeded")

// defaultMinWaitBeforeRetry is the default minimum wait time before
// a request is retried.
const defaultMinWaitBeforeRetry = 100

// defaultMaxWaitBeforeRetry is the default max wait time before
// a request is retried.
const defaultMaxWaitBeforeRetry = 5000

// RetryForever means we should retry requests forever. If requests shouldn't
// be retried forever then the upload.retryCount should be set to a positive
// number.
const RetryForever = -1

// RetryNever never retries a request.
const RetryNever = 0

type Retrier struct {
	MinWait    int
	MaxWait    int
	RetryCount int
}

func NewRetrier() Retrier {
	return Retrier{
		MinWait:    defaultMinWaitBeforeRetry,
		MaxWait:    defaultMaxWaitBeforeRetry,
		RetryCount: RetryForever,
	}
}

func (r Retrier) WithRetry(fn func() bool) error {
	retryCounter := 0
	for {
		if fn() {
			return nil
		}

		if r.RetryCount != RetryForever {
			retryCounter++
			if retryCounter > r.RetryCount {
				return ErrRetriesExceeded
			}
		}
		r.sleepRandom()
	}
}

func (r Retrier) sleepRandom() {
	// sleep a random amount between minWait and maxWait
	rand.Seed(time.Now().Unix())
	randomSleepTime := rand.Intn(r.MaxWait) + r.MinWait
	time.Sleep(time.Duration(randomSleepTime))
}
