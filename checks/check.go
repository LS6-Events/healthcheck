package checks

import "time"

type Implementation string
type CheckStatus string

const (
	// the check is preparing required resources.
	STARTUP CheckStatus = "startup"
	// the check is ongoing.
	CHECKING CheckStatus = "checking"
	// the check is complete.
	DONE CheckStatus = "done"
)

type CheckInterface interface {
	// GetImp returns the check's Implementation.
	GetImp() Implementation
	// GetStatus returns the check's current status.
	GetStatus() CheckStatus
	// GetLastCheck return the time the check last attempt finished.
	GetLastCheck() time.Time
	// GetError returns the error the last check attempt encountered,
	// if the check passed succesfully returns nil.
	GetError() error
	// HealthCheck runs the check.
	// If the check fails, returns the error encountered.
	// If the check succeeds returns nil.
	HealthCheck() error
	// Cleans up any resources and dependencies required by the check.
	Cleanup()
}
