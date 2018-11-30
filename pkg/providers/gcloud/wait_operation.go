package gcloud

import (
	"github.com/sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"google.golang.org/api/compute/v1"
	"time"
)

var defaultTimeout = time.Minute * 5
var defaultInterval = time.Millisecond * 250

func (p *GcloudProvider) waitOperation(operation string, global bool) error {
	errc := make(chan error)
	finished := make(chan struct{})

	backoff := int64(1)
	go func() {
		for {
			done, err := p.waitCycle(operation, global)
			if err != nil {
				errc <- err
				return
			}
			if done {
				close(finished)
				return
			}
			backoff *= 2
			time.Sleep(time.Duration(backoff) * defaultInterval)
		}
	}()

	select {
	case err := <-errc:
		return err
	case <-finished:
		return nil
	case <-time.After(defaultTimeout):
		return errors.New("timed out waiting more than "+defaultTimeout.String()+" for "+operation+" to complete", nil)
	}
}

func (p *GcloudProvider) waitCycle(operation string, global bool) (bool, error) {
	var status *compute.Operation
	var err error
	if global {
		status, err = p.compute().GlobalOperations.Get(p.config.ProjectID, operation).Do()
	} else {
		status, err = p.compute().ZoneOperations.Get(p.config.ProjectID, p.config.Zone, operation).Do()
	}
	if err != nil {
		return false, errors.New("getting status for operation "+operation, err)
	}
	logrus.Debugf("status for %v is %+v", operation, status)
	if status.Status == "DONE" {
		return true, nil
	}
	return false, nil
}
