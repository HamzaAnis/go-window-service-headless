package main

import (
	"log"
	"time"

	"github.com/rgzr/sshtun"
)

func openForwarPort(node Response) *sshtun.SSHTun {
	sshTun := sshtun.New(int(node.SourcePort), node.Server, int(node.TargetPort))

	sshTun.SetUser(node.User)
	sshTun.SetPassword(node.Password)
	sshTun.SetRemoteHost(node.Server)

	// We enable debug messages to see what happens
	sshTun.SetDebug(true)

	// We set a callback to know when the tunnel is ready
	sshTun.SetConnState(func(tun *sshtun.SSHTun, state sshtun.ConnState) {
		switch state {
		case sshtun.StateStarting:
			log.Printf("%v is Starting\n", node.ToString())
		case sshtun.StateStarted:
			log.Printf("%v is Started\n", node.ToString())
		case sshtun.StateStopped:
			log.Printf("%v is Stopped\n", node.ToString())
		}
	})

	// We start the tunnel (and restart it every time it is stopped)
	go func() {
		for {
			if err := sshTun.Start(); err != nil {
				log.Printf("SSH tunnel stopped: %s", err.Error())
				time.Sleep(time.Second) // don't flood if there's a start error :)
			}
		}
	}()

	// // We stop the tunnel every 20 seconds (just to see what happens)
	// for {
	// 	time.Sleep(time.Second * time.Duration(20))
	// 	log.Println("Lets stop the SSH tunnel...")
	// 	sshTun.Stop()
	// }
	return sshTun
}
