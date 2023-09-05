package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/andrewbenington/go-spotify/app"
	"github.com/andrewbenington/go-spotify/auth"
	"github.com/andrewbenington/go-spotify/playlist"
	"github.com/google/uuid"
	"github.com/zmb3/spotify/v2"
)

var (
	client   *spotify.Client
	session  string
	commands = map[string]func() error{
		"auth":       Authenticate,
		"auth fresh": FreshAuthenticate,
		"playlists":  playlist.ListPlaylists,
	}
)

func main() {
	a := app.App{}
	a.Initialize()
	addr := "localhost:5757"
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}
	a.Run(addr)
	// Authenticate()
	// GetCommands()
}

func GetCommands() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Enter a command: ")
		command, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading command:", err)
			continue
		}

		// Trim any leading/trailing white spaces and newlines
		command = strings.ToLower(strings.TrimSpace(command))

		// Check if the user wants to exit
		if command == "exit" {
			fmt.Println("Exiting...")
			break
		}

		function := commands[command]
		if function != nil {
			err := function()
			if err != nil {
				fmt.Println(err)
			}
			continue
		}
		fmt.Printf("Command '%s' not supported\n", command)

	}

	fmt.Println("Program ended.")
}

func Authenticate() error {
	session := uuid.New()
	client = auth.AuthenticateUser(session.String(), false)
	return nil
}

func FreshAuthenticate() error {
	session := uuid.New()
	client = auth.AuthenticateUser(session.String(), true)
	return nil
}
