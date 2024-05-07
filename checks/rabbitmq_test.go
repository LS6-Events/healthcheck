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

func TestNewRabbitmqCheck(t *testing.T) {
	t.Parallel()

	aCheck, err := NewRabbitmqCheck("wibble", 1234, "foo", "bar")
	if err != nil {
		t.Fatal(err)
	}
	defer aCheck.Cleanup()

	if aCheck.GetImp() != RABBITMQ {
		t.Fatalf("rabbitmqCheck.GetImp() returned unexpected value: %s", aCheck.GetImp())
	}

	if aCheck.GetStatus() != STARTUP {
		t.Fatalf("rabbitmqCheck.GetStatus() returned unexpected value after initialisation: %s", aCheck.GetStatus())
	}

	if aCheck.GetLastCheck().Equal(time.Time{}) {
		t.Fatalf("rabbitmqCheck.GetLastCheck() returned unexpected value after initialisation: %v", aCheck.GetLastCheck())
	}

	if aCheck.GetError() != nil {
		t.Fatalf("rabbitmqCheck.GetError() returned unexpected value after initialisation: %v", aCheck.GetError())
	}
}

func TestRabbitmqHealthCheck(t *testing.T) {
	const (
		rmqPort = 5672
		rmqUser = "rabbitmq"
		rmqPass = "password"
	)

	cntrCtx := context.Background()

	cntrPort, err := nat.NewPort("tcp", fmt.Sprint(rmqPort))
	if err != nil {
		t.Fatal(err)
	}

	rmqCntr, err := testcontainers.GenericContainer(cntrCtx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "rabbitmq:alpine",
			ExposedPorts: []string{fmt.Sprint(rmqPort)},
			WaitingFor:   wait.ForAll(wait.ForListeningPort(cntrPort), wait.ForLog("Server startup complete;.*").AsRegexp()),
			Env: map[string]string{
				"RABBITMQ_DEFAULT_USER": rmqUser,
				"RABBITMQ_DEFAULT_PASS": rmqPass,
			},
		},
		Started: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := rmqCntr.Terminate(cntrCtx); err != nil {
			t.Fatal(err)
		}
	}()

	cntrHost, err := rmqCntr.Host(cntrCtx)
	if err != nil {
		t.Fatal(err)
	}
	cntrPort, err = rmqCntr.MappedPort(cntrCtx, cntrPort)
	if err != nil {
		t.Fatal(err)
	}

	aCheck, err := NewRabbitmqCheck(cntrHost, cntrPort.Int(), rmqUser, rmqPass)
	if err != nil {
		t.Fatal(err)
	}
	defer aCheck.Cleanup()

	err = aCheck.HealthCheck()
	if err != nil {
		t.Fatal(err)
	}

	if aCheck.GetStatus() != DONE {
		t.Fatalf("rabbitmqCheck.GetStatus() returned unexpected value after calling HealthCheck(): %s", aCheck.GetStatus())
	}

	if !(aCheck.GetLastCheck().Before(time.Now()) && aCheck.GetLastCheck().After(time.Time{})) {
		t.Fatalf("rabbitmqCheck.GetLastCheck() returned unexpected value after calling HealthCheck(): %v", aCheck.GetLastCheck())
	}

	if aCheck.GetError() != nil {
		t.Fatalf("rabbitmqCheck.GetError() returned unexpected value after calling HealthCheck(): %v", aCheck.GetError())
	}
}

func TestRabbitmqHealthCheckNoServer(t *testing.T) {
	t.Parallel()

	aCheck, err := NewRabbitmqCheck("wibble", 1234, "foo", "bar")
	if err != nil {
		t.Fatal(err)
	}
	defer aCheck.Cleanup()

	err = aCheck.HealthCheck()
	if err == nil {
		t.Fatalf("rabbitmqCheck.HealthCheck() did not return an error for a failing check")
	}

	if aCheck.GetStatus() != DONE {
		t.Fatalf("rabbitmqCheck.GetStatus() returned unexpected value after calling HealthCheck(): %s", aCheck.GetStatus())
	}

	if !(aCheck.GetLastCheck().Before(time.Now()) && aCheck.GetLastCheck().After(time.Time{})) {
		t.Fatalf("rabbitmqCheck.GetLastCheck() returned unexpected value after calling HealthCheck(): %v", aCheck.GetLastCheck())
	}

	if !errors.Is(aCheck.GetError(), err) {
		t.Fatalf("rabbitmqCheck.GetError() returned unexpected value after calling HealthCheck(): %v", aCheck.GetError())
	}
}
