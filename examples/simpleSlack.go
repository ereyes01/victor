package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/FogCreek/victor"
	"github.com/FogCreek/victor/pkg/chat/slackRealtime"
	"github.com/FogCreek/victor/pkg/events"
	"github.com/FogCreek/victor/pkg/events/definedEvents"
)

const SLACK_TOKEN = "SLACK_TOKEN"
const BOT_NAME = "BOT_NAME"

func main() {

	defer func() {
		// this is only necessary since the slack api used by the slack adapter
		// does not currently implement a "Stop" or "Disconnect" method
		if e := recover(); e != nil {
			fmt.Println("bot.Stop() exited with panic: ", e)
			os.Exit(1)
		}
	}()

	bot := victor.New(victor.Config{
		ChatAdapter:   "slackRealtime",
		AdapterConfig: slackRealtime.NewConfig(SLACK_TOKEN),
		Name:          BOT_NAME,
	})
	addHandlers(bot)
	// optional help built in help command
	bot.EnableHelpCommand()
	bot.Run()
	go monitorErrors(bot.ChatErrors())
	go monitorEvents(bot.ChatEvents())
	// keep the process (and bot) alive
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	<-sigs

	bot.Stop()
}

func monitorErrors(errorChannel <-chan events.ErrorEvent) {
	for {
		err, ok := <-errorChannel
		if !ok {
			return
		}
		if err.IsFatal() {
			log.Panic(err.Error())
		}
		log.Println("Chat Adapter Error Event:", err.Error())
	}
}

func monitorEvents(eventsChannel chan events.ChatEvent) {
	for {
		event, ok := <-eventsChannel
		if !ok {
			return
		}
		switch e := event.(type) {
		case *definedEvents.ConnectingEvent:
			log.Println("Connecting Event fired")
		case *definedEvents.ConnectedEvent:
			log.Println("Connected Event fired")
		case *definedEvents.UserEvent:
			log.Printf("User Event: %+v", e)
		case *definedEvents.ChannelEvent:
			log.Printf("Channel Event: %+v", e)
		default:
			log.Println("Unrecognized Chat Event:", e)
		}
	}
}

func addHandlers(r victor.Robot) {
	// Add a typical command that will be displayed using the "help" command
	// if it is enabled.
	r.HandleCommand(&victor.HandlerDoc{
		CmdHandler:     byeFunc,
		CmdName:        "hi",
		CmdDescription: "Says goodbye when the user says hi!",
		CmdUsage:       []string{""},
	})
	// Add a hidden command that isn't displayed in the "help" command unless
	// mentioned by name
	r.HandleCommand(&victor.HandlerDoc{
		CmdHandler:     echoFunc,
		CmdName:        "echo",
		CmdDescription: "Hidden `echo` command!",
		CmdUsage:       []string{"", "`text to echo`"},
		CmdIsHidden:    true,
	})
	// Add a command to show the "Fields" method
	r.HandleCommand(&victor.HandlerDoc{
		CmdHandler:     fieldsFunc,
		CmdName:        "fields",
		CmdDescription: "Show the fields/parameters of a command message!",
		CmdUsage:       []string{"`param0` `param1` `...`"},
	})
	// Add a general pattern which is only checked on "non-command" messages
	// which are described in dispatch.go
	r.HandlePattern("\b(thanks|thank\\s+you)\b", thanksFunc)
	// Add default handler to show "unrecognized command" on "command" messages
	r.SetDefaultHandler(defaultFunc)
}

func byeFunc(s victor.State) {
	msg := fmt.Sprintf("Bye %s!", s.Message().User().Name())
	s.Chat().Send(s.Message().Channel().ID(), msg)
}

func echoFunc(s victor.State) {
	s.Chat().Send(s.Message().Channel().ID(), s.Message().Text())
}

func thanksFunc(s victor.State) {
	msg := fmt.Sprintf("You're welcome %s!", s.Message().User().Name())
	s.Chat().Send(s.Message().Channel().ID(), msg)
}

func fieldsFunc(s victor.State) {
	var fStr string
	for _, f := range s.Fields() {
		fStr += f + "\n"
	}
	s.Chat().Send(s.Message().Channel().ID(), fStr)
}

func defaultFunc(s victor.State) {
	s.Chat().Send(s.Message().Channel().ID(),
		"Unrecognized command. Type `help` to see supported commands.")
}
