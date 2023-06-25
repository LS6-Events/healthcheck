package checks

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/pkg/errors"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestNewPostgresCheck(t *testing.T) {
	t.Parallel()

	aCheck, err := NewPostgresCheck("localhost", 1234, "wibble", "foo", "bar", "nil")
	if err != nil {
		t.Fatal(err)
	}
	defer aCheck.Cleanup()

	if aCheck.GetImp() != POSTGRES {
		t.Fatalf("postgresCheck.GetImp() returned unexpected value: %s", aCheck.GetImp())
	}

	if aCheck.GetStatus() != STARTUP {
		t.Fatalf("postgresCheck.GetStatus() returned unexpected value after initialisation: %s", aCheck.GetStatus())
	}

	if aCheck.GetLastCheck().Equal(time.Time{}) {
		t.Fatalf("postgresCheck.GetLastCheck() returned unexpected value after initialisation: %v", aCheck.GetLastCheck())
	}

	if aCheck.GetError() != nil {
		t.Fatalf("postgresCheck.GetError() returned unexpected value after initialisation: %v", aCheck.GetError())
	}
}

func TestPostgresHealthCheck(t *testing.T) {
	const (
		psqlPort    = 5432
		psqlUser    = "postgres"
		psqlPass    = "password"
		psqlDb      = "postgres"
		psqlAddrStr = "disable"
	)

	cntrCtx := context.Background()

	cntrPort, err := nat.NewPort("tcp", fmt.Sprint(psqlPort))
	if err != nil {
		t.Fatal(err)
	}

	psqlCntr, err := testcontainers.GenericContainer(cntrCtx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:latest",
			ExposedPorts: []string{fmt.Sprint(psqlPort)},
			WaitingFor:   wait.ForAll(wait.ForListeningPort(cntrPort), wait.ForLog("database system is ready to accept connections")),
			Env: map[string]string{
				"POSTGRES_USER":     psqlUser,
				"POSTGRES_PASSWORD": psqlPass,
				"POSTGRES_DB":       psqlDb,
			},
		},
		Started: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := psqlCntr.Terminate(cntrCtx); err != nil {
			t.Fatal(err)
		}
	}()

	cntrHost, err := psqlCntr.Host(cntrCtx)
	if err != nil {
		t.Fatal(err)
	}
	cntrPort, err = psqlCntr.MappedPort(cntrCtx, cntrPort)
	if err != nil {
		t.Fatal(err)
	}

	aCheck, err := NewPostgresCheck(cntrHost, cntrPort.Int(), psqlDb, psqlUser, psqlPass, psqlAddrStr)
	if err != nil {
		t.Fatal(err)
	}
	defer aCheck.Cleanup()

	err = aCheck.HealthCheck()
	if err != nil {
		t.Fatal(err)
	}

	if aCheck.GetStatus() != DONE {
		t.Fatalf("postgresCheck.GetStatus() returned unexpected value after calling HealthCheck(): %s", aCheck.GetStatus())
	}

	if !(aCheck.GetLastCheck().Before(time.Now()) && aCheck.GetLastCheck().After(time.Time{})) {
		t.Fatalf("postgresCheck.GetLastCheck() returned unexpected value after calling HealthCheck(): %v", aCheck.GetLastCheck())
	}

	if aCheck.GetError() != nil {
		t.Fatalf("postgresCheck.GetError() returned unexpected value after calling HealthCheck(): %v", aCheck.GetError())
	}
}

func TestPostgresHealthCheckNoServer(t *testing.T) {
	t.Parallel()

	aCheck, err := NewPostgresCheck("localhost", 1234, "wibble", "foo", "bar", "nil")
	if err != nil {
		t.Fatal(err)
	}
	defer aCheck.Cleanup()

	err = aCheck.HealthCheck()
	if err == nil {
		t.Fatalf("postgresCheck.HealthCheck() did not return an error for a failing check")
	}

	if aCheck.GetStatus() != DONE {
		t.Fatalf("postgresCheck.GetStatus() returned unexpected value after calling HealthCheck(): %s", aCheck.GetStatus())
	}

	if !(aCheck.GetLastCheck().Before(time.Now()) && aCheck.GetLastCheck().After(time.Time{})) {
		t.Fatalf("postgresCheck.GetLastCheck() returned unexpected value after calling HealthCheck(): %v", aCheck.GetLastCheck())
	}

	if !errors.Is(aCheck.GetError(), err) {
		t.Fatalf("postgresCheck.GetError() returned unexpected value after calling HealthCheck(): %v", aCheck.GetError())
	}
}
