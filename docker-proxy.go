package main

import (
	"fmt"
	"github.com/fsouza/go-dockerclient"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strconv"
	"flag"
	"io/ioutil"
	"strings"
	"os"
)

func GetDockerHosts() []DockerHost {
	endpoint := "unix:///var/run/docker.sock"
	client, err := docker.NewClient(endpoint)
	if err != nil {
		panic(err)
	}
	containers, err := client.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		panic(err)
	}
	hosts := []DockerHost{}
	for _, container := range containers {

		if len(container.Ports) > 0 && len(container.Names) > 0 {
			for _, port := range container.Ports {
				if port.PrivatePort == 80 || port.PrivatePort == 8080 {
					fmt.Println(port)
					hosts = append(hosts, DockerHost{ContainerName: container.Names[0], ContainerPort: port.PublicPort})
				}
			}
		}
	}
	return hosts
}

type DockerHost struct {
	ContainerName string
	ContainerPort int64
}

type DockerContainerProxyTarget struct {
	Url        url.URL
	DockerHost DockerHost
	HostEntry  string
}

func (d *DockerContainerProxyTarget) setUrl(url url.URL) {
	d.Url = url
}

func (d *DockerContainerProxyTarget) setDockerHost(dockerHost DockerHost) {
	d.DockerHost = dockerHost
}

func NewMultipleHostReverseProxy(urlMap map[string]url.URL, port string) *httputil.ReverseProxy {
	director := func(req *http.Request) {
		re := regexp.MustCompile(`:[0-9]+`)
		host := re.ReplaceAllString(req.Host, "")
		target := urlMap[host]
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		fmt.Println(req)
	}
	return &httputil.ReverseProxy{Director: director}
}

func StripSpecialChars(input string) string {
	reg, err := regexp.Compile("[^a-zA-Z0-9_]+")
	if err != nil {
		log.Fatal(err)
	}
	processedString := reg.ReplaceAllString(input, "")
	return processedString
}

type HostInterface interface {
	Add(host string)
}

type ShittyHost struct {

}

func (s ShittyHost) Add(host string)  {
	hostsPath := "/etc/hosts"
	file, err := ioutil.ReadFile(hostsPath)
	if err != nil {
		//return err
	}


	contents := string(file)
	rows := strings.Split(contents, "\n")
	hostEntry :=  "127.0.0.1 " + host + " ###docker-proxy"
	inHost := false
	for _, row := range rows {
		if row ==  hostEntry {
			inHost = true
		}
	}
	if inHost == false {
		f, err := os.OpenFile(hostsPath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		if err != nil {
			fmt.Println(err)
			//return err
		}
		defer f.Close()

		_, err = f.WriteString(hostEntry + "\n")
		if err != nil {
			fmt.Println(err)
			//return err
		}
		//return nil
	}
}

func addHostToFile(s HostInterface, host string) {
	s.Add(host)
}

func main() {

	portFlag := flag.Int("port", 9090, "port. default: 9090")
	flag.Parse()
	port := ":" + strconv.Itoa(*portFlag)

	dockerHosts := GetDockerHosts()
	targets := []DockerContainerProxyTarget{}
	for _, dockerHost := range dockerHosts {
			target := DockerContainerProxyTarget{Url: url.URL{Scheme: "http", Host: "localhost:" + strconv.FormatInt(dockerHost.ContainerPort, 10)},
				DockerHost: dockerHost,
				//HostEntry: StripSpecialChars(dockerHost.ContainerName) + strconv.FormatInt(dockerHost.ContainerPort, 10),
				HostEntry: StripSpecialChars(dockerHost.ContainerName),
			}
			targets = append(targets, target)
	}

	urlMap := map[string]url.URL{}
	var s HostInterface = ShittyHost{}
	for _, target := range targets {
		fmt.Println("http://" + target.HostEntry + port, " --> ", target.Url.Host)
		urlMap[target.HostEntry] = target.Url
		addHostToFile(s, target.HostEntry)
	}

	proxy := NewMultipleHostReverseProxy(urlMap, port)
	log.Fatal(http.ListenAndServe(port, proxy))
}
