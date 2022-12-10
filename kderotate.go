package main

/**
 * a golang wrapper around xrandr and monitor-sensor
 *
 *  https://github.com/donbowman/kde-auto-rotate/blob/master/auto-rotate
 */

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

var primaryDisplayName string

func setScreenOrientation(orientation string) {
	out, err := exec.Command("xrandr", "--output", primaryDisplayName, "--rotate", orientation).Output()
	if err != nil {
		fmt.Println(err.Error())
	} else {
		outS := string(out)
		if len(outS) != 0 {
			fmt.Println(out)
		}
	}
}

func main() {

	success := false
	var cmdOut []byte
	var err error
	for !success {
		cmdOut, err = exec.Command("xrandr", "--current").CombinedOutput()
		if err != nil {
			fmt.Println(err.Error())
			fmt.Println(string(cmdOut))
			time.Sleep(1 * time.Second)
		} else {
			success = true
		}
	}
	currentSetup := string(cmdOut)
	lines := strings.Split(currentSetup, "\n")
	primaryDisplayName = ""
	for _, line := range lines {
		index := strings.Index(line, " primary ")
		if index != -1 {
			index := strings.Index(line, " ")
			primaryDisplayName = line[:index]
		}
	}

	if primaryDisplayName == "" {
		fmt.Print("Unable to find primary display")
		os.Exit(2)
	}

	// The command you want to run along with the argument
	cmd := exec.Command("monitor-sensor")

	// Get a pipe to read from standard out
	r, _ := cmd.StdoutPipe()

	// Use the same pipe for standard error
	cmd.Stderr = cmd.Stdout

	// Make a new channel which will be used to ensure we get all output
	done := make(chan struct{})

	// Create a scanner which scans r in a line-by-line fashion
	scanner := bufio.NewScanner(r)

	// Use the scanner to scan the output line by line and log it
	// It's running in a goroutine so that it doesn't block
	go func() {

		// Read line by line and process it
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Index(line, "Accelerometer orientation changed") != -1 {
				index := strings.Index(line, ":")
				if index != -1 {
					orientation := line[index+2:]
					if orientation == "normal" {
						setScreenOrientation("normal")
					} else if orientation == "bottom-up" {
						setScreenOrientation("inverted")
					} else if orientation == "right-up" {
						setScreenOrientation("right")
					} else if orientation == "left-up" {
						setScreenOrientation("left")
					} else {
						fmt.Println("unexpected orientation: " + orientation)
						setScreenOrientation("normal")
					}
				}
			}
		}

		// We're all done, unblock the channel
		done <- struct{}{}

	}()

	// Start the command and check for errors
	err = cmd.Start()
	if err != nil {
		fmt.Println(err.Error())
	}

	// Wait for all output to be processed
	<-done

	// Wait for the command to finish
	err = cmd.Wait()
	if err != nil {
		fmt.Println(err.Error())
	}
}
