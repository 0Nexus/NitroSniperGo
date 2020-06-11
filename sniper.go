package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"github.com/valyala/fasthttp"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var (
	Token   string
	re      = regexp.MustCompile("(discord.com/gifts/|discordapp.com/gifts/|discord.gift/)([a-zA-Z0-9]+)")
	magenta = color.New(color.FgMagenta)
	green   = color.New(color.FgGreen)
	red     = color.New(color.FgRed)
	strPost = []byte("POST")
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
	color.Cyan("Sniping Discord Nitro on " + strconv.Itoa(len(dg.State.Guilds)) + " Servers 🔫\n\n")

	magenta.Print(t.Format("15:04:05 "))
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

	if strings.Contains(m.Content, "discord.com/gifts") || strings.Contains(m.Content, "discord.gift/") || strings.Contains(m.Content, "discordapp.com/gifts/") {
		start := time.Now()

		code := re.FindStringSubmatch(m.Content)

		if len(code[2]) < 16 {
			magenta.Print(start.Format("15:04:05 "))
			red.Print("[=] Auto-detected a fake code: ")
			red.Print(code[2])
			println(" from " + m.Author.String())
			return
		}

		var strRequestURI = []byte("https://discordapp.com/api/v6/entitlements/gift-codes/" + code[2] + "/redeem")
		req := fasthttp.AcquireRequest()
		req.Header.SetContentType("application/json")
		req.Header.Set("authorization", Token)
		req.SetBody([]byte(`{"channel_id":` + m.ChannelID + "}"))
		req.Header.SetMethodBytes(strPost)
		req.SetRequestURIBytes(strRequestURI)
		res := fasthttp.AcquireResponse()

		if err := fasthttp.Do(req, res); err != nil {
			panic("handle error")
		}

		fasthttp.ReleaseRequest(req)

		body := res.Body()

		bodyString := string(body)
		magenta := color.New(color.FgMagenta)
		magenta.Print(start.Format("15:04:05 "))
		green.Print("[-] Sniped code: ")
		red.Print(code[2])
		println(" from " + m.Author.String())
		magenta.Print(start.Format("15:04:05 "))
		if strings.Contains(bodyString, "This gift has been redeemed already.") {
			color.Yellow("[-] Code has been already redeemed")
		}
		if strings.Contains(bodyString, "nitro") {
			green.Println("[+] Code applied")
		}
		if strings.Contains(bodyString, "Unknown Gift Code") {
			red.Println("[x] Invalid Code")
		}
		fasthttp.ReleaseResponse(res)

	}

}
