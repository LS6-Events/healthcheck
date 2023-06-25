package checks

import (
	"context"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/pkg/errors"
)

const (
	PUBSUB Implementation = "pubsub"

	healthCheckTopic = "health-check"
)

type pubsubCheck struct {
	status    CheckStatus
	lastCheck time.Time
	err       error
	projectId string
}

func NewPubsubCheck(projectID string) (CheckInterface, error) {
	check := pubsubCheck{
		status:    STARTUP,
		lastCheck: time.Unix(0, 0),
		err:       nil,
		projectId: projectID,
	}

	return &check, nil
}

func (c *pubsubCheck) GetImp() Implementation {
	return PUBSUB
}

func (c *pubsubCheck) GetStatus() CheckStatus {
	return c.status
}

func (c *pubsubCheck) GetLastCheck() time.Time {
	return c.lastCheck
}

func (c *pubsubCheck) GetError() error {
	return c.err
}

func (c *pubsubCheck) HealthCheck() error {
	c.status = STARTUP

	c.status = CHECKING

	// Creating pubsub client
	var client *pubsub.Client
	client, c.err = pubsub.NewClient(context.Background(), c.projectId)
	if c.err != nil {
		c.err = errors.Wrap(c.err, "error creating new pubsub client")
		c.lastCheck = time.Now()
		c.status = DONE
		return c.err
	}
	defer client.Close()

	c.lastCheck = time.Now()
	c.status = DONE
	return c.err
}

func (c *pubsubCheck) Cleanup() {}
