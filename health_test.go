package healthcheck

import (
	"errors"
	"testing"
	"time"

	"github.com/LS6-Events/healthcheck/checks"
)

// ========== Custom checks required for unit testing ==========

const (
	TEST checks.Implementation = "test"
	FAIL checks.Implementation = "fail"
)

var ErrFailCheck = errors.New("fail check is configured to always fail")

type testCheck struct {
	status    checks.CheckStatus
	lastCheck time.Time
	err       error
}

type failCheck struct {
	status    checks.CheckStatus
	lastCheck time.Time
	err       error
}

func NewTestCheck() checks.CheckInterface {
	check := testCheck{
		status:    checks.STARTUP,
		lastCheck: time.Unix(0, 0),
		err:       nil,
	}

	return &check
}

func NewFailCheck() checks.CheckInterface {
	check := failCheck{
		status:    checks.STARTUP,
		lastCheck: time.Unix(0, 0),
		err:       nil,
	}

	return &check
}

func (c *testCheck) GetImp() checks.Implementation {
	return TEST
}

func (c *testCheck) GetStatus() checks.CheckStatus {
	return c.status
}

func (c *testCheck) GetLastCheck() time.Time {
	return c.lastCheck
}

func (c *testCheck) GetError() error {
	return c.err
}

func (c *testCheck) HealthCheck() error {
	c.status = checks.STARTUP

	c.status = checks.CHECKING

	c.err = nil
	c.lastCheck = time.Now()
	c.status = checks.DONE

	return c.err
}

func (c *testCheck) Cleanup() {}

func (c *failCheck) GetImp() checks.Implementation {
	return TEST
}

func (c *failCheck) GetStatus() checks.CheckStatus {
	return c.status
}

func (c *failCheck) GetLastCheck() time.Time {
	return c.lastCheck
}

func (c *failCheck) GetError() error {
	return c.err
}

func (c *failCheck) HealthCheck() error {
	c.status = checks.STARTUP

	c.status = checks.CHECKING

	c.err = ErrFailCheck
	c.lastCheck = time.Now()
	c.status = checks.DONE

	return c.err
}

func (c *failCheck) Cleanup() {}

// ========== Unit Tests ==========

func TestNew(t *testing.T) {
	t.Parallel()

	aHealthManager, err := New(time.Second, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	defer aHealthManager.Cleanup()

	if aHealthManager.GetHealth() != false {
		t.Fatal("GetHealth() returned unexpected value after initialisation")
	}
}

func TestNewInvalidCheckFrequency(t *testing.T) {
	t.Parallel()

	_, err := New(0, time.Minute)
	if err != ErrInvalidConfig {
		t.Fatalf("New() did not returned expected error\nexpected: %v\ngot: %v", ErrInvalidConfig, err)
	}
}

func TestNewInvalidTimeout(t *testing.T) {
	t.Parallel()

	_, err := New(time.Second, 0)
	if err != ErrInvalidConfig {
		t.Fatalf("New() did not returned expected error\nexpected: %v\ngot: %v", ErrInvalidConfig, err)
	}
}

func TestRegister(t *testing.T) {
	t.Parallel()

	aHealthManager, err := New(time.Second, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	defer aHealthManager.Cleanup()

	err = aHealthManager.Register(NewTestCheck())
	if err != nil {
		t.Fatal(err)
	}

	err = aHealthManager.Register(NewFailCheck())
	if err != nil {
		t.Fatal(err)
	}
}

func TestRegisterInvalid(t *testing.T) {
	t.Parallel()

	aHealthManager, err := New(time.Second, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	defer aHealthManager.Cleanup()

	err = aHealthManager.Register(nil)
	if err != ErrInvalidCheck {
		t.Fatalf("Register() did not returned expected error\nexpected: %v\ngot: %v", ErrInvalidCheck, err)
	}
}

func TestHealthManagerRun(t *testing.T) {
	aHealthManager, err := New(time.Second, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	defer aHealthManager.Cleanup()

	err = aHealthManager.Register(NewTestCheck())
	if err != nil {
		t.Fatal(err)
	}

	err = aHealthManager.Run()
	if err != nil {
		t.Fatal(err)
	}
}

func TestHealthManagerRunFail(t *testing.T) {
	aHealthManager, err := New(time.Millisecond, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer aHealthManager.Cleanup()

	err = aHealthManager.Register(NewFailCheck())
	if err != nil {
		t.Fatal(err)
	}

	err = aHealthManager.Run()
	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("Run() did not returned expected error for a failing test\nexpected: %v\ngot: %v", ErrTimeout, err)
	}
}

func TestHealthManagerMultipleChecks(t *testing.T) {
	aHealthManager, err := New(time.Second, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	defer aHealthManager.Cleanup()

	err = aHealthManager.Register(NewTestCheck())
	if err != nil {
		t.Fatal(err)
	}

	err = aHealthManager.Register(NewTestCheck())
	if err != nil {
		t.Fatal(err)
	}

	err = aHealthManager.Register(NewFailCheck())
	if err != nil {
		t.Fatal(err)
	}

	if len(aHealthManager.checks) != 3 {
		t.Fatalf("Register() did not successfully register all checks")
	}
}
