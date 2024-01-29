## The HealthCheck Module

#### The Problem
Modern software, particularly microservice based systems, often have 
multiple third party systems which they depend on, e.g. databases, 
message brokers, ip-geolocation APIs, etc...
Our software must be able to validate whether these dependencies are available.

#### The Solution
HealthCheck provides both a framework for any arbitrary check as well as 
many pre-existing checks for common software.
Upon system startup HealthCheck can be used to validate any and all 
dependencies, providing a guarantee that they are available.
Whilst Enabling clean and reliable handling of cases where required dependencies 
aren't available.

#### Installation
```bash
go get github.com/LS6-Events/healthcheck
```

#### Example
```go
aHealthManager, err = healthcheck.New(time.second, time.minute)
if err != nil {
    // handle error
}
defer aHealthManager.Cleanup()

aHttpCheck, err = checks.NewHttpCheck(
    "localhost", 80, 10*time.Second
)
if err != nil {
    // handle error
}

err = aHealthManager.Register(aHttpCheck)
if err != nil {
    // handle error
}

err = aHealthManager.Run()
if err != nil {
    // handle error
}
```

#### New Checks
If you require a check for an application, which we do not provide, and decide to  
build the check yourself, please create a PR to add it to the *checks* package.
