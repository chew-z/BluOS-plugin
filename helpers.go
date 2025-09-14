package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	
	"time"

	"github.com/hashicorp/mdns"
	"github.com/johnmccabe/go-bitbar"
)

const defaultDbStep = 2.0 // Typical dB step for volume up/down

// sendVolumeCommand sends a command to the /Volume endpoint with parameters
func sendVolumeCommand(playerUrl string, params map[string]string) (*VolumeStatus, error) {
	baseURL := fmt.Sprintf("%s/Volume", playerUrl)
	reqURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing base URL %s: %w", baseURL, err)
	}

	query := reqURL.Query()
	for key, value := range params {
		query.Set(key, value)
	}
	reqURL.RawQuery = query.Encode()

	xmlBytes, err := getXML(reqURL.String())
	if err != nil {
		return nil, fmt.Errorf("volume command failed: %w", err)
	}

	var status VolumeStatus
	if err := xml.Unmarshal(xmlBytes, &status); err != nil {
		log.Printf("Failed to parse volume XML: %v\nXML: %s", err, string(xmlBytes))
		return nil, fmt.Errorf("XML parsing error: %w", err)
	}

	return &status, nil
}



// createVolumeCommand creates a bitbar command for volume control operations
func createVolumeCommand(playerUrl string, params map[string]string) bitbar.Cmd {
	baseURL := fmt.Sprintf("%s/Volume", playerUrl)
	reqURL, err := url.Parse(baseURL)
	if err != nil {
		// For a command creation function, log the error but continue with a default
		log.Printf("Error parsing URL %s: %v", baseURL, err)
		return createCommand(baseURL) // Fallback to basic command
	}

	query := reqURL.Query()
	for key, value := range params {
		query.Set(key, value)
	}
	reqURL.RawQuery = query.Encode()

	return bitbar.Cmd{
		Bash:     "curl",
		Params:   []string{"-sf", reqURL.String()},
		Terminal: BoolPointer(false),
		Refresh:  BoolPointer(true),
	}
}



// Db2vol converts dB to volume percentage (0-100)
// This is the inverse of the vol2db function and maintains compatibility
func Db2vol(db float64) float64 {
	return 100.0 * math.Pow(10.0, db/60.0)
}

// tweaked from: https://stackoverflow.com/a/42718113/1170664
func getXML(url string) ([]byte, error) {
	log.Printf("Fetching XML from: %s", url)

	// Set timeout for requests - increased from 5s to 10s for better reliability
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Implement retry logic (3 attempts)
	maxRetries := 3
	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Printf("Attempt %d/%d to fetch from %s", attempt, maxRetries, url)

		resp, err := client.Get(url)
		if err != nil {
			log.Printf("Error connecting to %s (attempt %d/%d): %v", url, attempt, maxRetries, err)
			lastErr = fmt.Errorf("GET error: %v", err)
			if attempt < maxRetries {
				time.Sleep(500 * time.Millisecond) // Short delay between retries
				continue
			}
			break
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("Bad status code from %s (attempt %d/%d): %d", url, attempt, maxRetries, resp.StatusCode)
			lastErr = fmt.Errorf("Status error: %v", resp.StatusCode)
			if attempt < maxRetries {
				time.Sleep(500 * time.Millisecond) // Short delay between retries
				continue
			}
			break
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error reading response body from %s (attempt %d/%d): %v", url, attempt, maxRetries, err)
			lastErr = fmt.Errorf("Read body: %v", err)
			if attempt < maxRetries {
				time.Sleep(500 * time.Millisecond) // Short delay between retries
				continue
			}
			break
		}

		// If we get here, we succeeded
		log.Printf("Successfully retrieved %d bytes from %s on attempt %d/%d", len(data), url, attempt, maxRetries)
		return data, nil
	}

	// If we get here, all attempts failed
	return []byte{}, lastErr
}

func BoolPointer(b bool) *bool {
	return &b
}

// createCommand is a helper to create curl commands for BluOS API endpoints
func createCommand(url string) bitbar.Cmd {
	return bitbar.Cmd{
		Bash:     "curl",
		Params:   []string{"-sf", url},
		Terminal: BoolPointer(false),
		Refresh:  BoolPointer(true),
	}
}

// discoverBluOSDevices discovers BluOS players on the local network using mDNS/Bonjour
// Returns a slice of device URLs (http://ip:port) found on the network
func discoverBluOSDevices(timeout time.Duration) ([]string, error) {
	log.Printf("Starting BluOS device discovery (timeout: %v)", timeout)

	// Channel to collect discovered services
	entriesCh := make(chan *mdns.ServiceEntry, 10)
	var devices []string
	seen := make(map[string]bool) // Prevent duplicates

	// Browse for BluOS service types
	serviceTypes := []string{"_musc._tcp", "_musp._tcp", "_mush._tcp"}

	// Start discovery in a goroutine
	go func() {
		defer close(entriesCh)

		for _, serviceType := range serviceTypes {
			log.Printf("Browsing for service type: %s", serviceType)

			// Create a new query for each service type
			err := mdns.Query(&mdns.QueryParam{
				Service: serviceType,
				Domain:  "local",
				Timeout: timeout / time.Duration(len(serviceTypes)), // Split timeout among service types
				Entries: entriesCh,
			})

			if err != nil {
				log.Printf("Error browsing for %s: %v", serviceType, err)
			}

			// Small delay between queries to avoid overwhelming the network
			time.Sleep(100 * time.Millisecond)
		}
	}()

	// Collect discovered devices with overall timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for {
		select {
		case entry, ok := <-entriesCh:
			if !ok {
				// Channel closed, discovery finished
				log.Printf("Discovery completed. Found %d unique BluOS device(s)", len(devices))
				return devices, nil
			}

			if entry != nil && entry.AddrV4 != nil {
				// Use the IPv4 address found
				deviceURL := fmt.Sprintf("http://%s:%d", entry.AddrV4, entry.Port)

				// Avoid duplicates
				if !seen[deviceURL] {
					log.Printf("Discovered BluOS device: %s (service: %s, hostname: %s)",
						deviceURL, entry.Name, entry.Host)
					devices = append(devices, deviceURL)
					seen[deviceURL] = true
				}
			}
		case <-ctx.Done():
			// Timeout reached
			log.Printf("Discovery timeout reached. Found %d unique BluOS device(s)", len(devices))
			return devices, nil
		}
	}
}

// findValidBluOSDevice discovers BluOS devices and returns the first working one
// It tests each discovered device with a simple /Status call to verify it's accessible
func findValidBluOSDevice(timeout time.Duration) (string, error) {
	devices, err := discoverBluOSDevices(timeout)
	if err != nil {
		return "", fmt.Errorf("device discovery failed: %w", err)
	}

	if len(devices) == 0 {
		return "", fmt.Errorf("no BluOS devices found on network")
	}

	// Test each device to find a working one
	client := &http.Client{Timeout: 3 * time.Second}

	for _, deviceURL := range devices {
		statusURL := fmt.Sprintf("%s/Status", deviceURL)
		log.Printf("Testing BluOS device: %s", statusURL)

		resp, err := client.Get(statusURL)
		if err != nil {
			log.Printf("Device %s unreachable: %v", deviceURL, err)
			continue
		}
		if err := resp.Body.Close(); err != nil {
			log.Printf("Failed to close response body for %s: %v", deviceURL, err)
		}

		if resp.StatusCode == http.StatusOK {
			log.Printf("Found working BluOS device: %s", deviceURL)
			return deviceURL, nil
		}

		log.Printf("Device %s returned status %d", deviceURL, resp.StatusCode)
	}

	return "", fmt.Errorf("no working BluOS devices found (tested %d device(s))", len(devices))
}

// getBluOSPlayerURL returns the BluOS player URL using discovery first, then fallback to env var
func getBluOSPlayerURL(fallbackURL string) (string, error) {
	// Try automatic discovery first (5 second timeout)
	if discoveredURL, err := findValidBluOSDevice(5 * time.Second); err == nil {
		log.Printf("Using discovered BluOS device: %s", discoveredURL)
		return discoveredURL, nil
	} else {
		log.Printf("Auto-discovery failed: %v", err)
	}

	// Fall back to manually configured URL
	if fallbackURL != "" {
		log.Printf("Using configured BluOS device: %s", fallbackURL)
		return fallbackURL, nil
	}

	return "", fmt.Errorf("no BluOS device found via discovery and no BLUE_URL configured")
}

// isDeviceReachable performs a simple network check to see if the device is reachable,
// even if the main API might be having issues
func isDeviceReachable(url string) bool {
	// Try a few different endpoints to increase chances of success
	endpoints := []string{"/Status", "/Volume", "/"}

	for _, endpoint := range endpoints {
		fullURL := fmt.Sprintf("%s%s", url, endpoint)
		log.Printf("Checking device reachability via: %s", fullURL)

		// Use a simple check with long timeout for reachability testing
		client := &http.Client{
			Timeout: 15 * time.Second, // Longer timeout for reachability check
		}

		resp, err := client.Get(fullURL)
		if err == nil {
			log.Printf("Device is reachable via %s (status: %d)", endpoint, resp.StatusCode)
			if closeErr := resp.Body.Close(); closeErr != nil {
				log.Printf("Failed to close response body for %s: %v", fullURL, closeErr)
			}
			return true
		}

		log.Printf("Failed to reach device via %s: %v", endpoint, err)
	}

	log.Printf("Device appears to be completely unreachable")
	return false
}
