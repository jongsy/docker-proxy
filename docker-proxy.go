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
			hosts = append(hosts, DockerHost{ContainerName: container.Names[0], ContainerPort: container.Ports[0].PublicPort})
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
		fmt.Println(target)
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
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

func main() {
	port := ":9090"
	dockerHosts := GetDockerHosts()
	targets := []DockerContainerProxyTarget{}
	for _, dockerHost := range dockerHosts {
		target := DockerContainerProxyTarget{Url: url.URL{Scheme: "http", Host: "localhost:" + strconv.FormatInt(dockerHost.ContainerPort, 10)},
			DockerHost: dockerHost,
			HostEntry:  StripSpecialChars(dockerHost.ContainerName),
		}
		targets = append(targets, target)
	}
	//make host to url map
	urlMap := map[string]url.URL{}
	for _, target := range targets {
		urlMap[target.HostEntry] = target.Url
	}

	proxy := NewMultipleHostReverseProxy(urlMap, port)
	log.Fatal(http.ListenAndServe(port, proxy))
}
