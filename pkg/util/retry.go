package util

import (
	"time"
)

func Retry(retries int, sleep time.Duration, action func() error) error {
	if err := action(); err != nil {
		if retries < 1 {
			return err
		}
		time.Sleep(sleep)
		return Retry(retries-1, sleep, action)
	}
	return nil
}
