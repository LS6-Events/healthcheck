package checks

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	rabbitmq "github.com/rabbitmq/amqp091-go"
)

const (
	RABBITMQ Implementation = "rabbitmq"

	healthCheckQueue = "health-check"
)

type rabbitmqCheck struct {
	status    CheckStatus
	lastCheck time.Time
	err       error
	user      string
	pass      string
	host      string
	port      int
}

func NewRabbitmqCheck(host string, port int, user, pass string) (CheckInterface, error) {
	check := rabbitmqCheck{
		status:    STARTUP,
		lastCheck: time.Unix(0, 0),
		err:       nil,
		user:      user,
		pass:      pass,
		host:      host,
		port:      port,
	}

	return &check, nil
}

func (c *rabbitmqCheck) GetImp() Implementation {
	return RABBITMQ
}

func (c *rabbitmqCheck) GetStatus() CheckStatus {
	return c.status
}

func (c *rabbitmqCheck) GetLastCheck() time.Time {
	return c.lastCheck
}

func (c *rabbitmqCheck) GetError() error {
	return c.err
}

func (c *rabbitmqCheck) HealthCheck() error {
	c.status = STARTUP

	rmqConnStr := fmt.Sprintf("amqp://%s:%s@%s:%d/", c.user, c.pass, c.host, c.port)

	c.status = CHECKING

	// Establishing connection to rabbitmq instance.
	var rmqConn *rabbitmq.Connection
	rmqConn, c.err = rabbitmq.Dial(rmqConnStr)
	if c.err != nil {
		c.err = errors.Wrap(c.err, "error connecting to rabbitmq instance")
		c.lastCheck = time.Now()
		c.status = DONE
		return c.err
	}
	defer rmqConn.Close()

	// Establishing channel to rabbitmq instance.
	var rmqChan *rabbitmq.Channel
	rmqChan, c.err = rmqConn.Channel()
	if c.err != nil {
		c.err = errors.Wrap(c.err, "error establishing channel to rabbitmq")
		c.lastCheck = time.Now()
		c.status = DONE
		return c.err
	}
	defer rmqChan.Close()

	c.lastCheck = time.Now()
	c.status = DONE
	return c.err
}

func (c *rabbitmqCheck) Cleanup() {}
