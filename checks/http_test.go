package checks

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/pkg/errors"
)

const testHttpPort = 4560

func testHttpServer(port int, shutdown chan bool) {
	server := &http.Server{
		Addr: fmt.Sprintf(":%d", port),
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	go func() {
		<-shutdown
		server.Close()
	}()

	go server.ListenAndServe()
}

func TestNewHttpCheck(t *testing.T) {
	t.Parallel()

	aCheck, err := NewHttpCheck("localhost", testHttpPort, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer aCheck.Cleanup()

	if aCheck.GetImp() != HTTP {
		t.Fatalf("httpCheck.GetImp() returned unexpected value: %s", aCheck.GetImp())
	}

	if aCheck.GetStatus() != STARTUP {
		t.Fatalf("httpCheck.GetStatus() returned unexpected value after initialisation: %s", aCheck.GetStatus())
	}

	if aCheck.GetLastCheck().Equal(time.Time{}) {
		t.Fatalf("httpCheck.GetLastCheck() returned unexpected value after initialisation: %v", aCheck.GetLastCheck())
	}

	if aCheck.GetError() != nil {
		t.Fatalf("httpCheck.GetError() returned unexpected value after initialisation: %v", aCheck.GetError())
	}
}

func TestHttpHealthCheck(t *testing.T) {
	var shutdown chan bool = make(chan bool)

	testHttpServer(testHttpPort, shutdown)
	defer func() { shutdown <- true }()

	aCheck, err := NewHttpCheck("localhost", testHttpPort, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer aCheck.Cleanup()

	err = aCheck.HealthCheck()
	if err != nil {
		t.Fatal(err)
	}

	if aCheck.GetStatus() != DONE {
		t.Fatalf("httpCheck.GetStatus() returned unexpected value after calling HealthCheck(): %s", aCheck.GetStatus())
	}

	if !(aCheck.GetLastCheck().Before(time.Now()) && aCheck.GetLastCheck().After(time.Time{})) {
		t.Fatalf("httpCheck.GetLastCheck() returned unexpected value after calling HealthCheck(): %v", aCheck.GetLastCheck())
	}

	if aCheck.GetError() != nil {
		t.Fatalf("httpCheck.GetError() returned unexpected value after calling HealthCheck(): %v", aCheck.GetError())
	}
}

func TestHttpHealthCheckNoServer(t *testing.T) {
	t.Parallel()

	aCheck, err := NewHttpCheck("localhost", testHttpPort, 1*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer aCheck.Cleanup()

	err = aCheck.HealthCheck()
	if err == nil {
		t.Fatalf("httpCheck.HealthCheck() did not return an error for a failing check")
	}

	if aCheck.GetStatus() != DONE {
		t.Fatalf("httpCheck.GetStatus() returned unexpected value after calling HealthCheck(): %s", aCheck.GetStatus())
	}

	if !(aCheck.GetLastCheck().Before(time.Now()) && aCheck.GetLastCheck().After(time.Time{})) {
		t.Fatalf("httpCheck.GetLastCheck() returned unexpected value after calling HealthCheck(): %v", aCheck.GetLastCheck())
	}

	if !errors.Is(aCheck.GetError(), err) {
		t.Fatalf("httpCheck.GetError() returned unexpected value after calling HealthCheck(): %v", aCheck.GetError())
	}
}
