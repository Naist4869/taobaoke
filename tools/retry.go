package tools

import (
	"errors"
	"math/rand"
	"syscall"
	"time"
)

func Retry(f func() (err error, mayRetry bool)) error {
	var (
		bestErr     error
		lowestErrno syscall.Errno
		start       time.Time
		nextSleep   = 1 * time.Second
	)
	for {
		err, mayRetry := f()
		if err == nil || !mayRetry {
			return err
		}
		var errno syscall.Errno
		// 是系统调用错误
		if errors.As(err, &errno) && (lowestErrno == 0 || errno < lowestErrno) {
			bestErr = err
			lowestErrno = errno
		} else if bestErr == nil {
			bestErr = err
		}

		if start.IsZero() {
			start = time.Now()
			// 超过1分钟还报错的话就返回错误
		} else if d := time.Since(start) + nextSleep; d >= time.Minute {
			break
		}
		time.Sleep(nextSleep)
		nextSleep += time.Duration(rand.Int63n(int64(nextSleep)))
	}
	return bestErr
}
