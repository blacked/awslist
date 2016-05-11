package main

import (
	"flag"
	"sync"
	"time"
)

var (
	// @readonly
	portMsg     = "Listen port"
	serviceMsg  = "Run as service"
	intervalMsg = "Interval to pool data in seconds"
)

// PrintInstances runs go routines to print instances from all regions within account
func PrintInstances(profile string, regions []string) {
	defer wg.Done()
	for _, region := range regions {
		wg.Add(1)
		go NewAWSList(profile, region).ListInstances("")
	}
}

// getInstances run go routines to print all instances from all regions and all accounts
func getInstances() {
	var regions []string

	// Clear output_buffer
	output_buffer = []string{}

	// Get list of profiles from ~/.aws/config file
	profiles, _ := ListProfiles()

	// Run go routines to print instances
	for _, profile := range profiles {
		// If we didn't load regions already, then fill regions slice
		if len(regions) == 0 {
			r, _ := NewAWSList(profile, "").ListRegions()
			for _, profile := range r {
				regions = append(regions, profile.Region)
			}
		}

		wg.Add(1)
		go PrintInstances(profile, regions)
	}

	// Wait until receive info about all instances
	wg.Wait()

	// Resize and fill screen buffer with output data
	screen_buffer = make([]string, len(output_buffer), (cap(output_buffer)+1)*2)
	copy(screen_buffer, output_buffer)
}

// Continuously pool list of instances from aws.
func runInstancesPoller(ticker *time.Ticker) {
	for range ticker.C {
		// Get list of all instances
		getInstances()
	}
}

var output_buffer []string
var screen_buffer []string
var service *bool
var port *int
var interval *int
var counter int
var wg sync.WaitGroup

func main() {
	// Parse arguments
	port = flag.Int("port", 8080, portMsg)
	service = flag.Bool("service", false, serviceMsg)
	interval = flag.Int("interval", 30, intervalMsg)
	flag.Parse()

	// Get list of instances
	getInstances()

	// If specified service mode, run program as a service, and listen port
	if *service {
		// Each 30 seconds (by default)
		ticker := time.NewTicker(time.Second * time.Duration(*interval))
		go runInstancesPoller(ticker)

		// Run http server on specifig port
		new(HttpServer).Run(*port)
	}
}
