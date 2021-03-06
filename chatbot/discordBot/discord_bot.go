package discordBot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Variables used for command line parameters
var (
	token = os.Getenv("TOKEN")
	dg    *discordgo.Session
)

// RunBot starts the bot
func RunBot() {
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}
	args := strings.Split(m.Content, "/")

	fmt.Println(len(args))
	if strings.HasPrefix(m.Content, "!user/") && len(args) == 2 {
		getUserInfo(s, m)
		return
	} else if strings.HasPrefix(m.Content, "!user/") && len(args) == 4 {
		getUserPlaylists(s, m, args)
		return
	} else if strings.HasPrefix(m.Content, "!volume") {
		volumeControl(s, m)
		return
	}

	switch strings.ToLower(m.Content) {
	case "greeting", "hi", "hello", "hey", "hei", "hola", "yo", "wassup", "sup", "halla":
		greeting(s, m)
	case "!author", "whoami", "myname":
		author(s, m)
	case "!calculate":
		calculate(s, m)
	case "!ping", "!pong":
		pingPong(s, m)
	case "!me":
		meSpotify(s, m)
	case "!api", "!spotify":
		getAPIInfo(s, m)
	case "!norris", "!chuck":
		getNorrisJokes(s, m)
	case "!trump", "!tjokes":
		getTrumpJoke(s, m)
	case "!current song", "!listening on?", "!which song", "song":
		getCurrentSong(s, m)
	case "!devices", "!devs", "!enheter":
		getDevices(s, m)
	case "!next", "!next song", "!play next":
		playNext(s, m)
	case "!back", "!previous song", "!play back":
		playPrevious(s, m)
	case "!my playlists", "!my albums", "!all my albums", "!albums":
		usersPlaylists(s, m)
	case "!pause", "!stop":
		pauseSong(s, m)
	case "!play", "!continue":
		resumeSong(s, m)
	case "!repeat":
		repeat(s, m)
	case "!været", "!weather", "!hva er temperaturen nå", "!hva er tempen nå", "!temp":
		weatherNow(s, m)
	default:
		fmt.Println(m.Content)
		s.ChannelMessageSend(m.ChannelID, "Say what?")
	}
}

//
func weatherNow(s *discordgo.Session, m *discordgo.MessageCreate) {
	temp, err := GetTemp()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error")
		return
	}
	s.ChannelMessageSend(m.ChannelID, temp)
}

// Sets repeat mode on
func volumeControl(s *discordgo.Session, m *discordgo.MessageCreate) {
	parts := strings.Split(m.Content, " ")
	fmt.Println(parts)
	url := "http://spotify_api:5001/volume/" + parts[1]

	_, err := http.Get(url)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error setting the volume level!")
		fmt.Println(err)
		return
	}
	s.ChannelMessageSend(m.ChannelID, "Volume adjusted :musical_note:")
}

// Sets repeat mode on
func repeat(s *discordgo.Session, m *discordgo.MessageCreate) {
	url := "http://spotify_api:5001/repeat"

	_, err := http.Get(url)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error setting on repeat mode!")
		fmt.Println(err)
		return
	}
	s.ChannelMessageSend(m.ChannelID, "Repeat mode on :musical_note:")
}

// Resumes the song
func resumeSong(s *discordgo.Session, m *discordgo.MessageCreate) {
	url := "http://spotify_api:5001/play"

	_, err := http.Get(url)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error resuming song!")
		fmt.Println(err)
		return
	}
	s.ChannelMessageSend(m.ChannelID, "Continue listening :musical_note:")
}

// Pauses the current playing song
func pauseSong(s *discordgo.Session, m *discordgo.MessageCreate) {
	url := "http://spotify_api:5001/pause"

	_, err := http.Get(url)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error pausing song!")
		fmt.Println(err)
		return
	}
	s.ChannelMessageSend(m.ChannelID, "Paused")
}

// Gets user's playlists
func usersPlaylists(s *discordgo.Session, m *discordgo.MessageCreate) {
	url := "http://spotify_api:5001/me/playlists"

	resp, err := http.Get(url)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error getting playlists!")
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error reading response body!")
		return
	}
	res := string(body)
	//fmt.Println(res)
	parts := strings.Split(res, "\"%\",")
	for _, each := range parts {
		fmt.Println(each)
		_, err = s.ChannelMessageSend(m.ChannelID, each)
		fmt.Println(err)
	}

}

// Plays previous song
func playPrevious(s *discordgo.Session, m *discordgo.MessageCreate) {
	values := map[string]string{"change": "song"}
	jsonValue, _ := json.Marshal(values)
	resp, err := http.Post("http://spotify_api:5001/previous", "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Oppps, error making request")
		return
	}

	fmt.Println(resp.StatusCode)
	if resp.StatusCode != 204 {
		s.ChannelMessageSend(m.ChannelID, "Oppps, got unexpected error code")
		return
	}
	s.ChannelMessageSend(m.ChannelID, "Enjoy :musical_note:")
}

// Plays next song
func playNext(s *discordgo.Session, m *discordgo.MessageCreate) {
	values := map[string]string{"change": "song"}
	jsonValue, _ := json.Marshal(values)
	resp, err := http.Post("http://spotify_api:5001/next", "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Oppps, error making request")
		return
	}

	fmt.Println(resp.StatusCode)
	if resp.StatusCode != 204 {
		s.ChannelMessageSend(m.ChannelID, "Oppps, got unexpected error code")
		return
	}
	s.ChannelMessageSend(m.ChannelID, "Enjoy :musical_note:")
}

// Gets all devices you are listening spotify on.
func getDevices(s *discordgo.Session, m *discordgo.MessageCreate) {
	resp, err := http.Get("http://spotify_api:5001/devices")
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error from the api!")
		return
	}

	var devices Response
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&devices); err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error parsing into json xP")
		return
	}
	fmt.Println(devices)
	var reply string
	code := strconv.Itoa(devices.Response.StatusCode)
	if devices.Response.StatusCode != 200 {
		reply = "Response code not 200, got response code" + code
		s.ChannelMessageSend(m.ChannelID, reply)
		return
	}

	count := strconv.Itoa(len(devices.Response.Devices))
	reply = "Nr of devices: " + count + "\n"
	for k, v := range devices.Response.Devices {
		index := strconv.Itoa(k + 1)
		reply += "\nDevice_nr: " + index
		if v.IsActive {
			reply += "\nIsActice: " + "True"
		} else {
			reply += "\nIsActice: " + "False"
		}
		reply += "\nName: " + v.Name
		reply += "\nType: " + v.Type
		vPercent := strconv.Itoa(v.VolumePercent)
		reply += "\nVolumePercent: " + vPercent + "\n"
	}

	s.ChannelMessageSend(m.ChannelID, reply)
}

// Gets current playing spotify song
func getCurrentSong(s *discordgo.Session, m *discordgo.MessageCreate) {
	resp, err := http.Get("http://spotify_api:5001/current_song")
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error from the api!")
		return
	}

	var song CurrentSongWrapper
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&song); err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error parsing into json xP")
		return
	}

	var reply string
	if song.Response.Error {
		reply = "Response code not 200, got response code" + string(song.Response.ErrorCode)
		s.ChannelMessageSend(m.ChannelID, reply)
		return
	}
	reply = "You are currently listening to: " + song.Response.Name + "\nArtists: "
	for _, name := range song.Response.Artists {
		reply += name + " "
	}
	reply += "\nRelease date: " + song.Response.ReleaseDate
	if song.Response.IsPlaying {
		reply += "\nis_playing: Yes"
	} else {
		reply += "\nis_playing: No"
	}
	s.ChannelMessageSend(m.ChannelID, reply)
}

// For random DT jokes
func getTrumpJoke(s *discordgo.Session, m *discordgo.MessageCreate) {

	url := "http://spotify_api:5001/tjokes"

	resp, err := http.Get(url)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error getting Tjokes!")
		fmt.Println(err)
		return
	}
	var tjoke TJoke
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&tjoke); err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error parsing into json xP")
		return
	}
	s.ChannelMessageSend(m.ChannelID, tjoke.Value)
}

// Gets random chuck norris jokes :P
func getNorrisJokes(s *discordgo.Session, m *discordgo.MessageCreate) {
	resp, err := http.Get("https://api.chucknorris.io/jokes/random")
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error getting Norris jokes xP")
		return
	}

	var joke Joke
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&joke); err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error getting Norris jokes xP")
		return
	}
	s.ChannelMessageSend(m.ChannelID, joke.Value)
}

// api.add_resource(User_playlists, '/user/<username>/playlists/<int:limit>')
func getUserPlaylists(s *discordgo.Session, m *discordgo.MessageCreate, params []string) {

	if len(params) != 4 {
		s.ChannelMessageSend(m.ChannelID, "Expected 4 args, got "+string(len(params)))
		return
	}

	url := "http://spotify_api:5001/user/" + params[1] + "/" + params[2] + "/" + params[3]
	fmt.Println(url)

	resp, err := http.Get(url)
	if err != nil || len(params) != 4 {
		s.ChannelMessageSend(m.ChannelID, "Error getting playlists for the user!")
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error getting the user!")
		return
	}
	res := string(body)
	s.ChannelMessageSend(m.ChannelID, res)
}

// Getting basic user info from the api
func getUserInfo(s *discordgo.Session, m *discordgo.MessageCreate) {
	params := strings.Split(m.Content, "/")
	if len(params) != 2 {
		s.ChannelMessageSend(m.ChannelID, "Wrong number of args given! Expected '!user/bob'")
		return
	}
	resp, err := http.Get("http://spotify_api:5001/user/" + params[1])
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error getting the user!")
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error getting the user!")
		return
	}
	res := string(body)
	s.ChannelMessageSend(m.ChannelID, res)

}

// Getting basic API info from the api
func getAPIInfo(s *discordgo.Session, m *discordgo.MessageCreate) {
	//fmt.Println("Prøver å kontakte api serveren")
	resp, err := http.Get("http://spotify_api:5001/")
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error from the api!")
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error from the api!")
		return
	}
	res := string(body)
	s.ChannelMessageSend(m.ChannelID, res)
}

// Getting current user info from spotify API.
func meSpotify(s *discordgo.Session, m *discordgo.MessageCreate) {
	resp, err := http.Get("http://spotify_api:5001/me")
	if err != nil {
		log.Fatalf("Got error while making request to spotify API (/me), %v", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error while reading body %v", err)
	}
	res := string(body)
	s.ChannelMessageSend(m.ChannelID, res)
}

// Handles greetings
func greeting(s *discordgo.Session, m *discordgo.MessageCreate) {
	answers := []string{"Salutations!", "Greetings!", "Hey there!", "Hello!", "Welcome back!", "Hola!", "Hi! :)", ":wave:"}
	sec := time.Now().Second()

	// Making greetings random
	rand.Seed(int64(sec))
	s.ChannelMessageSend(m.ChannelID, answers[rand.Intn(len(answers))])
}

func author(s *discordgo.Session, m *discordgo.MessageCreate) {
	s.ChannelMessageSend(m.ChannelID, m.Author.Mention())
}

func pingPong(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Content == "!ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	} else {
		s.ChannelMessageSend(m.ChannelID, "Ping!")
	}

}

func calculate(s *discordgo.Session, m *discordgo.MessageCreate) {

}

// fungerer ikke helt
func apps(s *discordgo.Session, m *discordgo.MessageCreate) {
	var myApps []*discordgo.Application
	var err error
	myApps, err = s.Applications()
	if err != nil {
		log.Fatalf("error getting apps, ", err)
	}
	for _, v := range myApps {
		fmt.Println(v)
		//s.ChannelMessageSend(m.ChannelID, string(v))
	}

}
