package main

import (
	"flag"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"golang.org/x/crypto/ssh"
)

const configFile = "./conf.toml"

type config struct {
	RelayHostName  string
	RelayUserName  string
	ForwardPort    int
	LocalPort      int
	PrivateKeyPath string
}

var (
	cnf             config
	retryInterval   = 5 * time.Second
	logger          = log.New(os.Stdout, "", 0)
	forwardPort     *int
	localPort       *int
	interEndpoint   *endpoint
	forwardEndpoint *endpoint
	localEndpoint   *endpoint
)

func init() {
	if _, err := toml.DecodeFile(configFile, &cnf); err != nil {
		log.Fatalf("error while parsing conf toml: %s", err)
	}

	forwardPort = flag.Int("remote_port", cnf.ForwardPort, "remote port number")
	localPort = flag.Int("local_port", cnf.LocalPort, "local port number")
	flag.Parse()

	logger.Println("")
	logger.Printf("host: %s", cnf.RelayHostName)
	logger.Printf("user: %s ", cnf.RelayUserName)
	logger.Printf("port: %d", *forwardPort)
	logger.Printf("private key: %s", cnf.PrivateKeyPath)

	interEndpoint = &endpoint{ // Endpoint for reverse ssh
		host: cnf.RelayHostName,
		port: 22,
	}
	forwardEndpoint = &endpoint{
		host: "localhost",
		port: *forwardPort,
	}
	localEndpoint = &endpoint{
		host: "localhost",
		port: *localPort,
	}
}

func handleClient(client, remote net.Conn) {
	defer client.Close()
	done := make(chan bool)

	go func() {
		if _, err := io.Copy(client, remote); err != nil {
			logger.Printf("error while copy remote -> local: %s", err)
		}
		done <- true
	}()

	go func() {
		if _, err := io.Copy(remote, client); err != nil {
			logger.Printf("error while copy local -> remote: %s", err)
		}
		done <- true
	}()

	<-done
}

func connect(sshConfig *ssh.ClientConfig, quit chan<- struct{}) {
	serverConn, err := ssh.Dial("tcp", interEndpoint.String(), sshConfig)
	if err != nil {
		logger.Print(err)
		close(quit)
		return
	}
	logger.Print("server connected")

	listener, err := serverConn.Listen("tcp", forwardEndpoint.String())
	if err != nil {
		logger.Print(err)
		close(quit)
		return
	}
	logger.Print("server listener created")
	defer listener.Close()

	for {
		local, err := net.Dial("tcp", localEndpoint.String()) // local and forwared request are conneced
		if err != nil {
			logger.Print(err)
			close(quit)
			return
		}
		logger.Print("local connected")

		client, err := listener.Accept()
		if err != nil {
			logger.Print(err)
			close(quit)
			return
		}
		logger.Print("client accepted")

		handleClient(client, local)
	}
}

func main() {
	auth, err := publicKeyFile(cnf.PrivateKeyPath)
	if err != nil {
		log.Fatalf("error processing private key: %s", err)
	}
	log.Printf("private key file %s read successfully", cnf.PrivateKeyPath)

	sshConfig := &ssh.ClientConfig{
		User:            cnf.RelayUserName,
		Auth:            []ssh.AuthMethod{auth},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Never dies
	for {
		quit := make(chan struct{})
		connect(sshConfig, quit)
		<-quit
		logger.Printf("retry connection after %v seconds", retryInterval)
		time.Sleep(retryInterval)
	}
}
