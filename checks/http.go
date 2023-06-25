package checks

import (
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

const HTTP Implementation = "http"

type httpCheck struct {
	status    CheckStatus
	lastCheck time.Time
	err       error
	host      string
	port      int
	timeout   time.Duration
}

func NewHttpCheck(host string, port int, timeout time.Duration) (CheckInterface, error) {
	check := httpCheck{
		status:    STARTUP,
		lastCheck: time.Unix(0, 0),
		err:       nil,
		host:      host,
		port:      port,
		timeout:   timeout,
	}

	return &check, nil
}

func (c *httpCheck) GetImp() Implementation {
	return HTTP
}

func (c *httpCheck) GetStatus() CheckStatus {
	return c.status
}

func (c *httpCheck) GetLastCheck() time.Time {
	return c.lastCheck
}

func (c *httpCheck) GetError() error {
	return c.err
}

func (c *httpCheck) HealthCheck() error {
	c.status = STARTUP

	urlStr := fmt.Sprintf("http://%s:%d", c.host, c.port)
	client := http.Client{
		Timeout: c.timeout,
	}

	c.status = CHECKING

	_, c.err = client.Get(urlStr)
	if c.err != nil {
		c.err = errors.Wrap(c.err, "error making http GET request")
		c.lastCheck = time.Now()
		c.status = DONE
		return c.err
	}

	c.lastCheck = time.Now()
	c.status = DONE
	return c.err
}

func (c *httpCheck) Cleanup() {}
