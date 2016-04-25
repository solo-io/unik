package util

import (
	"github.com/Sirupsen/logrus"
	"time"
)

func Retry(retries int, sleep time.Duration, action func() error) error {
	if err := action(); err != nil {
		logrus.WithError(err).Warnf("retrying... %v", retries)
		if retries < 1 {
			return err
		}
		time.Sleep(sleep)
		return Retry(retries-1, sleep, action)
	}
	return nil
}
