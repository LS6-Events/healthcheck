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

func TestNewPubsubCheck(t *testing.T) {
	t.Parallel()

	aCheck, err := NewPubsubCheck("wibble-foo")
	if err != nil {
		t.Fatal(err)
	}
	defer aCheck.Cleanup()

	if aCheck.GetImp() != PUBSUB {
		t.Fatalf("pubsubCheck.GetImp() returned unexpected value: %s", aCheck.GetImp())
	}

	if aCheck.GetStatus() != STARTUP {
		t.Fatalf("pubsubCheck.GetStatus() returned unexpected value after initialisation: %s", aCheck.GetStatus())
	}

	if aCheck.GetLastCheck().Equal(time.Time{}) {
		t.Fatalf("pubsubCheck.GetLastCheck() returned unexpected value after initialisation: %v", aCheck.GetLastCheck())
	}

	if aCheck.GetError() != nil {
		t.Fatalf("pubsubCheck.GetError() returned unexpected value after initialisation: %v", aCheck.GetError())
	}
}

func TestPubsubHealthCheck(t *testing.T) {
	const (
		pubsubPort      = 8681
		pubsubProjectID = "wibble-foo"
	)

	cntrCtx := context.Background()

	cntrPort, err := nat.NewPort("tcp", fmt.Sprint(pubsubPort))
	if err != nil {
		t.Fatal(err)
	}

	psCntr, err := testcontainers.GenericContainer(cntrCtx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "gcr.io/google.com/cloudsdktool/cloud-sdk:367.0.0-emulators",
			ExposedPorts: []string{fmt.Sprint(pubsubPort)},
			WaitingFor:   wait.ForAll(wait.ForLog(fmt.Sprintf("Server started, listening on %d", pubsubPort))),
			Env: map[string]string{
				"PUBSUB_PROJECT1": fmt.Sprintf("%s,%s", pubsubProjectID, "health-check:health-check-sub"),
			},
			Cmd: []string{
				"/bin/sh",
				"-c",
				fmt.Sprintf("gcloud beta emulators pubsub start --host-port 0.0.0.0:%d", pubsubPort),
			},
		},
		Started: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := psCntr.Terminate(cntrCtx); err != nil {
			t.Fatal(err)
		}
	}()

	cntrHost, err := psCntr.Host(cntrCtx)
	if err != nil {
		t.Fatal(err)
	}
	cntrPort, err = psCntr.MappedPort(cntrCtx, cntrPort)
	if err != nil {
		t.Fatal(err)
	}

	t.Setenv("PUBSUB_EMULATOR_HOST", fmt.Sprintf("%s:%d", cntrHost, cntrPort.Int()))
	// t.Setenv("PUBSUB_EMULATOR", "localhost:8681")

	aCheck, err := NewPubsubCheck(pubsubProjectID)
	if err != nil {
		t.Fatal(err)
	}
	defer aCheck.Cleanup()

	err = aCheck.HealthCheck()
	if err != nil {
		t.Fatal(err)
	}

	if aCheck.GetStatus() != DONE {
		t.Fatalf("pubsubCheck.GetStatus() returned unexpected value after calling HealthCheck(): %s", aCheck.GetStatus())
	}

	if !(aCheck.GetLastCheck().Before(time.Now()) && aCheck.GetLastCheck().After(time.Time{})) {
		t.Fatalf("pubsubCheck.GetLastCheck() returned unexpected value after calling HealthCheck(): %v", aCheck.GetLastCheck())
	}

	if aCheck.GetError() != nil {
		t.Fatalf("pubsubCheck.GetError() returned unexpected value after calling HealthCheck(): %v", aCheck.GetError())
	}
}

func TestPubsubHealthCheckNoServer(t *testing.T) {
	t.Parallel()

	aCheck, err := NewPubsubCheck("wibble-foo")
	if err != nil {
		t.Fatal(err)
	}
	defer aCheck.Cleanup()

	err = aCheck.HealthCheck()
	if err == nil {
		t.Fatalf("postgresCheck.HealthCheck() did not return an error for a failing check")
	}

	if aCheck.GetStatus() != DONE {
		t.Fatalf("pubsubCheck.GetStatus() returned unexpected value after calling HealthCheck(): %s", aCheck.GetStatus())
	}

	if !(aCheck.GetLastCheck().Before(time.Now()) && aCheck.GetLastCheck().After(time.Time{})) {
		t.Fatalf("pubsubCheck.GetLastCheck() returned unexpected value after calling HealthCheck(): %v", aCheck.GetLastCheck())
	}

	if !errors.Is(aCheck.GetError(), err) {
		t.Fatalf("pubsubCheck.GetError() returned unexpected value after calling HealthCheck(): %v", aCheck.GetError())
	}
}
