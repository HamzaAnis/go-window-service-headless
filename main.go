package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/parnurzeal/gorequest"
)

var (
	c Config
)

// Config store the config variables
type Config struct {
	URL        string `json:"url"`
	Ports      int64  `json:"ports"`
	APIKey     string `json:"apiKey"`
	SSHKeyPath string `json:"sshKeyPath"`
}

// Response store the response of the endpoint
type Response struct {
	Server     string `json:"server"`
	Port       int64  `json:"port"`
	User       string `json:"user"`
	Password   string `json:"password"`
	Direction  string `json:"direction"`
	Target     string `json:"target"`
	TargetPort int64  `json:"targetPort"`
	SourcePort int64  `json:"sourcePort"`
}

func (r *Response) print() {
	log.Printf("\nServer: %v\nPort: %v\nUser: %v\nPassword: %v\nDirection: %v\nTarget: %v\nTargetPort: %v\nSourcePort: %v\n\n", r.Server, r.Port, r.User, r.Password, r.Direction, r.Target, r.TargetPort, r.SourcePort)
}

// ToString returns the string form of response
func (r *Response) ToString() string {
	flag := "-R"
	if r.Direction == "reverse" {
		flag = "-R"
	} else if r.Direction == "forward" {
		flag = "-L"
	}
	return fmt.Sprintf("ssh %v %v:%v:%v %v@%v -N", flag, r.SourcePort, r.Target, r.TargetPort, r.User, r.Server)

}

// loadConfig loads reads the configuration file
func loadConfig(configFile string) Config {
	log.Printf("Loading config file %v\n", configFile)
	cnfg, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatalf("ERROR: %v\n", err)
	}
	var c Config
	json.Unmarshal(cnfg, &c)
	return c
}

// buildURL builds the url from the config file
func buildURL(c Config) string {
	baseURL, err := url.Parse(c.URL)
	if err != nil {
		log.Fatal("Malformed URL: ", err.Error())
	}

	baseURLWithAddr := net.JoinHostPort(baseURL.Hostname(), strconv.FormatInt(c.Ports, 10))

	// Parameters
	params := url.Values{}
	params.Add("apiKey", c.APIKey)

	var newQuery *url.URL
	newQuery, err = url.Parse(fmt.Sprintf("%v://%v", baseURL.Scheme, baseURLWithAddr))
	if err != nil {
		log.Fatal("Malformed URL: ", err.Error())
	}

	newQuery.RawQuery = params.Encode()
	return newQuery.String()
}

// compareResponse the two responses
func compareResponse(a Response, b Response) bool {
	if a.Direction != b.Direction || a.Password != b.Password || a.Port != b.Port || a.Server != b.Server || a.SourcePort != b.SourcePort || a.Target != b.Target || a.TargetPort != b.TargetPort || a.User != b.User {
		return false
	}
	return true
}

// getUniqueNodes compare two responses and return the nodes that are similar
func getUniqueNodes(a []Response, b []Response) {
	result := make([]Response, 0, 11)
	for _, v := range a {
		exist := false
		for _, w := range b {
			if compareResponse(v, w) {
				exist = true
				break
			}
		}
		if exist {
			result = append(result, v)
		}
	}
	fmt.Println(result) // [F5 F7 C6 G5]
}

// getDistinctNodes compares two response and returns the response in a that are not present in b
func getDistinctNodes(a []Response, b []Response) []Response {
	result := make([]Response, 0, 11)
	for _, v := range a {
		exist := false
		for _, w := range b {
			if compareResponse(v, w) {
				exist = true
			}
		}
		if !exist {
			result = append(result, v)
		}
	}
	return result
}

// getResponse calls the rest endpoint and store the response
func getResponse(url string) []Response {
	log.Printf("Processing %v\n", url)

	var getRestResponse []Response

	request := gorequest.New()
	_, body, _ := request.Get(url).EndStruct(&getRestResponse)

	log.Printf("Response:\n%v\n", string(body))
	return getRestResponse
}

// StartReverseTunnel opens reverse tunnel
func StartReverseTunnel(node Response) {
	addr := fmt.Sprintf("%v:%v:%v", node.SourcePort, node.Target, node.TargetPort)
	host := fmt.Sprintf("%v@%v", node.User, node.Server)

	subProcess := exec.Command("plink", "-ssh", "-N", "-pw", node.Password, "-R", addr, host) //Just for testing, replace with your subProcess

	stdin, err := subProcess.StdinPipe()
	if err != nil {
		fmt.Println(err) //replace with logger, or anything you want
	}
	defer stdin.Close() // the doc says subProcess.Wait will close it, but I'm not sure, so I kept this line

	subProcess.Stdout = os.Stdout
	subProcess.Stderr = os.Stderr

	log.Println(node.ToString())
	if err = subProcess.Start(); err != nil { //Use start, not run
		log.Println("An error occured: ", err) //replace with logger, or anything you want
	}
	io.WriteString(stdin, "y\r\ny\r\ny\r\n")
	io.WriteString(stdin, "y\r\ny\r\ny\r\n")
	io.WriteString(stdin, "y\r\ny\r\ny\r\n")

	subProcess.Wait()
}

// StartForwardTunnel opens forward tunnel
func StartForwardTunnel(node Response) {
	addr := fmt.Sprintf("%v:%v:%v", node.SourcePort, node.Target, node.TargetPort)
	host := fmt.Sprintf("%v@%v", node.User, node.Server)

	subProcess := exec.Command("plink", "-ssh", "-N", "-pw", node.Password, "-L", addr, host) //Just for testing, replace with your subProcess

	stdin, err := subProcess.StdinPipe()
	if err != nil {
		fmt.Println(err) //replace with logger, or anything you want
	}
	defer stdin.Close() // the doc says subProcess.Wait will close it, but I'm not sure, so I kept this line

	subProcess.Stdout = os.Stdout
	subProcess.Stderr = os.Stderr

	log.Println(node.ToString())
	if err = subProcess.Start(); err != nil { //Use start, not run
		log.Println("An error occured: ", err)
	}
	io.WriteString(stdin, "y\r\ny\r\ny\r\n")
	io.WriteString(stdin, "y\r\ny\r\ny\r\n")
	io.WriteString(stdin, "y\r\ny\r\ny\r\n")

	subProcess.Wait()
}

func openTunnels(nodes []Response) {
	if len(nodes) > 0 {
		log.Println("Opening new tunnels in the response that are following: ")
		for _, node := range nodes {
			node.print()
		}
		for _, node := range nodes {
			if node.Direction == "forward" {
				go StartForwardTunnel(node)
			} else if node.Direction == "reverse" {
				go StartReverseTunnel(node)
			}
		}
	} else {
		log.Println("No new tunnels to open.")
	}
}

func closeTunnels(nodes []Response) {
	if len(nodes) > 0 {
		log.Println("Closing old tunnels that are not in the response that are following: ")
		for _, node := range nodes {
			node.print()
		}
	} else {
		log.Println("No tunnels to close.")
	}
}

func main() {
	c = loadConfig("config.json")
	url := buildURL(c)

	startNodes := getResponse(url)
	openTunnels(startNodes)
	log.Println("Waiting for 1 minute till next request")
	time.Sleep(time.Minute * 1)

	for {
		newNodes := getResponse(url)

		// Getting the nodes that are new
		distinctNodes := getDistinctNodes(newNodes, startNodes)
		openTunnels(distinctNodes)

		// Getting the nodes that are not present
		closedNodes := getDistinctNodes(startNodes, newNodes)
		closeTunnels(closedNodes)

		// Updating start nodes to the new nodes
		startNodes = newNodes

		log.Println("Waiting for 1 minute till next request")
		time.Sleep(time.Minute * 1)
	}

	// r := Response{
	// 	Server:     "public.nsplice.com",
	// 	Port:       22,
	// 	User:       "tun-user",
	// 	Password:   "somePassword",
	// 	Direction:  "forward",
	// 	Target:     "localhost",
	// 	TargetPort: 8000,
	// 	SourcePort: 3000,
	// }
	// StartForwardTunnel(r)
}
