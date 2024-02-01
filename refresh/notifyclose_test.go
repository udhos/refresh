package refresh

import (
	"testing"
	"time"
)

// go test -v ./refresh -run=TestCloseNotifyChannel
func TestCloseNotifyChannel(t *testing.T) {

	debug := true

	mock := &amqpMock{}

	o := Options{
		AmqpURL:      "amqp://guest:guest@rabbitmq:5672/",
		ConsumerTag:  "app",
		Applications: []string{"app"},
		Debug:        debug,
		AmqpClient:   mock,
	}

	r := New(o)

	if r == nil {
		t.Errorf("ugh")
	}

	wait := time.Second
	t.Logf("giving a time before close-notifying the amqp consumer goroutine: %v", wait)
	time.Sleep(wait)

	t.Logf("close-notifying the amqp consumer")

	begin := time.Now()

	timeout := 10 * time.Second
	sleep := 2 * time.Second

	// get non-zero initial connection count

	var connsInitial int

	for {
		connsInitial = r.statConnGet()
		t.Logf("connections count: initial=%d", connsInitial)
		if connsInitial > 0 {
			break
		}
		elap := time.Since(begin)
		if elap > timeout {
			t.Errorf("could not get non-zero initial connection count: timeout=%v", timeout)
			return
		}
		t.Logf("could not get non-zero initial connection count, sleeping for %v", sleep)
		time.Sleep(sleep)
	}

	// send notify close

	mock.channelNotifyCloseSendMock(nil)

	// look for connection count increase

	for {
		conns := r.statConnGet()
		t.Logf("connections count: initial=%d current=%d", connsInitial, conns)
		if conns > connsInitial {
			break
		}
		elap := time.Since(begin)
		if elap > timeout {
			t.Errorf("refresh goroutine has never reconnected: timeout=%v", timeout)
			return
		}
		t.Logf("refresh goroutine has not reconnected elap=%v timeout=%v, sleeping for %v", elap, timeout, sleep)
		time.Sleep(sleep)
	}
}

// go test -v ./refresh -run=TestCloseNotifyConnection
func TestCloseNotifyConnection(t *testing.T) {

	debug := true

	mock := &amqpMock{}

	o := Options{
		AmqpURL:      "amqp://guest:guest@rabbitmq:5672/",
		ConsumerTag:  "app",
		Applications: []string{"app"},
		Debug:        debug,
		AmqpClient:   mock,
	}

	r := New(o)

	if r == nil {
		t.Errorf("ugh")
	}

	wait := time.Second
	t.Logf("giving a time before close-notifying the amqp consumer goroutine: %v", wait)
	time.Sleep(wait)

	t.Logf("close-notifying the amqp consumer")

	begin := time.Now()

	timeout := 10 * time.Second
	sleep := 2 * time.Second

	// get non-zero initial connection count

	var connsInitial int

	for {
		connsInitial = r.statConnGet()
		t.Logf("connections count: initial=%d", connsInitial)
		if connsInitial > 0 {
			break
		}
		elap := time.Since(begin)
		if elap > timeout {
			t.Errorf("could not get non-zero initial connection count: timeout=%v", timeout)
			return
		}
		t.Logf("could not get non-zero initial connection count, sleeping for %v", sleep)
		time.Sleep(sleep)
	}

	// send notify close

	mock.connectionNotifyCloseSendMock(nil)

	// look for connection count increase

	for {
		conns := r.statConnGet()
		t.Logf("connections count: initial=%d current=%d", connsInitial, conns)
		if conns > connsInitial {
			break
		}
		elap := time.Since(begin)
		if elap > timeout {
			t.Errorf("refresh goroutine has never reconnected: timeout=%v", timeout)
			return
		}
		t.Logf("refresh goroutine has not reconnected elap=%v timeout=%v, sleeping for %v", elap, timeout, sleep)
		time.Sleep(sleep)
	}
}
