[![license](http://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/udhos/refresh/blob/main/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/udhos/refresh)](https://goreportcard.com/report/github.com/udhos/refresh)
[![Go Reference](https://pkg.go.dev/badge/github.com/udhos/refresh.svg)](https://pkg.go.dev/github.com/udhos/refresh)

# refresh

[refresh](https://github.com/udhos/refresh) delivers [Spring Cloud Config Bus](https://www.baeldung.com/spring-cloud-bus) refresh notifications to Go applications.

# Usage

See example: [examples/refresh-example/main.go](examples/refresh-example/main.go)

    import "github.com/udhos/refresh/refresh"

    amqpURL := "amqp://guest:guest@rabbitmq:5672/"
    me := filepath.Base(os.Args[0])
    debug := true

    // "#" means receive notification for all applications
    refresher := refresh.New(amqpURL, me, []string{"#"}, debug, nil)

    for app := range refresher.C {
        log.Printf("refresh: received notification for application='%s'", app)
    }

# Test

1. Run ./run-local-rabbit.sh

2. Run refresh-example

Build and run:

    go install ./examples/refresh-example
    refresh-example

3. Post a message

Open http://localhost:15672/

Login with guest:guest

Open exchange springCloudBus: http://localhost:15672/#/exchanges/%2F/springCloudBus

Expand "Publish message".

Publish a sample message like this:

    {
        "type": "RefreshRemoteApplicationEvent",
        "timestamp": 1649804650957,
        "originService": "config-server:0:0a36277496365ee8621ae8f3ce7032ce",
        "destinationService": "refresh-example:**",
        "id": "5a4cb501-652a-4ae2-9d3e-279e1d2a2169"
    }

Check the logs for refresh-example application.
