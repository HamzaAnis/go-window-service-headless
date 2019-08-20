package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
)

// Get default location of a private key
func privateKeyPath() string {
	path := filepath.FromSlash("C:/Users/Sardard/.ssh/id_rsa")
	return path
}

// Get private key for ssh authentication
func parsePrivateKey(keyPath string) (ssh.Signer, error) {
	buff, _ := ioutil.ReadFile(keyPath)
	return ssh.ParsePrivateKey(buff)
}

// Get ssh client config for our connection
// SSH config will use 2 authentication strategies: by key and by password
func makeSshConfig(user, password string) (*ssh.ClientConfig, error) {
	key, err := parsePrivateKey(privateKeyPath())
	if err != nil {
		return nil, err
	}

	config := ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
			ssh.PublicKeys(key),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	return &config, nil
}

// Handle local client connections and tunnel data to the remote server
// Will use io.Copy - http://golang.org/pkg/io/#Copy
func handleClient(client net.Conn, remote net.Conn) {
	defer client.Close()
	chDone := make(chan bool)

	// Start remote -> local data transfer
	go func() {
		_, err := io.Copy(client, remote)
		if err != nil {
			log.Println("error while copy remote->local:", err)
		}
		chDone <- true
	}()

	// Start local -> remote data transfer
	go func() {
		_, err := io.Copy(remote, client)
		if err != nil {
			log.Println(err)
		}
		chDone <- true
	}()

	<-chDone
}
func getHostKey(host string) (ssh.PublicKey, error) {
	file, err := os.Open(filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts"))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var hostKey ssh.PublicKey
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), " ")
		if len(fields) != 3 {
			continue
		}
		if strings.Contains(fields[0], host) {
			var err error
			hostKey, _, _, _, err = ssh.ParseAuthorizedKey(scanner.Bytes())
			if err != nil {
				return nil, errors.New(fmt.Sprintf("error parsing %q: %v", fields[2], err))
			}
			break
		}
	}

	if hostKey == nil {
		return nil, errors.New(fmt.Sprintf("no hostkey for %s", host))
	}
	return hostKey, nil
}

func StartForwardTunnel(node Response) {
	// Connection settings
	sshAddr := fmt.Sprintf("%v:%v", node.Server, node.Port)
	localAddr := fmt.Sprintf("%v:%v", node.Target, node.SourcePort)
	remoteAddr := fmt.Sprintf("%v:%v", node.Target, node.TargetPort)

	// Build SSH client configuration
	cfg, err := makeSshConfig(node.User, node.Password)
	if err != nil {
		log.Fatalln(err)
	}

	// Establish connection with SSH server
	conn, err := ssh.Dial("tcp", sshAddr, cfg)
	if err != nil {
		log.Println(err)
	}
	if conn != nil {
		defer conn.Close()

		// Establish connection with remote server
		remote, err := conn.Dial("tcp", remoteAddr)
		if err != nil {
			log.Fatalln(err)
		}
		if remote != nil {
			// Start local server to forward traffic to remote connection
			local, err := net.Listen("tcp", localAddr)
			if err != nil {
				log.Fatalln(err)
			}
			if local != nil {
				defer local.Close()

				// Handle incoming connections
				for {
					client, err := local.Accept()
					if err != nil {
						log.Println(err)
					}
					handleClient(client, remote)
				}
			}
		}
	}
}
