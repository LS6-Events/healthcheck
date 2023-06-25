package healthcheck

import (
	"fmt"
	"sync"
	"time"

	"github.com/LS6-Events/healthcheck/checks"
	"github.com/pkg/errors"
)

var ErrInvalidConfig = errors.New("invalid health manager configuration")
var ErrInvalidCheck = errors.New("cannot register invalid check")
var ErrTimeout = errors.New("health check timed out with checks failing")

type HealthManager struct {
	checkFreq time.Duration
	timeout   time.Duration
	healthy   bool
	mtx       sync.Mutex
	checks    []checks.CheckInterface
	startTime time.Time
}

// New returns a new HealthManager instance.
func New(CheckFrequency, timeout time.Duration) (*HealthManager, error) {
	if CheckFrequency <= 0 {
		return nil, ErrInvalidConfig
	}

	if timeout <= 0 {
		return nil, ErrInvalidConfig
	}

	return &HealthManager{
		checkFreq: CheckFrequency,
		timeout:   timeout,
		healthy:   false,
		checks:    make([]checks.CheckInterface, 0),
	}, nil
}

// GetHealth returns the health manager's current healthy status.
func (hm *HealthManager) GetHealth() bool {
	return hm.healthy
}

// Register registers a new check with the health manager.
func (hm *HealthManager) Register(c checks.CheckInterface) error {
	if c == nil {
		return ErrInvalidCheck
	}

	hm.mtx.Lock()
	defer hm.mtx.Unlock()

	hm.checks = append(hm.checks, c)
	return nil
}

// Run executes the registered checks until all return healthy or the timeout elapses.
// Runs all checks then sleeps until the check frequency elapses before re-running checks.
// New checks cannot be registerd whilst Run is ongoing.
func (hm *HealthManager) Run() error {
	hm.mtx.Lock()
	defer hm.mtx.Unlock()

	hm.healthy = false
	hm.startTime = time.Now()

	for !hm.healthy {
		var wg sync.WaitGroup
		for i := 0; i < len(hm.checks); i++ {
			wg.Add(1)
			x := i
			go func() {
				hm.checks[x].HealthCheck()
				wg.Done()
			}()
		}
		wg.Wait()

		var failed bool = false
		for i := 0; i < len(hm.checks); i++ {
			if hm.checks[i].GetError() != nil {
				failed = true
			}
		}

		if !failed {
			hm.healthy = true
			return nil
		}

		if hm.startTime.Add(hm.timeout).Before(time.Now()) {
			var timeoutErr error = ErrTimeout
			for i := 0; i < len(hm.checks); i++ {
				if hm.checks[i].GetError() != nil {
					timeoutErr = errors.Wrap(timeoutErr, fmt.Sprintf("%s:%s", hm.checks[i].GetImp(), hm.checks[i].GetError().Error()))
				}
			}
			return timeoutErr
		}

		time.Sleep(hm.checkFreq)
	}

	return nil
}

// Cleanup cleans up any resources required by the health manager and any registered checks.
func (hm *HealthManager) Cleanup() {
	hm.mtx.Lock()
	defer hm.mtx.Unlock()

	for _, check := range hm.checks {
		check.Cleanup()
	}
}
