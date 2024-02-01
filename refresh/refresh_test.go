package refresh

import (
	"sync"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func TestCloseRefresh(t *testing.T) {

	debug := true

	o := Options{
		AmqpURL:      "amqp://guest:guest@rabbitmq:5672/",
		ConsumerTag:  "app",
		Applications: []string{"app"},
		Debug:        debug,
	}

	r := New(o)

	wait := time.Second
	t.Logf("giving a time before interrupting the refresh goroutine: %v", wait)
	time.Sleep(wait)

	t.Logf("closing refresh")
	r.Close()

	begin := time.Now()

	timeout := 20 * time.Second

	for !r.hasExited() {
		elap := time.Since(begin)
		if elap > timeout {
			t.Errorf("refresh goroutine has never exited: timeout=%v", timeout)
			return
		}
		sleep := 2 * time.Second
		t.Logf("refresh goroutine has not exited elap=%v timeout=%v, sleeping for %v", elap, timeout, sleep)
		time.Sleep(sleep)
	}
}

func TestCloseDelivery(t *testing.T) {

	var lock sync.Mutex
	var closed bool
	var exited bool

	ch := make(chan string)
	done := make(chan struct{})

	isClosed := func() bool {
		lock.Lock()
		defer lock.Unlock()
		return closed
	}

	doClose := func() {
		lock.Lock()
		defer lock.Unlock()
		close(done)
		closed = true
	}

	hasExited := func() bool {
		lock.Lock()
		defer lock.Unlock()
		return exited
	}

	exit := func() {
		lock.Lock()
		defer lock.Unlock()
		exited = true
	}

	body := []byte(`{"type":"RefreshRemoteApplicationEvent","destinationService":"app:"}`)

	debug := true

	go func() {
		t.Logf("delivery goroutine started")
		handleDelivery(DefaultMatchApplication, isClosed, body, []string{"app"}, ch, done,
			nil,
			nil,
			debug)
		exit()
	}()

	wait := time.Second
	t.Logf("giving a time before interrupting delivery goroutine: %v", wait)
	time.Sleep(wait)

	t.Logf("interrupting delivery goroutine")
	doClose()

	begin := time.Now()

	timeout := 20 * time.Second

	for !hasExited() {
		if time.Since(begin) > timeout {
			t.Errorf("delivery goroutine has never exited")
			return
		}
		sleep := 2 * time.Second
		t.Logf("delivery goroutine has not exited, sleeping for %v", sleep)
		time.Sleep(sleep)
	}
}

// go test -v ./refresh -run=TestNotifyCloseChanDelivery
func TestNotifyCloseChanDelivery(t *testing.T) {

	var lock sync.Mutex
	var exited bool

	ch := make(chan string)
	done := make(chan struct{})

	isClosed := func() bool {
		return false
	}

	hasExited := func() bool {
		lock.Lock()
		defer lock.Unlock()
		return exited
	}

	exit := func() {
		lock.Lock()
		defer lock.Unlock()
		exited = true
	}

	body := []byte(`{"type":"RefreshRemoteApplicationEvent","destinationService":"app:"}`)

	debug := true

	chanNotifyClose := make(chan *amqp.Error, 1)

	go func() {
		t.Logf("delivery goroutine started")
		errClose := handleDelivery(DefaultMatchApplication, isClosed, body, []string{"app"}, ch, done,
			chanNotifyClose,
			nil,
			debug)
		t.Logf("handleDelivery: notifyClose: %v", errClose)
		exit()
	}()

	wait := time.Second
	t.Logf("giving a time before sending notify closed channel: %v", wait)
	time.Sleep(wait)

	t.Logf("sending notify closed channel")
	chanNotifyClose <- nil

	begin := time.Now()

	timeout := 10 * time.Second

	for !hasExited() {
		if time.Since(begin) > timeout {
			t.Errorf("delivery goroutine has never exited")
			return
		}
		sleep := 2 * time.Second
		t.Logf("delivery goroutine has not exited, sleeping for %v", sleep)
		time.Sleep(sleep)
	}
}

// go test -v ./refresh -run=TestNotifyCloseConnDelivery
func TestNotifyCloseConnDelivery(t *testing.T) {

	var lock sync.Mutex
	var exited bool

	ch := make(chan string)
	done := make(chan struct{})

	isClosed := func() bool {
		return false
	}

	hasExited := func() bool {
		lock.Lock()
		defer lock.Unlock()
		return exited
	}

	exit := func() {
		lock.Lock()
		defer lock.Unlock()
		exited = true
	}

	body := []byte(`{"type":"RefreshRemoteApplicationEvent","destinationService":"app:"}`)

	debug := true

	connNotifyClose := make(chan *amqp.Error, 1)

	go func() {
		t.Logf("delivery goroutine started")
		errClose := handleDelivery(DefaultMatchApplication, isClosed, body, []string{"app"}, ch, done,
			nil,
			connNotifyClose,
			debug)
		t.Logf("handleDelivery: notifyClose: %v", errClose)
		exit()
	}()

	wait := time.Second
	t.Logf("giving a time before sending notify closed conn: %v", wait)
	time.Sleep(wait)

	t.Logf("sending notify closed conn")
	connNotifyClose <- nil

	begin := time.Now()

	timeout := 10 * time.Second

	for !hasExited() {
		if time.Since(begin) > timeout {
			t.Errorf("delivery goroutine has never exited")
			return
		}
		sleep := 2 * time.Second
		t.Logf("delivery goroutine has not exited, sleeping for %v", sleep)
		time.Sleep(sleep)
	}
}

// go test -v ./refresh -run=TestCloseConsume
func TestCloseConsume(t *testing.T) {

	debug := true

	o := Options{
		AmqpURL:      "amqp://guest:guest@rabbitmq:5672/",
		ConsumerTag:  "app",
		Applications: []string{"app"},
		Debug:        debug,
		AmqpClient:   &amqpMock{},
	}

	r := New(o)

	if r == nil {
		t.Errorf("ugh")
	}

	wait := time.Second
	t.Logf("giving a time before interrupting the refresh goroutine: %v", wait)
	time.Sleep(wait)

	t.Logf("closing refresh")
	r.Close()

	begin := time.Now()

	timeout := 20 * time.Second

	for !r.hasExited() {
		elap := time.Since(begin)
		if elap > timeout {
			t.Errorf("refresh goroutine has never exited: timeout=%v", timeout)
			return
		}
		sleep := 2 * time.Second
		t.Logf("refresh goroutine has not exited elap=%v timeout=%v, sleeping for %v", elap, timeout, sleep)
		time.Sleep(sleep)
	}
}

// go test -v ./refresh -run=TestCloseConsumeWithChannel
func TestCloseConsumeWithChannel(t *testing.T) {

	debug := true

	o := Options{
		AmqpURL:      "amqp://guest:guest@rabbitmq:5672/",
		ConsumerTag:  "app",
		Applications: []string{"app"},
		Debug:        debug,
		AmqpClient:   &amqpMock{},
	}

	r := New(o)

	if r == nil {
		t.Errorf("ugh")
	}

	wait := time.Second
	t.Logf("giving a time before interrupting the refresh goroutine: %v", wait)
	time.Sleep(wait)

	t.Logf("closing refresh")
	r.Close()

	timeout := time.NewTimer(20 * time.Second)
LOOP:
	for {
		select {
		case _, ok := <-r.C:
			if !ok {
				t.Logf("refresh gorouting exited!")
				break LOOP
			}
		case <-timeout.C:
			t.Errorf("refresh goroutine has never exited: timeout=%v", timeout)
			break LOOP
		}
	}
}

type amqpMock struct {
	chanNotifyClose []chan *amqp.Error
	connNotifyClose []chan *amqp.Error
	lock            sync.Mutex
}

func (a *amqpMock) dial(isClosed func() bool, amqpURL string, sleep, timeout time.Duration) *amqp.Connection {
	return &amqp.Connection{}
}

func (a *amqpMock) closeConn(conn *amqp.Connection) error {
	return nil
}

func (a *amqpMock) channel(conn *amqp.Connection) (*amqp.Channel, error) {
	return &amqp.Channel{}, nil
}

func (a *amqpMock) closeChannel(ch *amqp.Channel) error {
	return nil
}
func (a *amqpMock) cancel(ch *amqp.Channel, consumerTag string) error {
	return nil
}

func (a *amqpMock) exchangeDeclare(ch *amqp.Channel, exchangeName, exchangeType string) error {
	return nil
}

func (a *amqpMock) queueDeclare(ch *amqp.Channel, queueName string) (amqp.Queue, error) {
	return amqp.Queue{}, nil
}

func (a *amqpMock) queueBind(ch *amqp.Channel, queueName, routingKey, exchangeName string) error {
	return nil
}

func (a *amqpMock) consume(ch *amqp.Channel, queueName, consumerTag string) (<-chan amqp.Delivery, error) {
	return make(<-chan amqp.Delivery), nil // reading this will block forever
}

func (a *amqpMock) channelNotifyClose(ch *amqp.Channel, receiver chan *amqp.Error) chan *amqp.Error {
	a.lock.Lock()
	a.chanNotifyClose = append(a.chanNotifyClose, receiver)
	a.lock.Unlock()
	return nil
}

func (a *amqpMock) connectionNotifyClose(conn *amqp.Connection, receiver chan *amqp.Error) chan *amqp.Error {
	a.lock.Lock()
	a.connNotifyClose = append(a.connNotifyClose, receiver)
	a.lock.Unlock()
	return nil
}

func (a *amqpMock) channelNotifyCloseSendMock(err *amqp.Error) {
	a.lock.Lock()
	for _, ch := range a.chanNotifyClose {
		ch <- err
	}
	a.lock.Unlock()
}

func (a *amqpMock) connectionNotifyCloseSendMock(err *amqp.Error) {
	a.lock.Lock()
	for _, ch := range a.connNotifyClose {
		ch <- err
	}
	a.lock.Unlock()
}
