package main

import (
	"fmt"
	"github.com/nlopes/slack"
	"io/ioutil"
	"strings"
	//"bytes"
	//"strconv"
	"strconv"
)

var api = slack.New("")

var gifs = make(map[string]string)
var sanity = make(map[string]int)

func panicCheck(e error) {
	if e != nil {
		panic(e)
	}
}

func warningCheck(e error) {
	if e != nil {
		fmt.Printf("Warning : %s\n", e)
	}
}

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
	fmt.Printf("gifs :\n%s\n", gifs)
}

func initSanity() {
	users, e := api.GetUsers()
	if e != nil {
		warningCheck(e)
		return
	}
	for i := 0; i < len(users); i++ {
		if users[i].IsBot == false && users[i].Deleted == false {
			sanity[users[i].Name] = 10
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
	if getUsername(message.User) == "billcipher" {
		return
	}
	text := message.Text
	if strings.Contains(message.Text, "<@U5EFD52R0>") == true || message.Channel[0] == 'D' {
		text = strings.Replace(text, "<@U5EFD52R0>", "", 1)
		text = strings.TrimSpace(text)
		words := strings.Split(text, "=")
		fmt.Println("Recognize command mode")
		if len(words) > 0 {
			for k, v := range(commands) {
				if strings.Contains(text, k) {
					go v.(func(*slack.MessageEvent))(message)
					return
				}
			}
		}
	}
	for meme := range gifs {
		go messageEventGifPoster(text, meme, message.Channel)
	}
}

func getUsername(id string) string {
	users, e := api.GetUsers()
	if e != nil {
		warningCheck(e)
		return ""
	}
	for i := 0; i < len(users); i++ {
		if users[i].ID == id {
			return users[i].Name
		}
	}
	return ""
}

func helpCommand(message *slack.MessageEvent) {
	params := slack.PostMessageParameters{}
	params.AsUser = true
	params.UnfurlLinks = true
	params.UnfurlMedia = true
	api.PostMessage(message.Channel, "https://media.giphy.com/media/WfE3yNXrkMAAo/giphy.gif?response_id=5921b583a79f5c11408fde74", params)
	username := getUsername(message.User)
	if username != "" {
		api.PostMessage(message.Channel, "-1 en sanité pour " + username, params)
		if user, ok := sanity[username]; ok {
			sanity[username] = user - 1
		} else {
			sanity[username] = 9
		}
	}
}

func sanityCommand(message *slack.MessageEvent) {
	text := "Sanity board :\n"
	for user, points := range(sanity) {
		text = text + user + " : " + strconv.Itoa(points) + "\n"
	}
	simpleMessagePoster(message.Channel, text)
}

var commands = map[string]interface{}{
	"help": helpCommand,
	"aide-moi": helpCommand,
	"aidez-moi": helpCommand,
	"aide moi": helpCommand,
	"aidez moi": helpCommand,
	"a l'aide": helpCommand,
	"à l'aide": helpCommand,
	"a laide": helpCommand,
	"sanity": sanityCommand,
	"sanite": sanityCommand,
	"sanité": sanityCommand,
}

func main() {
	api.SetDebug(true)
	_, err := api.AuthTest()
	panicCheck(err)
	go initGifs()
	go initSanity()
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
