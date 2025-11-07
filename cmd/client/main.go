package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril client...")
	const rabbitConnURL = "amqp://guest:guest@localhost:5672/"

	conn, err := amqp.Dial(rabbitConnURL)
	if err != nil {
		log.Panicf("could not connect to RabbitMQ: %v", err)
	}
	defer conn.Close()
	fmt.Println("Peril game client connected to RabbitMQ!")

	userName, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Panicf("could not get username: %v", err)
	}

	queueName := fmt.Sprintf("%s.%s", routing.PauseKey, userName)
	_, queue, err := pubsub.DeclareAndBind(conn, routing.ExchangePerilDirect, queueName, routing.PauseKey, pubsub.SimpleQueueTransient)
	if err != nil {
		log.Panicf("could not declare and bind queue: %v", err)
	}
	fmt.Printf("Queue %s declared and bound!\n", queue.Name)

	gs := gamelogic.NewGameState(userName)

	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			gamelogic.PrintClientHelp()
			continue
		}

		switch input[0] {
		case "spawn":
			err = gs.CommandSpawn(input)
			if err != nil {
				fmt.Println(err)
				continue
			}
		case "move":
			_, err = gs.CommandMove(input)
			if err != nil {
				fmt.Println(err)
				continue
			}
		case "status":
			gs.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Printf("Unrecognized command: %s\n", input[0])
		}
	}
}
