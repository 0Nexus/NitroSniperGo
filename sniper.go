package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Variables used for command line parameters
var (
	Token string
)

func init() {
	file, err := ioutil.ReadFile("token.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed read file: %s\n", err)
		os.Exit(1)
	}

	var f interface{}
	err = json.Unmarshal(file, &f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse JSON: %s\n", err)
		os.Exit(1)
	}

	// Type-cast `f` to a map by means of type assertion.
	m := f.(map[string]interface{})
	fmt.Printf("Parsed data: %v\n", m)

	str := fmt.Sprintf("%v", m["token"])

	flag.StringVar(&Token, "t", str, "Token")
	flag.Parse()
}

func main() {
	fmt.Print("\033[2J")
	c := exec.Command("clear")

	c.Stdout = os.Stdout
	c.Run() // Create a new Discord session using the provided bot token.
	color.Red(`
▓█████▄  ██▓  ██████  ▄████▄   ▒█████   ██▀███  ▓█████▄      ██████  ███▄    █  ██▓ ██▓███  ▓█████  ██▀███
▒██▀ ██▌▓██▒▒██    ▒ ▒██▀ ▀█  ▒██▒  ██▒▓██ ▒ ██▒▒██▀ ██▌   ▒██    ▒  ██ ▀█   █ ▓██▒▓██░  ██▒▓█   ▀ ▓██ ▒ ██▒
░██   █▌▒██▒░ ▓██▄   ▒▓█    ▄ ▒██░  ██▒▓██ ░▄█ ▒░██   █▌   ░ ▓██▄   ▓██  ▀█ ██▒▒██▒▓██░ ██▓▒▒███   ▓██ ░▄█ ▒
░▓█▄   ▌░██░  ▒   ██▒▒▓▓▄ ▄██▒▒██   ██░▒██▀▀█▄  ░▓█▄   ▌     ▒   ██▒▓██▒  ▐▌██▒░██░▒██▄█▓▒ ▒▒▓█  ▄ ▒██▀▀█▄
░▒████▓ ░██░▒██████▒▒▒ ▓███▀ ░░ ████▓▒░░██▓ ▒██▒░▒████▓    ▒██████▒▒▒██░   ▓██░░██░▒██▒ ░  ░░▒████▒░██▓ ▒██▒
▒▒▓  ▒ ░▓  ▒ ▒▓▒ ▒ ░░ ░▒ ▒  ░░ ▒░▒░▒░ ░ ▒▓ ░▒▓░ ▒▒▓  ▒    ▒ ▒▓▒ ▒ ░░ ▒░   ▒ ▒ ░▓  ▒▓▒░ ░  ░░░ ▒░ ░░ ▒▓ ░▒▓░
░ ▒  ▒  ▒ ░░ ░▒  ░ ░  ░  ▒     ░ ▒ ▒░   ░▒ ░ ▒░ ░ ▒  ▒    ░ ░▒  ░ ░░ ░░   ░ ▒░ ▒ ░░▒ ░      ░ ░  ░  ░▒ ░ ▒░
░ ░  ░  ▒ ░░  ░  ░  ░        ░ ░ ░ ▒    ░░   ░  ░ ░  ░    ░  ░  ░     ░   ░ ░  ▒ ░░░          ░     ░░   ░
░     ░        ░  ░ ░          ░ ░     ░        ░             ░           ░  ░              ░  ░   ░
░                   ░                           ░
	`)
	dg, err := discordgo.New(Token)

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
	t := time.Now()
	m := color.New(color.FgMagenta)
	color.Cyan("Sniping Discord Nitro on " + strconv.Itoa(len(dg.State.Guilds)) + " Servers 🔫\n\n")

	m.Print(t.Format("15:04:05 "))
	fmt.Println("[+] Bot is ready")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	if strings.Contains(m.Content, "discordapp.com/gifts/") || strings.Contains(m.Content, "discord.gift/") {
		t := time.Now()

		re := regexp.MustCompile("(discordapp.com/gifts/|discord.gift/)([a-zA-Z0-9]+)")
		code := re.FindStringSubmatch(m.Content)
		var jsonStr = []byte(`{"channel_id":` + m.ChannelID + "}")

		if len(code[2]) != 16 {
			color.Red("[x] Invalid Code")

			return
		}

		req, err := http.NewRequest("POST", "https://discordapp.com/api/v6/entitlements/gift-codes/"+code[2]+"/redeem", bytes.NewBuffer(jsonStr))
		req.Header.Set("Content-Type", "application/json")

		req.Header.Set("authorization", Token)
		client := &http.Client{}
		resp, err := client.Do(req)
		t2 := time.Now()

		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		bodyString := string(bodyBytes)
		magenta := color.New(color.FgMagenta)
		magenta.Print(t.Format("15:04:05 "))
		color.Green("[-] Snipped code: " + code[2] + " by " + m.Author.String())
		magenta.Print(t2.Format("15:04:05 "))
		if strings.Contains(bodyString, "This gift has been redeemed already.") {
			color.Yellow("[-] Code has been already redeemed")
		}
		if strings.Contains(bodyString, "nitro") {
			color.Green("[+] Code applied")
		}
		if strings.Contains(bodyString, "Unknown Gift Code") {
			color.Red("[x] Invalid Code")
		}
	}

}
