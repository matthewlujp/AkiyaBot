package main

import (
	"io"
	"log"
	"net"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

var (
	interUsername  = "ishige"
	interHost      = "akiyagri.akg.t.u-tokyo.ac.jp"
	forwardPort    = 3030
	privateKeyFile = "/Users/luning/.ssh/akiya_pc"
	retryInterval  = 5 * time.Second
	logger         = log.New(os.Stdout, "", 0)
)

func init() {
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
	auth, err := publicKeyFile(privateKeyFile)
	if err != nil {
		log.Fatalf("error processing private key: %s", err)
	}
	log.Printf("private key file %s read successfully", privateKeyFile)

	interEndpoint := &endpoint{ // Endpoint for reverse ssh
		host: interHost,
		port: 22,
	}
	forwardEndpoint := &endpoint{
		host: "localhost",
		port: forwardPort,
	}
	localEndpoint := &endpoint{
		host: "localhost",
		port: 22,
	}

	sshConfig := &ssh.ClientConfig{
		User:            interUsername,
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
