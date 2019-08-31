package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
)

// StartReverse opena reverse tunnel
func StartReverse(node Response) {
	addr := fmt.Sprintf("%v:%v:%v", node.SourcePort, node.Target, node.TargetPort)
	host := fmt.Sprintf("%v@%v", node.User, node.Server)

	subProcess := exec.Command("plink.exe", "-ssh", "-N", "-pw", node.Password, "-L", addr, host) //Just for testing, replace with your subProcess

	stdin, err := subProcess.StdinPipe()
	if err != nil {
		fmt.Println(err) //replace with logger, or anything you want
	}
	defer stdin.Close() // the doc says subProcess.Wait will close it, but I'm not sure, so I kept this line

	subProcess.Stdout = os.Stdout
	subProcess.Stderr = os.Stderr

	log.Println(node.ToString())
	if err = subProcess.Start(); err != nil { //Use start, not run
		fmt.Println("An error occured: ", err) //replace with logger, or anything you want
	}
	io.WriteString(stdin, "y\r\ny\r\ny\r\n")
	io.WriteString(stdin, "y\r\ny\r\ny\r\n")
	io.WriteString(stdin, "y\r\ny\r\ny\r\n")

	subProcess.Wait()
}
