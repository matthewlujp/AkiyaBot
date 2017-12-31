package main

import (
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
	PrivateKeyPath string
}

var (
	cnf           config
	retryInterval = 5 * time.Second
	logger        = log.New(os.Stdout, "", 0)
)

func init() {
	if _, err := toml.DecodeFile(configFile, &cnf); err != nil {
		log.Fatalf("error while parsing conf toml: %s", err)
	}
	logger.Println("")
	logger.Printf("host: %s", cnf.RelayHostName)
	logger.Printf("user: %s ", cnf.RelayUserName)
	logger.Printf("port: %d", cnf.ForwardPort)
	logger.Printf("private key: %s", cnf.PrivateKeyPath)
}

func handleClient(client, remote net.Conn) {
	defer client.Close()
	chDone := make(chan bool)

	go func() {
		if _, err := io.Copy(client, remote); err != nil {
			logger.Printf("error while copy remote -> local: %s", err)
		}
		chDone <- true
	}()

	go func() {
		if _, err := io.Copy(remote, client); err != nil {
			logger.Printf("error while copy local -> remote: %s", err)
		}
		chDone <- true
	}()

	<-chDone
}

func main() {
	auth, err := publicKeyFile(cnf.PrivateKeyPath)
	if err != nil {
		log.Fatalf("error processing private key: %s", err)
	}
	log.Printf("private key file %s read successfully", cnf.PrivateKeyPath)

	interEndpoint := &endpoint{ // Endpoint for reverse ssh
		host: cnf.RelayHostName,
		port: 22,
	}
	forwardEndpoint := &endpoint{
		host: "localhost",
		port: cnf.ForwardPort,
	}
	localEndpoint := &endpoint{
		host: "localhost",
		port: 22,
	}

	sshConfig := &ssh.ClientConfig{
		User:            cnf.RelayUserName,
		Auth:            []ssh.AuthMethod{auth},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Never dies
	for {
		errChan := make(chan error)

		go func() {
			serverConn, err := ssh.Dial("tcp", interEndpoint.String(), sshConfig)
			if err != nil {
				errChan <- err
				logger.Print(err)
				return
			}
			logger.Print("server connected")

			listener, err := serverConn.Listen("tcp", forwardEndpoint.String())
			if err != nil {
				logger.Print(err)
				errChan <- err
				return
			}
			logger.Print("server listener created")
			defer listener.Close()

			for {
				local, err := net.Dial("tcp", localEndpoint.String()) // local and forwared request are conneced
				if err != nil {
					logger.Print(err)
					errChan <- err
					return
				}
				logger.Print("local connected")

				client, err := listener.Accept()
				if err != nil {
					logger.Print(err)
					errChan <- err
					return
				}
				logger.Print("client accepted")

				handleClient(client, local)
			}
		}()

		<-errChan
		logger.Printf("retry connection after %v seconds", retryInterval)
		time.Sleep(retryInterval)
	}
}
