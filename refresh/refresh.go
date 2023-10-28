package refresh

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/segmentio/ksuid"
)

// Refresh holds data for worker that reads events from amqp and delivers applications found in channel `Refresh.C`.
type Refresh struct {
	amqpURL           string
	applications      []string
	consumerTag       string
	debug             bool
	dialTimeout       time.Duration
	dialRetryInterval time.Duration
	C                 chan string // Channel C delivers applications names found in events.
	done              chan struct{}
	amqpClient        amqpEngine
	matchAppFunc      func(notification, application string) bool

	lock   sync.Mutex
	closed bool
	exited bool

	statCountConnections int
}

// DefaultMatchApplication checks whether a notification matches an application name.
func DefaultMatchApplication(notification, application string) bool {
	if application == "#" {
		return true
	}
	notification = strings.TrimSuffix(notification, ":**")
	notification = strings.Replace(notification, ":", "-", 1)
	return notification == application
}

// New spawns a Refresh worker that reads events from amqp and delivers applications found in channel `Refresh.C`.
func New(amqpURL, consumerTag string, applications []string, debug bool, engine amqpEngine) *Refresh {
	if len(applications) == 0 {
		log.Panicln("refresh.New: slice applications must be non-empty")
	}
	r := &Refresh{
		amqpURL:           amqpURL,
		applications:      applications,
		debug:             debug,
		C:                 make(chan string),
		dialTimeout:       10 * time.Second,
		dialRetryInterval: 5 * time.Second,
		consumerTag:       consumerTag,
		done:              make(chan struct{}),
		amqpClient:        engine,
	}

	if r.amqpClient == nil {
		r.amqpClient = &amqpReal{}
	}

	if r.matchAppFunc == nil {
		r.matchAppFunc = DefaultMatchApplication
	}

	go serve(r)
	return r
}

// Close interrupts the refresh goroutine to release resources.
func (r *Refresh) Close() {
	r.lock.Lock()
	defer r.lock.Unlock()
	if r.closed {
		return
	}
	close(r.done)
	r.closed = true
}

func (r *Refresh) isClosed() bool {
	r.lock.Lock()
	defer r.lock.Unlock()
	return r.closed
}

func (r *Refresh) statConnInc() int {
	r.lock.Lock()
	r.statCountConnections++
	count := r.statCountConnections
	r.lock.Unlock()
	return count
}

func (r *Refresh) statConnGet() int {
	r.lock.Lock()
	count := r.statCountConnections
	r.lock.Unlock()
	return count
}

func serve(r *Refresh) {
	// serve forever, unless interrupted by Close()
	for i := r.statConnInc(); !r.isClosed(); i = r.statConnInc() {
		begin := time.Now()
		serveOnce(r, i, begin)
		log.Printf("serve connCount=%d uptime=%v: will restart amqp connection", i, time.Since(begin))
	}
	if r.debug {
		log.Print("refresh.serve: refresh closed, exiting...")
	}

	close(r.C) // close delivery channel to notify readers we finished

	r.lock.Lock()
	r.exited = true
	r.lock.Unlock()
	if r.debug {
		log.Print("refresh.serve: refresh closed, exiting...done")
	}
}

// hasExited is used for testing to check the goroutine has exited.
func (r *Refresh) hasExited() bool {
	r.lock.Lock()
	defer r.lock.Unlock()
	return r.exited
}

func newUID() string {
	const cooldown = 5 * time.Second
	for {
		k, err := ksuid.NewRandom()
		if err != nil {
			log.Printf("newUID: ksuid: %v, sleeping for %v", err, cooldown)
			time.Sleep(cooldown)
			continue
		}
		return k.String()
	}
}

// serve one amqp connection
func serveOnce(r *Refresh, connCount int, begin time.Time) {

	const me = "refresh.serveOne"

	const queuePrefix = "config-event-queue"

	exchangeName := "springCloudBus"
	exchangeType := "topic"

	queue := queuePrefix + "." + newUID()

	log.Printf("%s: connection count:    %d", me, connCount)
	log.Printf("%s: amqp URL:            %s", me, r.amqpURL)
	log.Printf("%s: exchangeName:        %s", me, exchangeName)
	log.Printf("%s: exchangeType:        %s", me, exchangeType)
	log.Printf("%s: queue:               %s", me, queue)
	log.Printf("%s: consumerTag:         %s", me, r.consumerTag)
	log.Printf("%s: applications:        %v", me, r.applications)
	log.Printf("%s: dial timeout:        %v", me, r.dialTimeout)
	log.Printf("%s: dial retry interval: %v", me, r.dialRetryInterval)
	log.Printf("%s: debug:               %t", me, r.debug)

	conn := r.amqpClient.dial(r.isClosed, r.amqpURL, r.dialRetryInterval, r.dialTimeout)
	if conn == nil {
		log.Printf("%s: dial failed, refresh must have been closed by caller", me)
		return
	}
	defer r.amqpClient.closeConn(conn)

	connNotifyClose := make(chan *amqp.Error, 1)
	r.amqpClient.connectionNotifyClose(conn, connNotifyClose)

	ch, err := r.amqpClient.channel(conn)
	if err != nil {
		log.Printf("%s: failed to open channel: %v", me, err)
		return
	}
	defer r.amqpClient.closeChannel(ch)

	chanNotifyClose := make(chan *amqp.Error, 1)
	r.amqpClient.channelNotifyClose(ch, chanNotifyClose)

	{
		log.Printf("%s: got channel, declaring exchange: exchangeName=%s exchangeType=%s", me, exchangeName, exchangeType)
		err := r.amqpClient.exchangeDeclare(ch, exchangeName, exchangeType)
		if err != nil {
			log.Printf("%s: failed to declare exchange: %v", me, err)
			return
		}
	}

	log.Printf("%s: declared exchange, declaring queue: %s", me, queue)
	q, err := r.amqpClient.queueDeclare(ch, queue)
	if err != nil {
		log.Printf("%s: failed to declare queue: %v", me, err)
		return
	}

	{
		const routingKey = "#"
		log.Printf("%s: declared queue (%d messages, %d consumers), binding to exchange '%s' with routing key '%s'",
			me, q.Messages, q.Consumers, exchangeName, routingKey)
		err := r.amqpClient.queueBind(ch, q.Name, routingKey, exchangeName)
		if err != nil {
			log.Printf("%s: failed to bind queue to exchange: %v", me, err)
			return
		}
	}

	if r.debug {
		log.Printf("DEBUG %s: entering consume loop", me)
	}

	msgs, err := r.amqpClient.consume(ch, q.Name, r.consumerTag)
	if err != nil {
		log.Printf("%s: conn=%d uptime=%v: failed to register consumer: %v",
			me, connCount, time.Since(begin), err)
		return
	}
	defer r.amqpClient.cancel(ch, r.consumerTag)

	if r.debug {
		log.Printf("DEBUG %s: entering delivery loop", me)
	}

	var msg int

	for {
		select {
		case d, ok := <-msgs:
			if !ok {
				log.Printf("%s: conn=%d uptime=%v: consume channel closed", me, connCount, time.Since(begin))
				return
			}
			msg++
			if r.debug {
				log.Printf(
					"DEBUG %s: conn=%d uptime=%v msg=%d: ConsumerTag=[%s] DeliveryTag=[%v] RoutingKey=[%s] ContentType=[%s] Body='%s'",
					me,
					connCount,
					time.Since(begin),
					msg,
					d.ConsumerTag,
					d.DeliveryTag,
					d.RoutingKey,
					d.ContentType,
					d.Body)
			}
			errClose := handleDelivery(r.matchAppFunc, r.isClosed, d.Body, r.applications,
				r.C, r.done,
				chanNotifyClose,
				connNotifyClose,
				r.debug)
			if errClose != nil {
				log.Printf("%s: NotifyClose: handleDelivery: %v", me, errClose)
				return
			}
		case <-r.done:
			log.Printf("%s: done channel has been closed", me)
			return
		case err := <-chanNotifyClose:
			log.Printf("%s: NotifyClose: channel: %v", me, err)
			return
		case err := <-connNotifyClose:
			log.Printf("%s: NotifyClose: connection: %v", me, err)
			return
		}
	}

}

/*
	{
	    "type": "RefreshRemoteApplicationEvent",
	    "timestamp": 1649804650957,
	    "originService": "config-server:0:0a36277496365ee8621ae8f3ce7032ce",
	    "destinationService": "config-cli-example:**",
	    "id": "5a4cb501-652a-4ae2-9d3e-279e1d2a2169"
	}
*/
func handleDelivery(matchAppFunc func(notification, app string) bool,
	isClosed func() bool, body []byte, applications []string,
	ch chan<- string, done <-chan struct{},
	chanNotifyClose <-chan *amqp.Error,
	connNotifyClose <-chan *amqp.Error,
	debug bool) error {

	const me = "handleDelivery"

	event := map[string]interface{}{}

	err := json.Unmarshal(body, &event)
	if err != nil {
		log.Printf("%s: body json error: %v", me, err)
		return nil
	}

	et := event["type"]
	eventType, typeIsStr := et.(string)
	if !typeIsStr {
		log.Printf("%s: 'type' is not a string: type=%[2]T value=%[2]v", me, et)
		return nil
	}

	if eventType != "RefreshRemoteApplicationEvent" {
		if debug {
			log.Printf("DEBUG %s: ignoring event type=[%s]", me, eventType)
		}
		return nil
	}

	destinationService := event["destinationService"]
	dst, isStr := destinationService.(string)
	if !isStr {
		log.Printf("%s: 'destinationService' is not a string: type=%[2]T value=%[2]v", me, destinationService)
		return nil
	}

	// find if any application matches the service
	for _, app := range applications {
		if isClosed() {
			return fmt.Errorf("%s: isClosed()=true", me)
		}
		matched := matchAppFunc(dst, app)
		if debug {
			log.Printf("DEBUG %s: destinationService='%s' <=> application='%s' matched=%t", me, dst, app, matched)
		}
		if !matched {
			continue
		}

		// match found

		select {
		case ch <- app:
			return nil // clean return
		case <-done:
			return fmt.Errorf("%s: done channel has been closed", me)
		case err := <-chanNotifyClose:
			return fmt.Errorf("%s: NotifyClose: channel: %v", me, err)
		case err := <-connNotifyClose:
			return fmt.Errorf("%s: NotifyClose: connection: %v", me, err)
		}
	}

	return nil // not reached
}
