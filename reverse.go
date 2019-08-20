/*
Go-Language implementation of an SSH Reverse Tunnel, the equivalent of below SSH command:
   ssh -R 8080:127.0.0.1:8080 operatore@146.148.22.123
which opens a tunnel between the two endpoints and permit to exchange information on this direction:
   server:8080 -----> client:8080
   once authenticated a process on the SSH server can interact with the service answering to port 8080 of the client
   without any NAT rule via firewall
Copyright 2017, Davide Dal Farra
MIT License, http://www.opensource.org/licenses/mit-license.php
*/

package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"

	"golang.org/x/crypto/ssh"
)

type Endpoint struct {
	Host string
	Port int
}

func (endpoint *Endpoint) String() string {
	return fmt.Sprintf("%s:%d", endpoint.Host, endpoint.Port)
}

// From https://sosedoff.com/2015/05/25/ssh-port-forwarding-with-go.html
// Handle local client connections and tunnel data to the remote server
// Will use io.Copy - http://golang.org/pkg/io/#Copy
func handleReverseClient(client net.Conn, remote net.Conn) {
	defer client.Close()
	chDone := make(chan bool)

	// Start remote -> local data transfer
	go func() {
		_, err := io.Copy(client, remote)
		if err != nil {
			log.Println(fmt.Sprintf("error while copy remote->local: %s", err))
		}
		chDone <- true
	}()

	// Start local -> remote data transfer
	go func() {
		_, err := io.Copy(remote, client)
		if err != nil {
			log.Println(fmt.Sprintf("error while copy local->remote: %s", err))
		}
		chDone <- true
	}()

	<-chDone
}

func publicKeyFile() ssh.AuthMethod {
	file := c.SSHKeyPath
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalln(fmt.Sprintf("Cannot read SSH public key file %s", file))
		return nil
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		log.Fatalln(fmt.Sprintf("Cannot parse SSH public key file %s", file))
		return nil
	}
	return ssh.PublicKeys(key)
}

// StartReverse opena reverse tunnel
func StartReverse(node Response) {
	// local service to be forwarded
	var localEndpoint = Endpoint{
		Host: node.Target,
		Port: int(node.SourcePort),
	}

	// remote SSH server
	var serverEndpoint = Endpoint{
		Host: node.Server,
		Port: int(node.Port),
	}

	// remote forwarding port (on remote SSH server network)
	var remoteEndpoint = Endpoint{
		Host: node.Target,
		Port: int(node.TargetPort),
	}
	// refer to https://godoc.org/golang.org/x/crypto/ssh for other authentication types
	sshConfig := &ssh.ClientConfig{
		// SSH connection username
		User: node.User,
		Auth: []ssh.AuthMethod{
			publicKeyFile(),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Connect to SSH remote server using serverEndpoint
	serverConn, err := ssh.Dial("tcp", serverEndpoint.String(), sshConfig)
	if err != nil {
		log.Println(node.ToString())
		log.Println(fmt.Printf("Dial INTO remote server error: %s\n", err))
	}
	if serverConn != nil {
		// Listen on remote server port
		listener, err := serverConn.Listen("tcp", remoteEndpoint.String())
		if err != nil {
			log.Println(node.ToString())
			log.Println(fmt.Printf("Listen open port ON remote server error: %s\n", err))
		}
		if listener != nil {
			defer listener.Close()

			// handle incoming connections on reverse forwarded tunnel
			for {
				// Open a (local) connection to localEndpoint whose content will be forwarded so serverEndpoint
				local, err := net.Dial("tcp", localEndpoint.String())
				if err != nil {
					log.Println(node.ToString())
					log.Println(fmt.Printf("Dial INTO local service error: %s\n", err))
				}
				if local != nil {
					client, err := listener.Accept()
					if err != nil {
						log.Println(node.ToString())
						log.Println(err)
					}
					if client != nil {
						handleReverseClient(client, local)
					}
				}
			}
		}
	}
}
