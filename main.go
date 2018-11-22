package main // import "github.com/mrmagooey/sshaha/modules"

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"syscall"

	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"

	"github.com/kr/pty"
	"golang.org/x/crypto/ssh"
)

var (
	// defaultShell asdf
	defaultShell = "sh"
)

// generateSSHKeys returns two byte slices filled with new, pem block encoded private and public keys
func generateSSHKeys() (pub []byte, priv []byte, err error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	privKeyBuf := &bytes.Buffer{}
	privateKeyPEM := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}
	if err = pem.Encode(privKeyBuf, privateKeyPEM); err != nil {
		return nil, nil, err
	}
	priv = privKeyBuf.Bytes()
	// generate and write public key
	pkix, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, nil, err
	}
	publicKeyPEM := &pem.Block{Type: "RSA PUBLIC KEY", Bytes: pkix}
	pubKeyBuf := &bytes.Buffer{}
	if err = pem.Encode(pubKeyBuf, publicKeyPEM); err != nil {
		return nil, nil, err
	}
	pub = pubKeyBuf.Bytes()
	if err != nil {

	}
	return pub, priv, nil
}

// https://github.com/Scalingo/go-ssh-examples/blob/master/server_complex.go

func ptyRun(c *exec.Cmd, tty *os.File) (err error) {
	defer tty.Close()
	c.Stdout = tty
	c.Stdin = tty
	c.Stderr = tty
	c.SysProcAttr = &syscall.SysProcAttr{
		Setctty: true,
		Setsid:  true,
	}
	return c.Start()
}

// parseDims extracts two uint32s from the provided buffer.
func parseDims(b []byte) (uint32, uint32) {
	w := binary.BigEndian.Uint32(b)
	h := binary.BigEndian.Uint32(b[4:])
	return w, h
}

// Winsize stores the Height and Width of a terminal.
type Winsize struct {
	Height uint16
	Width  uint16
	x      uint16 // unused
	y      uint16 // unused
}

// SetWinsize sets the size of the given pty.
// func SetWinsize(fd uintptr, w, h uint32) {
// 	log.Printf("window resize %dx%d", w, h)
// 	ws := &Winsize{Width: uint16(w), Height: uint16(h)}
// 	syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TIOCSWINSZ), uintptr(unsafe.Pointer(ws)))
// }

func handleExec() bool {
	// command := string(req.Payload[4 : req.Payload[3]+4])

	// cmd := exec.Command(shell, []string{"-c", command}...)

	// cmd.Stdout = channel
	// cmd.Stderr = channel
	// cmd.Stdin = channel

	// err := cmd.Start()
	// if err != nil {
	// 	log.Printf("could not start command (%s)", err)
	// 	continue
	// }

	// // teardown session
	// go func() {
	// 	_, err := cmd.Process.Wait()
	// 	if err != nil {
	// 		log.Printf("failed to exit bash (%s)", err)
	// 	}
	// 	channel.Close()
	// 	log.Printf("session closed")
	// }()

	return true
}

func handlePty() bool {
	// Parse body...
	// termLen := req.Payload[3]
	// termEnv := string(req.Payload[4 : termLen+4])
	// w, h := parseDims(req.Payload[termLen+4:])
	// SetWinsize(f.Fd(), w, h)
	// log.Printf("pty-req '%s'", termEnv)
	return true
}

func handleEnv(req *ssh.Request) bool {
	// Parse body...
	// termLen := req.Payload[3]
	// termEnv := string(req.Payload[4 : termLen+4])
	// w, h := parseDims(req.Payload[termLen+4:])
	// SetWinsize(f.Fd(), w, h)
	// log.Printf("pty-req '%s'", termEnv)
	return true
}

func handleShell(channel ssh.Channel, req *ssh.Request, tty *os.File, f *os.File) bool {
	// gnomeKeychainTrick(channel)
	signalsTrick(channel)
	// line, more, err := bio.ReadLine()

	// fmt.Print(more)
	// fmt.Print(err)
	// s := string(line)
	// fmt.Println(s)

	// in case you need a string which contains the newline
	// s, err := bio.ReadString('\n')
	// fmt.Println(s)

	//	ioutil.ReadAll(channel)

	// scanner := bufio.NewScanner(channel) //

	//	scanner.Split(bufio.ScanWords)

	// for scanner.Scan() {
	// 	fmt.Println("c")
	// 	line := scanner.Text()
	// 	fmt.Println("c")
	// 	if line == "\n" {
	// 		break
	// 	}
	// 	fmt.Println(line)

	// }

	// if b, err := ioutil.ReadAll(channel); err == nil {
	// 	fmt.Println(string(b))
	// }

	// buf := new(bytes.Buffer)
	// buf.ReadFrom(channel)

	// s := buf.String()
	// fmt.Println("c")
	// fmt.Println(s)

	// shell := defaultShell
	// cmd := exec.Command(shell)
	// cmd.Env = []string{"TERM=xterm"}
	// err := ptyRun(cmd, tty)
	// if err != nil {
	// 	log.Printf("%s", err)
	// }
	// // Teardown session
	// var once sync.Once
	// close := func() {
	// 	channel.Close()
	// 	log.Printf("session closed")
	// }
	// // Pipe session to bash and visa-versa
	// go func() {
	// 	io.Copy(channel, f)
	// 	once.Do(close)
	// }()
	// go func() {
	// 	io.Copy(f, channel)
	// 	once.Do(close)
	// }()
	// // We don't accept any commands (Payload),
	// // only the default shell.
	// if len(req.Payload) == 0 {
	// 	return true
	// }

	return true

}

func handleRequest(in <-chan *ssh.Request, channel ssh.Channel, tty *os.File, f *os.File) error {
	for req := range in {
		// log.Printf("%v %s", req.Payload, req.Payload)
		ok := false
		switch req.Type {
		case "exec":
			ok = handleExec()
		case "shell":
			ok = handleShell(channel, req, tty, f)
		case "pty-req":
			ok = handlePty()
		case "window-change":
			continue // no response
		case "env":
			ok = handleEnv(req)
		}
		if !ok {
			return fmt.Errorf("declining %s request", req.Type)
		}
		req.Reply(ok, nil)
	}
	return nil
}

func handleChannels(chans <-chan ssh.NewChannel) {
	// Service the incoming Channel channel.
	for newChannel := range chans {
		if t := newChannel.ChannelType(); t != "session" {
			newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
			continue
		}
		channel, requests, err := newChannel.Accept()
		if err != nil {
			log.Printf("could not accept channel (%s)", err)
			continue
		}
		f, tty, err := pty.Open()
		if err != nil {
			log.Printf("could not start pty (%s)", err)
			continue
		}
		go handleRequest(requests, channel, tty, f)
	}
}

func main() {
	config := &ssh.ServerConfig{
		NoClientAuth: true,
		// Remove to disable password auth.
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			return nil, nil
		},

		//		Remove to disable public key auth.
		PublicKeyCallback: func(c ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
			return nil, nil
		},
	}

	_, priv, err := generateSSHKeys()
	if err != nil {
		log.Fatal("couldn't generate keys")
	}

	private, err := ssh.ParsePrivateKey(priv)
	if err != nil {
		log.Fatal("Failed to parse private key: ", err)
	}

	config.AddHostKey(private)

	// Once a ServerConfig has been configured, connections can be
	// accepted.
	listener, err := net.Listen("tcp", "127.0.0.1:2022")
	if err != nil {
		log.Fatal("failed to listen for connection: ", err)
	}

	nConn, err := listener.Accept()
	if err != nil {
		log.Fatal("failed to accept incoming connection: ", err)
	}

	// Before use, a handshake must be performed on the incoming
	// net.Conn.
	_, chans, reqs, err := ssh.NewServerConn(nConn, config)
	if err != nil {
		log.Fatal("failed to handshake: ", err)
	}

	// The incoming Request channel must be serviced.
	go ssh.DiscardRequests(reqs)

	// Service the incoming Channel channel.
	handleChannels(chans)
	// for newChannel := range chans {
	// 	// Channels have a type, depending on the application level
	// 	// protocol intended. In the case of a shell, the type is
	// 	// "session" and ServerShell may be used to present a simple
	// 	// terminal interface.
	// 	if newChannel.ChannelType() != "session" {
	// 		newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
	// 		continue
	// 	}
	// 	channel, requests, err := newChannel.Accept()
	// 	if err != nil {
	// 		log.Fatalf("Could not accept channel: %v", err)
	// 	}

	// 	// Sessions have out-of-band requests such as "shell",
	// 	// "pty-req" and "env".  Here we handle only the
	// 	// "shell" request.
	// 	go func(in <-chan *ssh.Request) {
	// 		for req := range in {
	// 			fmt.Println(req)
	// 			if req.Type == "shell" {
	// 				err = req.Reply(true, nil)
	// 				if err != nil {
	// 					fmt.Println("bad shell reply")
	// 				}
	// 			}
	// 			if req.Type == "pty-req" {

	// 			}
	// 		}
	// 	}(requests)

	// 	term := terminal.NewTerminal(channel, "> ")
	// 	go func() {
	// 		defer channel.Close()
	// 		for {
	// 			fmt.Println("prior ")
	// 			line, err := term.ReadLine()
	// 			fmt.Println("after")
	// 			if err != nil {
	// 				break
	// 			}
	// 			fmt.Println(line)
	// 		}
	// 	}()
	// }
}
