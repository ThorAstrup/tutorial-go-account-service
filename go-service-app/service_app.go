package goserviceapp

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/rabbitmq/amqp091-go"
)

// ServiceApp represents a service application.
type ServiceApp struct {
	Name          string                          // Name of the service application.
	AMQPUrl       string                          // URL of the AMQP server.
	EventHandlers map[string]func(string, []byte) // Map of event handlers.
}

// Start starts the service application.
// It establishes a connection to RabbitMQ, declares an exchange and a queue,
// binds the queue to the exchange for each configured event,
// registers a consumer to receive messages from the queue,
// and handles incoming messages by invoking the corresponding event handlers.
func (app *ServiceApp) Start() {
	log.Printf("Starting %s", app.Name)

	conn, err := amqp091.Dial(app.AMQPUrl)
	app.failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	app.failOnError(err, "Failed to open a channel")
	defer ch.Close()

	// Declare the exchange which the service-app will use to communicate with the queue with
	exchangeName := "account_topics"
	err = ch.ExchangeDeclare(
		exchangeName, // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	app.failOnError(err, "Failed to declare an exchange")

	q, err := ch.QueueDeclare(
		app.QueueName(), // name
		false,           // durable
		false,           // delete when unused
		true,            // exclusive
		false,           // no-wait
		nil,             // arguments
	)
	app.failOnError(err, "Failed to declare a queue")

	// Bind the queue to the exchange for each event configured
	for event := range app.EventHandlers {
		log.Printf("Binding queue \"%s\" to exchange \"%s\" with routing key \"%s\"", q.Name, exchangeName, event)
		err = ch.QueueBind(
			q.Name,       // queue name
			event,        // routing key
			exchangeName, // exchange
			false,
			nil,
		)
		app.failOnError(err, "Failed to bind a queue")
	}

	// Create the consumer
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto ack
		false,  // exclusive
		false,  // no local
		false,  // no wait
		nil,    // args
	)
	app.failOnError(err, "Failed to register a consumer")

	// Start a goroutine to handle incoming messages for this consumer
	go func() {
		for msg := range msgs {
			eventName := msg.RoutingKey
			eventHandlers := app.getEventHandlers(eventName)

			if len(eventHandlers) > 0 {
				for _, eventHandler := range eventHandlers {
					// Start a goroutine to handle the event
					go func(handler func(string, []byte), body []byte) {
						handler(eventName, body)
					}(eventHandler, msg.Body)
				}
			} else {
				log.Printf("Unknown event: %s", eventName)
			}
		}
	}()

	var forever chan struct{}
	log.Printf(" [*] Waiting for events. To exit press CTRL+C")
	<-forever
}

// Subscribe adds a new event subscription to the ServiceApp.
// It takes an event string and a handler function as parameters.
// The event string represents the name of the event to subscribe to.
// The handler function is a callback function that will be called when the event is triggered.
// The handler function takes two parameters: the event name as a string and the event data as a byte slice.
func (app *ServiceApp) Subscribe(event string, handler func(string, []byte)) {
	if app.EventHandlers == nil {
		app.EventHandlers = make(map[string]func(string, []byte))
	}

	app.EventHandlers[event] = handler
}

// getEventHandlers returns a slice of event handlers that match the given eventName.
// It searches for exact matches, wildcard matches with "*", and wildcard matches with "#".
// The returned slice contains functions of type func(string, []byte).
// If no matching event handlers are found, an empty slice is returned.
func (app *ServiceApp) getEventHandlers(eventName string) []func(string, []byte) {
	var handlers []func(string, []byte)
	if handler, found := app.EventHandlers[eventName]; found {
		handlers = append(handlers, handler)
	}

	// Check for event handlers with * wildcards that match the current "eventName"
	for handlerEvent, handler := range app.EventHandlers {
		// Check for event handlers with # wildcards that match the current "eventName"
		if strings.Contains(handlerEvent, "#") {
			if strings.HasPrefix(eventName, handlerEvent[:len(handlerEvent)-1]) {
				handlers = append(handlers, handler)
				continue
			}
		}

		// Check for event handlers with * wildcards that match the current "eventName"
		if strings.Contains(handlerEvent, "*") {
			pattern := strings.ReplaceAll(handlerEvent, ".", "\\.")
			pattern = strings.ReplaceAll(pattern, "*", ".*")
			pattern = "^" + pattern + "$"

			re, err := regexp.Compile(pattern)
			app.failOnError(err, fmt.Sprintf("Error compiling regex pattern: %v", err))

			if re.MatchString(eventName) {
				handlers = append(handlers, handler)
			}
		}
	}

	return handlers
}

// failOnError checks if an error occurred and panics with a formatted error message if it did.
// It takes an error and a message as parameters.
// If the error is not nil, it logs a panic message with the provided message and the error.
// This function is used in the ServiceApp struct to handle errors.
func (app *ServiceApp) failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

// QueueName returns the name of the queue associated with the ServiceApp.
// The queue name is generated by converting the app's name to lowercase and replacing spaces with underscores.
// Returns the queue name as a string.
func (app *ServiceApp) QueueName() string {
	queueName := strings.Replace(strings.ToLower(app.Name), " ", "_", -1)
	return queueName + "_queue"
}
