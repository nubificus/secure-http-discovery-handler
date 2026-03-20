package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/nubificus/go-akri/pkg/pb"
	verify "github.com/nubificus/secure-http-discovery-handler/internal/verify"
	"github.com/rs/zerolog/log"

	"github.com/smirzaei/parallel"
	"gopkg.in/yaml.v3"
)

const maxConcurrency = 10

type Device struct {
	Hostname   string
	Discovered bool
	Info       DeviceInfo
}

type DeviceInfo struct {
	Device      string `json:"device"`
	Application string `json:"application"`
	Version     string `json:"version"`
}

type DiscoveryDetails struct {
	IPStart     string `yaml:"ipStart"`
	IPEnd       string `yaml:"ipEnd"`
	Application string `yaml:"applicationType"`
	Secure      bool   `yaml:"secure"`
}

func (d DiscoveryDetails) toIPList() ([]string, error) {
	startIP := net.ParseIP(d.IPStart).To4()
	endIP := net.ParseIP(d.IPEnd).To4()
	if startIP == nil || endIP == nil {
		return nil, fmt.Errorf("invalid IPv4 address")
	}
	for i := 0; i < 3; i++ {
		if startIP[i] != endIP[i] {
			return nil, fmt.Errorf("start and end IPs must be in the same /24 subnet")
		}
	}
	var ips []string
	for i := startIP[3]; i <= endIP[3]; i++ {
		ip := net.IPv4(startIP[0], startIP[1], startIP[2], i).String()
		ips = append(ips, ip)
	}
	return ips, nil
}

func Discovery(details string) ([]*pb.Device, error) {
	// log.Println("Discovery called with details:", details)
	log.Info().Msgf("Discovery called with details: %s", details)

	// Unmarshal the YAML string into a DiscoveryDetails struct
	var discoveryDetails DiscoveryDetails
	err := yaml.Unmarshal([]byte(details), &discoveryDetails)
	if err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal discovery details")
		return nil, err
	}
	log.Info().Msgf("Unmarshalled discovery details: %+v", discoveryDetails)

	// Convert the IP range to a list of IP addresses
	ipAddresses, err := discoveryDetails.toIPList()
	if err != nil {
		log.Error().Err(err).Msg("Error converting discovery details to IP list")
		return nil, err
	}

	// Use parallel processing to scan devices
	resultIPs := parallel.MapLimit(ipAddresses, maxConcurrency, func(ip string) Device {
		res := scanDevice(ip, discoveryDetails.Application, discoveryDetails.Secure)
		return res
	})

	// Filter out devices that were discovered
	// and convert them to the protobuf format
	var discoveredDevices []*pb.Device
	for _, device := range resultIPs {
		if device.Discovered {
			discoveredDevices = append(discoveredDevices, &pb.Device{
				Id: device.Hostname,
				Properties: map[string]string{
					"AKRI_HTTP":        "http",
					"HOST_ENDPOINT":    device.Hostname,
					"APPLICATION_TYPE": device.Info.Application,
					"DEVICE":           device.Info.Device,
					"VERSION":          device.Info.Version,
				},
				Mounts:      []*pb.Mount{},
				DeviceSpecs: []*pb.DeviceSpec{},
			})
		}
	}

	var logMessage string
	if len(discoveredDevices) == 0 {
		logMessage = "No devices discovered"
	} else {
		var i int
		logMessage = ""
		for _, device := range resultIPs {
			if !device.Discovered {
				continue
			}
			line := fmt.Sprintf("%d: %s - %s - %s@%s", i, device.Hostname, device.Info.Device, device.Info.Application, device.Info.Version)
			logMessage += line + ", "
			i++
		}
	}

	log.Info().Msgf("Discovered devices: %s", logMessage)
	return discoveredDevices, nil
}

func scanDevice(ip string, expected string, secure bool) Device {
	dev := Device{
		Hostname:   ip,
		Discovered: false,
	}
	url := fmt.Sprintf("http://%s/info", ip)
	client := http.Client{
		Timeout: 3 * time.Second,
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}
	resp, err := client.Get(url)
	if err != nil {
		// TODO: Use log levels
		log.Trace().Err(err).Msgf("Error fetching %s", url)
		return dev
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msgf("Error reading response body from %s", url)
		return dev
	}

	var deviceInfo DeviceInfo
	err = json.Unmarshal(body, &deviceInfo)
	if err != nil {
		log.Error().Err(err).Msgf("Error parsing JSON from %s", url)
		return dev
	}

	if deviceInfo.Application != expected {
		log.Info().Msgf("Device %s does not match expected application type: %s", ip, deviceInfo.Application)
		return dev
	}

	if secure {
		trusted, err := verify.VerifyDevice(ip, AttestationServer)
		if err != nil {
			log.Error().Err(err).Msgf("Error verifying device %s", ip)
			return dev
		}
		if !trusted {
			log.Info().Msgf("Device %s is not trusted", ip)
			return dev
		}
	}
	dev.Discovered = true
	dev.Info = deviceInfo
	log.Info().Msgf("Discovered device %s: %+v", ip, deviceInfo)
	return dev
}
