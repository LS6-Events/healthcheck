package checks

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

const POSTGRES Implementation = "postgres"

type postgresCheck struct {
	status    CheckStatus
	lastCheck time.Time
	err       error
	host      string
	port      int
	dbName    string
	user      string
	pass      string
	sslMode   string
}

func NewPostgresCheck(host string, port int, dbName, user, pass, sslMode string) (CheckInterface, error) {
	check := postgresCheck{
		status:    STARTUP,
		lastCheck: time.Unix(0, 0),
		err:       nil,
		host:      host,
		port:      port,
		dbName:    dbName,
		user:      user,
		pass:      pass,
		sslMode:   sslMode,
	}

	return &check, nil
}

func (c *postgresCheck) GetImp() Implementation {
	return POSTGRES
}

func (c *postgresCheck) GetStatus() CheckStatus {
	return c.status
}

func (c *postgresCheck) GetLastCheck() time.Time {
	return c.lastCheck
}

func (c *postgresCheck) GetError() error {
	return c.err
}

func (c *postgresCheck) HealthCheck() error {
	c.status = STARTUP

	psqlConnStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", c.host, c.port, c.user, c.pass, c.dbName, c.sslMode)

	c.status = CHECKING

	var conn *sql.DB
	conn, c.err = sql.Open("postgres", psqlConnStr)
	if c.err != nil {
		c.err = errors.Wrap(c.err, "error opening postgres connection")
		c.lastCheck = time.Now()
		c.status = DONE
		return c.err
	}
	defer conn.Close()

	c.err = conn.Ping()
	if c.err != nil {
		c.err = errors.Wrap(c.err, "error pinging postgres database")
		c.lastCheck = time.Now()
		c.status = DONE
		return c.err
	}

	c.lastCheck = time.Now()
	c.status = DONE
	return c.err
}

func (c *postgresCheck) Cleanup() {}
