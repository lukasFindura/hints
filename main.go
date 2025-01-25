package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/buger/goterm"
	"github.com/google/goterm/term"
	"github.com/lukasFindura/gocliselect"
	"gopkg.in/yaml.v3"
)

type Data struct {
	Item     string `json:"name" yaml:"name"`
	Command  string `json:"command,omitempty" yaml:"command,omitempty"`
	SubItems []Data `json:"item,omitempty" yaml:"item,omitempty"`
}

const USAGE = "Usage: %s <file>\n  - file: Input file in JSON (.json) or YAML (.yaml)\n"


// ReadJSONFile reads JSON data from a file specified by filePath
func ReadJSONFile(filePath string) (Data, error) {
	var item Data

	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return item, err
	}

	// Parse based on file extension
	ext := filepath.Ext(filePath)
	switch ext {
	case ".json":
		err = json.Unmarshal(data, &item)
	case ".yaml":
		err = yaml.Unmarshal(data, &item)
	default:
		fmt.Printf("wrong extension: %s\n\n", ext)
		fmt.Printf(USAGE, os.Args[0])
		os.Exit(1)
	}
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

	// Ensure that a file path is provided as an argument
	if len(os.Args) < 2 {
		fmt.Printf(USAGE, os.Args[0])
		os.Exit(1)
	}

	gocliselect.Cursor.ItemPrompt = "❯"
	gocliselect.Cursor.SubMenuPrompt = "❯"
	gocliselect.Cursor.ItemColor = goterm.YELLOW
	gocliselect.Cursor.SubMenuColor = goterm.CYAN
	gocliselect.Cursor.Suffix = " "

	// Get the file path from the command line argument
	path := os.Args[1]

	// Read JSON file
	root, err := ReadJSONFile(path)
	if err != nil {
		log.Fatalf("Failed to read JSON file: %v", err)
	}

	// Initialize rootMenu
	rootMenu := createMenu(root.Item, 0, root.SubItems)

	lastMenu, choice := rootMenu.Display(rootMenu)

	for choice != nil {

		var cmd string = choice.ID
		var args []string
		var verbose bool = true

		// scripts starts with '!' - script file
		if strings.HasPrefix(cmd, "!") {
			cmd = strings.TrimPrefix(cmd, "!")
		} else {
			// scripts starts with '_' - do not print the command being executed
			if strings.HasPrefix(cmd, "_") {
				cmd = strings.TrimPrefix(cmd, "_")
				verbose = false
			}
			// fail if any command in the pipe fails (-o pipefail)
			cmd, args = "bash", []string{"-co", "pipefail", fmt.Sprintf(". ~/.bash_profile; %s", cmd)}
		}

		if verbose {
			// italic and gray color
			fmt.Println(term.Italic(fmt.Sprintf("\033[90mrunning… %s\033[0m", choice.ID)))
		} else {
			fmt.Println()
		}
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
			err_msg := fmt.Sprintf("\033[31m%s\033[0m", err.Error())
			fmt.Println(err_msg)
		}

		fmt.Println()  // empty line before showing menu again
		lastMenu, choice = lastMenu.Display(rootMenu)

	}
	// exit received, clean the menu options
	goterm.MoveCursorUp(gocliselect.LinesOnInput)
	// clear screen from cursor down
	fmt.Fprint(goterm.Screen, "\033[0J")
	goterm.Flush()
}
