package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	const rabbitConnURL = "amqp://guest:guest@localhost:5672/"

	conn, err := amqp.Dial(rabbitConnURL)
	if err != nil {
		log.Panicf("could not connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	fmt.Println("Peril game server connected to RabbitMQ!")

	// wait for interrupt signal to exit
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Println("RabbitMQ connection closed.")
}
