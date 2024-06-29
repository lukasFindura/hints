package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/buger/goterm"
	"github.com/google/goterm/term"
	"github.com/lukasFindura/gocliselect"
)

type Data struct {
	Item     string `json:"name"`
	Command  string `json:"command,omitempty"`
	SubItems []Data `json:"item,omitempty"`
}

// ReadJSONFile reads JSON data from a file specified by filePath
func ReadJSONFile(filePath string) (Data, error) {
	var item Data

	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return item, err
	}

	// Unmarshal JSON into the Data struct
	err = json.Unmarshal(data, &item)
	if err != nil {
		return item, err
	}

	return item, nil
}

// Function to create menu
func createMenu(name string, indent int, subItems []Data) *gocliselect.Menu {
	menu := gocliselect.NewMenu(name, indent)
	for _, item := range subItems {
		if item.SubItems != nil {
			subMenu := createMenu(item.Item, indent+1, item.SubItems)
			menu.AddItemMenu(item.Item, "", subMenu)
		} else {
			menu.AddItem(item.Item, item.Command)
		}
	}
	return menu
}

func main() {

	gocliselect.Cursor.ItemPrompt = "❯"
	gocliselect.Cursor.SubMenuPrompt = "❯"
	gocliselect.Cursor.ItemColor = goterm.YELLOW
	gocliselect.Cursor.SubMenuColor = goterm.CYAN
	gocliselect.Cursor.Suffix = " "

	path := "example.json"

	// Read JSON file
	root, err := ReadJSONFile(path)
	if err != nil {
		log.Fatalf("Failed to read JSON file: %v", err)
	}

	// Initialize rootMenu
	rootMenu := createMenu(root.Item, 0, root.SubItems)

	lastMenu, choice := rootMenu.Display(rootMenu)

	for choice != nil {

		var cmd string
		var args []string

		// scripts starts with '!'
		if strings.HasPrefix(choice.ID, "!") {
			cmd = strings.TrimPrefix(choice.ID, "!")
		} else {
			cmd, args = "bash", []string{"-c", fmt.Sprintf(". ~/.bash_profile; %s", choice.ID)}
		}

		// italic and gray color
		text := term.Italic(fmt.Sprintf("\033[90mrunning… %s\033[0m\n", choice.ID))
		fmt.Print(text)
		out := exec.Command(cmd, args...)

		// https://stackoverflow.com/questions/18106749/golang-catch-signals
		signalChannel := make(chan os.Signal, 2)
		signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
		go func() {
			sig := <-signalChannel
			switch sig {
			case os.Interrupt:
				//handle SIGINT
			case syscall.SIGTERM:
				//handle SIGTERM
			}
		}()

		// https://stackoverflow.com/questions/30207035/golang-exec-command-read-std-input
		out.Stdout = os.Stdout
		out.Stderr = os.Stderr
		out.Stdin = os.Stdin
		err := out.Run()
		if err != nil {
			fmt.Println(err.Error())
			// os.Exit(1)
		}

		lastMenu, choice = lastMenu.Display(rootMenu)

	}
}
