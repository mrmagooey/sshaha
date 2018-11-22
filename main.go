package main // import "github.com/mrmagooey/sshaha/modules"

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"reflect"

	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
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

func handleExec() bool {
	return true
}

func handlePty() bool {
	return true
}

func handleEnv(req *ssh.Request) bool {
	return true
}

func handleShell(channel ssh.Channel, req *ssh.Request, tty *os.File, f *os.File, connDetails map[string]string) bool {
	// gnomeKeychainTrick(channel)
	signalsTrick(channel, connDetails)
	return true
}

func handleRequest(in <-chan *ssh.Request, channel ssh.Channel, tty *os.File, f *os.File, connDetails map[string]string) error {
	for req := range in {
		ok := false
		switch req.Type {
		case "exec":
			ok = handleExec()
		case "shell":
			ok = handleShell(channel, req, tty, f, connDetails)
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

func handleChannels(chans <-chan ssh.NewChannel, connDetails map[string]string) {
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
		go handleRequest(requests, channel, tty, f, connDetails)
	}
}

func getUsername(f *ssh.Conn) string {
	v := reflect.ValueOf(*f)
	y := v.Elem().FieldByName("sshConn")
	return y.FieldByName("user").String()
}

func getClientSSHVersion(f *ssh.Conn) string {
	v := reflect.ValueOf(*f)
	y := v.Elem().FieldByName("sshConn")
	clientVersion := y.FieldByName("clientVersion").Bytes()
	return string(clientVersion)
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

	conn, chans, reqs, err := ssh.NewServerConn(nConn, config)
	if err != nil {
		log.Println("failed to handshake: ", err)
	}

	username := getUsername(&conn.Conn)
	sshVersion := getClientSSHVersion(&conn.Conn)
	connectionDetails := map[string]string{
		"username":   username,
		"sshVersion": sshVersion,
	}

	// The incoming Request channel must be serviced.
	go ssh.DiscardRequests(reqs)

	// Service the incoming Channel channel.
	handleChannels(chans, connectionDetails)

}
