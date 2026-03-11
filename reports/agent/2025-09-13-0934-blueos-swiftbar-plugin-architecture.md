---
Title: BluOS SwiftBar Plugin Architecture Analysis
Repo: BlueOS
Commit: 365b8e917ed5de9637f5f12dbad9eab54ec362a3
Index: Codanna index empty (0 symbols, 0 files)
Languages: Go
Date: September 13, 2025 at 09:34 AM
Model: claude-sonnet-4-20250514
---

# Code Research Report

## 1. Inputs and Environment

Tools: Direct file reading (Codanna index empty)
Limits: Unknown

## 2. Investigation Path

| Step | Tool         | Input                          | Output summary                           | Artifact          |
|------|--------------|--------------------------------|------------------------------------------|-------------------|
| 1    | get_index_info | n/a                           | Empty index (0 symbols, 0 files)        | see Evidence §5.1 |
| 2    | Read         | main.go, helpers.go, menu.go  | Core Go source files                     | see Evidence §5.2 |
| 3    | Read         | structs.go, README.md, go.mod | Data structures and project metadata     | see Evidence §5.3 |

## 3. Mechanics of the Code

- Entry point: main() function initializes config from .env file in SWIFTBAR_PLUGINS_PATH
- Device discovery: automatic mDNS discovery tries 3 service types (_musc._tcp, _musp._tcp, _mush._tcp), falls back to manual BLUE_URL config
- BitBar integration: builds status line and dropdown menu using johnmccabe/go-bitbar library
- BluOS API interaction: HTTP requests to /Status, /Volume, /Presets endpoints with XML parsing
- Menu architecture: modular design with buildPlayerMenu orchestrating status display, presets, and volume controls
- Error handling: retry logic (3 attempts), timeout management (10s for requests), reachability checks
- Volume control: supports dB-based and percentage-based operations with mute toggle

## 4. Quantified Findings

- Go source files: 4 (main.go, helpers.go, menu.go, structs.go)
- Lines of code: ~107 (main.go) + ~352 (helpers.go) + ~365 (menu.go) + ~85 (structs.go) = ~909 total
- Dependencies: 3 direct (go-bitbar, godotenv, mdns), 6 indirect
- API endpoints: 3 primary (/Status, /Volume, /Presets)
- XML data structures: 3 (StateXML, Presets, VolumeStatus)
- Volume presets: 4 levels (20%, 50%, 80%, 100%)
- Discovery timeout: 5 seconds
- HTTP timeout: 10 seconds with 3 retry attempts
- Plugin refresh interval: 15 seconds (per README)

## 5. Evidence

### 5.1 Codanna Index Status
```
Index contains 0 symbols across 0 files.
Breakdown:
  - Symbols: 0
  - Relationships: 0
```

### 5.2 Main Entry Point
```go
func main() {
	// Initialize configuration
	if m, e := strconv.Atoi(myConfig["MAX"]); e == nil {
		MAX = m
	}

	// Get BluOS device URL (try discovery first, fall back to config)
	bluePlayerUrl, err := getBluOSPlayerURL(myConfig["BLUE_URL"])
}
```
// main.go:36-43

### 5.3 Core Data Structure
```go
type StateXML struct {
	Text    string `xml:",chardata"`
	Etag    string `xml:"etag,attr"`
	Actions struct {
		Text   string `xml:",chardata"`
		Action []struct {
			Text     string `xml:",chardata"`
			Name     string `xml:"name,attr"`
			URL      string `xml:"url,attr"`
			Icon     string `xml:"icon,attr"`
			State    string `xml:"state,attr"`
			AttrText string `xml:"text,attr"`
			Type     string `xml:"type,attr"`
		} `xml:"action"`
	} `xml:"actions,omitempty"`
	Album           string `xml:"album,omitempty"`
	Artist          string `xml:"artist,omitempty"`
	// ... 30+ more fields
}
```
// structs.go:8-58

### 5.4 Device Discovery Function
```go
func discoverBluOSDevices(timeout time.Duration) ([]string, error) {
	log.Printf("Starting BluOS device discovery (timeout: %v)", timeout)

	// Channel to collect discovered services
	entriesCh := make(chan *mdns.ServiceEntry, 10)
	var devices []string
	seen := make(map[string]bool) // Prevent duplicates

	// Browse for BluOS service types
	serviceTypes := []string{"_musc._tcp", "_musp._tcp", "_mush._tcp"}
}
```
// helpers.go:200-209

### 5.5 Menu Builder Architecture
```go
func buildPlayerMenu(app *bitbar.Plugin, bluePlayerUrl string) {
	log.Printf("Building player menu for %s", bluePlayerUrl)
	statusUrl := fmt.Sprintf("%s/Status", bluePlayerUrl)
	presetsUrl := fmt.Sprintf("%s/Presets", bluePlayerUrl)
	volumeUrl := fmt.Sprintf("%s/Volume", bluePlayerUrl)

	submenu := app.NewSubMenu()

	// Process status data and create status bar
	createStatusDisplay(app, submenu, statusUrl, bluePlayerUrl)

	// Add separator
	submenu.Line("---")

	// Add radio presets directly (no header)
	addRadioPresets(submenu, presetsUrl, bluePlayerUrl)

	// Add separator
	submenu.Line("---")

	// Add volume info (no header)
	volStatus := addVolumeInfo(submenu, volumeUrl)

	// Add volume presets (no header)
	addVolumePresets(submenu, bluePlayerUrl, volStatus)

	// Add mute toggle
	addMuteToggle(submenu, bluePlayerUrl, volStatus)
}
```
// menu.go:13-43

## 6. Implications

- Network dependency: Plugin requires multicast/mDNS support (UDP 5353) for auto-discovery
- Polling overhead: 15-second refresh creates ~240 requests/hour to BluOS device
- Memory footprint: XML parsing creates temporary objects for each request cycle
- Error resilience: 3 retry attempts × 10s timeout = up to 30s delay in worst case
- UI responsiveness: SwiftBar integration allows async updates without blocking

## 7. Hidden Patterns

- Volume control supports both dB (-infinity to 0) and percentage (0-100) scales
- SF Symbols integration provides macOS-native iconography (:play.circle.fill:, :speaker.wave.3.fill:)
- Service-specific UI adaptations (AirPlay vs Spotify vs Capture modes)
- Unused discovery capability: discovers multiple devices but only uses first working one
- Extension point: VolumeStatus.Actions field suggests API supports additional operations
- Alternative display modes: Alternate(true) creates Option-key triggered views

## 8. Research Opportunities

- Analyze VolumeStatus.Actions to understand unused control capabilities
- Examine BluOS API PDF documentation for undocumented endpoints
- Test multi-device scenarios to understand selection logic
- Profile memory usage during XML parsing cycles
- Investigate SwiftBar plugin lifecycle and caching mechanisms

## 9. Code Map

| Component              | File              | Line | Purpose                           |
|------------------------|-------------------|------|-----------------------------------|
| `main`                 | `main.go`         | 36   | Entry point and error handling    |
| `StateXML`             | `structs.go`      | 8    | BluOS /Status response parser     |
| `VolumeStatus`         | `structs.go`      | 75   | Volume API response parser        |
| `Presets`              | `structs.go`      | 61   | Radio presets structure           |
| `getBluOSPlayerURL`    | `helpers.go`      | 306  | Device discovery orchestrator     |
| `discoverBluOSDevices` | `helpers.go`      | 200  | mDNS service discovery            |
| `getXML`               | `helpers.go`      | 128  | HTTP client with retry logic      |
| `buildPlayerMenu`      | `menu.go`         | 13   | SwiftBar menu orchestrator        |
| `createStatusDisplay`  | `menu.go`         | 46   | Status line generator             |
| `addVolumeInfo`        | `menu.go`         | 241  | Volume display with controls      |
| `sendVolumeCommand`    | `helpers.go`      | 22   | Volume API interface              |

## 10. Confidence and Limitations

- Architecture understanding: High (direct file access, complete source review)
- API integration patterns: High (BluOS XML responses and HTTP interface documented)
- SwiftBar plugin mechanics: Medium (library usage patterns observed)
- Performance characteristics: Medium (timeout and retry values documented)
- Unknown: Runtime behavior under network instability, actual memory usage patterns

## 11. Footer

GeneratedAt=September 13, 2025 at 09:34 AM  Model=claude-sonnet-4-20250514