package main

import (
	"log"

	goserviceapp "github.com/thorastrup/tutorial-go-account-service/go-service-app"
)

func main() {
	app := goserviceapp.ServiceApp{
		Name:    "Account Service",
		AMQPUrl: "amqp://guest:guest@localhost:5672/",
	}

	app.Subscribe("account.#", func(event string, body []byte) {
		log.Printf("[account.#]: %s", body)
	})

	app.Subscribe("account.created", func(event string, body []byte) {
		log.Printf("[account.created]: %s", body)
	})

	app.Subscribe("account.*.test", func(event string, body []byte) {
		log.Printf("[account.*.test]: %s", body)
	})

	app.Start()
}
