package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

func main() {
	subProcess := exec.Command("plink.exe", "-ssh", "-N", "-pw", "somePassword", "-R", "8001:localhost:3000", "tun-user@public.nsplice.com") //Just for testing, replace with your subProcess

	stdin, err := subProcess.StdinPipe()
	if err != nil {
		fmt.Println(err) //replace with logger, or anything you want
	}
	defer stdin.Close() // the doc says subProcess.Wait will close it, but I'm not sure, so I kept this line

	subProcess.Stdout = os.Stdout
	subProcess.Stderr = os.Stderr

	fmt.Println("START")                      //for debug
	if err = subProcess.Start(); err != nil { //Use start, not run
		fmt.Println("An error occured: ", err) //replace with logger, or anything you want
	}
	io.WriteString(stdin, "y\r\ny\r\ny\r\n")
	io.WriteString(stdin, "y\r\ny\r\ny\r\n")
	io.WriteString(stdin, "y\r\ny\r\ny\r\n")

	subProcess.Wait()
	fmt.Println("END") //for debug
}
