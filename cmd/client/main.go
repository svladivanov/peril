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

	ch, err := conn.Channel()
	if err != nil {
		log.Panicf("could not create channel: %v", err)
	}

	userName, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Panicf("could not get username: %v", err)
	}

	gs := gamelogic.NewGameState(userName)

	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+gs.GetUsername(),
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
		handlerPause(gs),
	)
	if err != nil {
		log.Panicf("could not subscribe to queue: %v", err)
	}

	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix+"."+gs.GetUsername(),
		routing.ArmyMovesPrefix+".*",
		pubsub.SimpleQueueTransient,
		handlerMove(gs),
	)
	if err != nil {
		log.Panicf("could not subscribe to army moves: %v", err)
	}

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
			move, err := gs.CommandMove(input)
			if err != nil {
				fmt.Println(err)
				continue
			}
			err = pubsub.PublishJSON(
				ch,
				routing.ExchangePerilTopic,
				routing.ArmyMovesPrefix+"."+move.Player.Username,
				move,
			)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Printf("Moved %d units to %s\n", len(move.Units), move.ToLocation)

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
