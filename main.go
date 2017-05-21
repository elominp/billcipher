package main

import (
	"fmt"
	"github.com/nlopes/slack"
	"io/ioutil"
	"strings"
	//"bytes"
)

var api = slack.New("xoxb-184523172850-7nbbBBEc0gvjSXaBFVq7Fqtk")

var gifs = make(map[string]string)

func panicCheck(e error) {
	if e != nil {
		panic(e)
	}
}

/* func warningCheck(e error) {
	if e != nil {
		fmt.Printf("Warning : %s\n", e)
	}
} */

func splitGifLine(line string, gifTuple chan []string) {
	gifTuple <- strings.Split(line, "=")
}

func initGifs() {
	data, e := ioutil.ReadFile("gifs.txt")
	panicCheck(e)
	lines := strings.Split(string(data), "\n")
	linesLen := len(lines)
	gifsTuples := make(chan []string, linesLen)
	for i := 0; i < len(lines); i++ {
		go splitGifLine(lines[i], gifsTuples)
	}
	for i := 0; i < len(lines); i++ {
		tuple := <- gifsTuples
		if len(tuple) > 1 {
			gifs[tuple[0]] = tuple[1]
		}
	}
}

func simpleMessagePoster(channel string, message string) {
	params := slack.PostMessageParameters{}
	params.AsUser = true
	api.PostMessage(channel, message, params)
}

func joinedChannel(channel string) {
	simpleMessagePoster(channel, "Hello there !")
}

func messageEventGifPoster(message string, meme string, channel string) {
	if strings.Contains(message, meme) {
		params := slack.PostMessageParameters{}
		params.AsUser = true
		params.UnfurlLinks = true
		params.UnfurlMedia = true
		api.PostMessage(channel, gifs[meme], params)
	}
}

func messageEvent(message *slack.MessageEvent) {
	for meme := range gifs {
		go messageEventGifPoster(message.Text, meme, message.Channel)
	}
}

func main() {
	api.SetDebug(true)
	_, err := api.AuthTest()
	panicCheck(err)
	initGifs()
	fmt.Printf("gifs :\n%s\n", gifs)
	rtm := api.NewRTM()
	go rtm.ManageConnection()
	for msg := range rtm.IncomingEvents {
		fmt.Printf("Event received : %s\n", msg.Type)
		switch event := msg.Data.(type) {
		case *slack.ChannelJoinedEvent:
			go joinedChannel(event.Channel.Name)
		case *slack.MessageEvent:
			go messageEvent(event)
		default:
			fmt.Println("Not handled")
		}
	}
}