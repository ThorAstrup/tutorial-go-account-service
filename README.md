# Account Service Application Example

Example of a account service app, made in go, which can subscribe to one or more AMQP events.

 ## How this app was initialized

```bash
$ mkdir tutorial-go-account-service && cd tutorial-go-account-service
$ go mod init github.com/{username}/tutorial-go-account-service
$ go get github.com/rabbitmq/amqp091-go
$ code main.go
```

 ## RabbitMQ docker container for development

 ```bash
 docker run -d \
    -p 5672:5672 \
    -p 15672:15672 \
    -e RABBITMQ_DEFAULT_USER=guest \
    -e RABBITMQ_DEFAULT_PASS=guest \
    rabbitmq:3-management
 ```

 ```yml
services:
    rabbitmq:
        image: rabbitmq:3-management
        ports:
            - "5672:5672"
            - "15672:15672"
        environment:
            RABBITMQ_DEFAULT_USER: "guest"
            RABBITMQ_DEFAULT_PASS: "guest"
 ```

 Admin interface: http://localhost:15672