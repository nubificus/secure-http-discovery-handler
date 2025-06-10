package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/smirzaei/parallel"
)

type Response struct {
	Device      string `json:"device"`
	Application string `json:"application"`
	Version     string `json:"version"`
}

func main() {
	// Read os.Args to get IP range
	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Println("Usage: scan-tool <ip-range> <debug>")
		os.Exit(1)
	}
	ipRange := args[0]
	debug := false
	if len(args) > 1 && args[1] == "debug" {
		debug = true
	}
	// ipStart, ipEnd :=
	parts := strings.Split(ipRange, "-")
	if len(parts) != 2 {
		fmt.Println("Invalid IP range format. Use <start-ip>-<end-ip>.")
		os.Exit(1)
	}
	ipStart := parts[0]
	ipEnd := parts[1]
	fmt.Println("Scanning IP range:", ipStart, "to", ipEnd)
	if !isValidIP(ipStart) || !isValidIP(ipEnd) {
		fmt.Println("Invalid IP address format.")
		os.Exit(1)
	}

	ips, err := getIPList(ipStart, ipEnd)
	if err != nil {
		fmt.Printf("Error generating IP list: %v\n", err)
		os.Exit(1)
	}

	discoverChan := make(chan struct {
		Device      string
		Application string
		Version     string
		IP          string
	})

	messageChan := make(chan string)

	go func() {
		for message := range messageChan {
			if debug {
				fmt.Println("Message:", message)
			}
		}
	}()

	go func() {
		for device := range discoverChan {
			fmt.Printf("IP: %s, Discovered device: %s, Application: %s, Version: %s\n",
				device.IP, device.Device, device.Application, device.Version)
		}
	}()

	maxConcurrency := 50
	parallel.ForEachLimit(ips, maxConcurrency, func(ip string) {
		endpoint := fmt.Sprintf("http://%s/info", ip)
		client := &http.Client{
			Timeout: 3 * time.Second, // Set a timeout for the request
			Transport: &http.Transport{
				DisableKeepAlives: true, // Disable keep-alive connections
			},
		}
		resp, err := client.Get(endpoint)
		if err != nil {
			messageChan <- fmt.Sprintf("Error fetching %s: %v", endpoint, err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			messageChan <- fmt.Sprintf("Received non-200 response from %s: %d", endpoint, resp.StatusCode)
			return
		}
		var deviceInfo Response
		if err := json.NewDecoder(resp.Body).Decode(&deviceInfo); err != nil {
			messageChan <- fmt.Sprintf("Error decoding response from %s: %v", endpoint, err)
			return
		}
		discoverChan <- struct {
			Device      string
			Application string
			Version     string
			IP          string
		}{
			Device:      deviceInfo.Device,
			Application: deviceInfo.Application,
			Version:     deviceInfo.Version,
			IP:          ip,
		}
	})
	close(discoverChan)
}

func isValidIP(ip string) bool {
	ipv4 := net.ParseIP(ip).To4()
	return ipv4 != nil
}

func getIPList(ipStart, ipEnd string) ([]string, error) {
	startIP := net.ParseIP(ipStart).To4()
	endIP := net.ParseIP(ipEnd).To4()
	if startIP == nil || endIP == nil {
		return nil, fmt.Errorf("invalid IPv4 address")
	}
	for i := 0; i < 3; i++ {
		if startIP[i] != endIP[i] {
			return nil, fmt.Errorf("start and end IPs must be in the same /24 subnet")
		}
	}
	var ips []string
	start := int(startIP[3])
	end := int(endIP[3])
	for i := start; i <= end; i++ {
		ip := fmt.Sprintf("%d.%d.%d.%d", int(startIP[0]), int(startIP[1]), int(startIP[2]), i)
		ips = append(ips, ip)
	}
	return ips, nil
}
