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

func joinedChannel(channel string) {

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
			go joinedChannel(event.Channel)
		}
	}
}