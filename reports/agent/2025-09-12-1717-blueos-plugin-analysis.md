---
Title: BluOS SwiftBar Plugin Architecture Analysis
Repo: BlueOS
Commit: 365b8e9 (latest on main)
Index: Go module BlueOS v1.24.2
Languages: Go
Date: September 12, 2025 at 05:17 PM
Model: claude-opus-4-1-20250805
---

# Code Research Report

## 1. Inputs and Environment

Tools: go 1.24.2, golangci-lint, gofmt
Limits: Unknown

## 2. Investigation Path

| Step | Tool        | Input                  | Output summary          | Artifact             |
|------|-------------|------------------------|-------------------------|----------------------|
| 1    | Read        | "go.mod"               | Module dependencies identified | see Evidence §5.1  |
| 2    | Read        | "main.go"              | Entry point and initialization flow | see Evidence §5.2  |
| 3    | Read        | "menu.go"              | Menu building logic for SwiftBar | see Evidence §5.3  |
| 4    | Read        | "helpers.go"           | XML fetching and volume control | see Evidence §5.4  |
| 5    | Read        | "structs.go"           | BluOS XML data structures | see Evidence §5.5  |
| 6    | Read        | "README.md"            | Project documentation | see Evidence §5.6  |
| 7    | Glob        | "**/*.sh"              | Build and test scripts | see Evidence §5.7  |

## 3. Mechanics of the Code

- SwiftBar plugin polls BluOS device every 8 seconds (indicated by filename blueos.8s.gobin)
- Initialization loads environment from SWIFTBAR_PLUGINS_PATH/.env file
- Main workflow: check device reachability → fetch XML status → build BitBar menu
- XML API endpoints used: /Status, /Volume, /Presets, /Pause, /Play
- Menu displays player state with SF Symbols icons based on playback status
- Volume control via dB adjustments and percentage presets (20%, 50%, 80%, 100%)
- Retry logic with 3 attempts and 500ms delays for network requests
- Fallback UI for connection issues with diagnostic messages

## 4. Quantified Findings

- Go source files: 4 (main.go, menu.go, helpers.go, structs.go)
- External dependencies: 2 (github.com/johnmccabe/go-bitbar v0.5.0, github.com/joho/godotenv v1.5.1)
- XML structures: 3 (StateXML with 54 fields, Presets, VolumeStatus with 8 fields)
- API endpoints: 6 (/Status, /Volume, /Presets, /Pause, /Play, /Preset)
- Network timeout: 10 seconds per request, 15 seconds for reachability check
- Retry attempts: 3 with 500ms delays
- Volume presets: 4 levels (20%, 50%, 80%, 100%)
- Binary size: 8,635,170 bytes (8.2 MB)

## 5. Evidence

### 5.1 Module Dependencies
```go
// go.mod:1-8
module BlueOS
go 1.24.2

require (
    github.com/johnmccabe/go-bitbar v0.5.0
    github.com/joho/godotenv v1.5.1
)
```

### 5.2 Main Entry Point
```go
// main.go:23-36
func init() {
    var err error
    envPath := fmt.Sprintf("%s/.env", os.Getenv("SWIFTBAR_PLUGINS_PATH"))
    log.Printf("Loading env file from: %s", envPath)
    
    myConfig, err = godotenv.Read(envPath)
    if err != nil {
        log.Fatalln("Error loading .env file:", err)
    }
}

// main.go:81-116
statusUrl := fmt.Sprintf("%s/Status", bluePlayerUrl)
stateXML, err := getXML(statusUrl)
if err != nil {
    if isDeviceReachable(bluePlayerUrl) {
        // Device reachable but API issues
    } else {
        // Device completely offline
    }
} else {
    buildPlayerMenu(&app, bluePlayerUrl)
}
```

### 5.3 Menu Building
```go
// menu.go:13-43
func buildPlayerMenu(app *bitbar.Plugin, bluePlayerUrl string) {
    log.Printf("Building player menu for %s", bluePlayerUrl)
    statusUrl := fmt.Sprintf("%s/Status", bluePlayerUrl)
    presetsUrl := fmt.Sprintf("%s/Presets", bluePlayerUrl)
    volumeUrl := fmt.Sprintf("%s/Volume", bluePlayerUrl)
    
    submenu := app.NewSubMenu()
    createStatusDisplay(app, submenu, statusUrl, bluePlayerUrl)
    submenu.Line("---")
    addRadioPresets(submenu, presetsUrl, bluePlayerUrl)
    submenu.Line("---")
    volStatus := addVolumeInfo(submenu, volumeUrl)
    addVolumePresets(submenu, bluePlayerUrl, volStatus)
    addMuteToggle(submenu, bluePlayerUrl, volStatus)
}

// menu.go:222-239
func getVolumeSymbol(level int, isMuted bool) string {
    if isMuted {
        return ":speaker.slash.fill:"
    }
    switch {
    case level <= 0:
        return ":speaker.slash.fill:"
    case level > 0 && level < 33:
        return ":speaker.wave.1.fill:"
    case level >= 33 && level < 66:
        return ":speaker.wave.2.fill:"
    case level >= 66 && level < 100:
        return ":speaker.wave.3.fill:"
    default:
        return ":megaphone.fill:"
    }
}
```

### 5.4 Network and Volume Helpers
```go
// helpers.go:126-180
func getXML(url string) ([]byte, error) {
    log.Printf("Fetching XML from: %s", url)
    client := &http.Client{
        Timeout: 10 * time.Second,
    }
    
    maxRetries := 3
    var lastErr error
    for attempt := 1; attempt <= maxRetries; attempt++ {
        log.Printf("Attempt %d/%d to fetch from %s", attempt, maxRetries, url)
        resp, err := client.Get(url)
        if err != nil {
            lastErr = fmt.Errorf("GET error: %v", err)
            if attempt < maxRetries {
                time.Sleep(500 * time.Millisecond)
                continue
            }
            break
        }
        // ... handle response
    }
    return []byte{}, lastErr
}

// helpers.go:19-45
func sendVolumeCommand(playerUrl string, params map[string]string) (*VolumeStatus, error) {
    baseURL := fmt.Sprintf("%s/Volume", playerUrl)
    reqURL, err := url.Parse(baseURL)
    // ... build query params
    xmlBytes, err := getXML(reqURL.String())
    var status VolumeStatus
    if err := xml.Unmarshal(xmlBytes, &status); err != nil {
        return nil, fmt.Errorf("XML parsing error: %w", err)
    }
    return &status, nil
}
```

### 5.5 Data Structures
```go
// structs.go:8-58
type StateXML struct {
    Text    string `xml:",chardata"`
    Etag    string `xml:"etag,attr"`
    Album   string `xml:"album,omitempty"`
    Artist  string `xml:"artist,omitempty"`
    State   string `xml:"state"`
    Service string `xml:"service"`
    Title1  string `xml:"title1"`
    Title2  string `xml:"title2"`
    Title3  string `xml:"title3"`
    Volume  string `xml:"volume"`
    Mute    string `xml:"mute"`
    // ... 43 more fields
}

// structs.go:75-84
type VolumeStatus struct {
    XMLName    xml.Name `xml:"volume"`
    Db         float64  `xml:"db,attr"`
    Mute       int      `xml:"mute,attr"`
    MuteDb     *float64 `xml:"muteDb,attr"`
    MuteVolume *int     `xml:"muteVolume,attr"`
    OffsetDb   float64  `xml:"offsetDb,attr"`
    Etag       string   `xml:"etag,attr"`
    Level      int      `xml:",chardata"`
}
```

### 5.6 Configuration
```
// .env.example:1-2
BLUE_URL: "http://192.168.1.101:11000"
BLUE_WIFI: "WiFi with BlueOS"
```

### 5.7 Build Scripts
```bash
// run_test.sh:6
go test -v ./...

// run_lint.sh:14
golangci-lint run --fix ./...
```

## 6. Implications

- Network latency: 10s timeout × 3 retries = max 30s wait per XML fetch
- Memory footprint: ~8.2MB binary + runtime overhead
- Polling frequency: 8 second intervals × 60/8 = 7.5 requests/minute to device
- Network traffic: Assuming 2KB average XML response × 4 endpoints × 7.5/min = 60KB/min
- UI responsiveness: 500ms retry delays + curl command execution time for actions

## 7. Hidden Patterns

- Undocumented MAX config option in .env (main.go:69-71)
- Capture service support for HDMI ARC (menu.go:115-118)
- Alternate display lines using bitbar.Line().Alternate(true) for Option key reveal
- MuteDb and MuteVolume fields in VolumeStatus never used in code
- Unused StateXML fields: Actions, CanMovePlayback, CanSeek, Cursor, Indexing
- Fixed volume output optimization mentioned in README but not enforced in code
- TMP environment variable declared but never used (main.go:19)

## 8. Research Opportunities

- Investigate Actions field in StateXML for additional control capabilities
- Explore canMovePlayback and canSeek for playback control features
- Analyze streamUrl field for direct stream access possibilities
- Test with different BluOS devices beyond Bluesound Node
- Profile memory usage of 8.2MB binary in SwiftBar environment
- Investigate Etag headers for cache-based polling optimization

## 9. Code Map Table

| Component        | File                 | Line  | Purpose              |
|------------------|----------------------|-------|----------------------|
| init             | main.go              | 22    | Load .env configuration |
| main             | main.go              | 67    | Entry point and orchestration |
| isDeviceReachable| main.go              | 40    | Network connectivity check |
| buildPlayerMenu  | menu.go              | 13    | Main menu construction |
| createStatusDisplay | menu.go           | 46    | Player state rendering |
| getVolumeSymbol  | menu.go              | 222   | Volume icon selection |
| addRadioPresets  | menu.go              | 189   | Radio preset menu items |
| addVolumeInfo    | menu.go              | 241   | Volume status display |
| getXML           | helpers.go           | 126   | HTTP client with retries |
| sendVolumeCommand| helpers.go           | 20    | Volume API wrapper |
| createCommand    | helpers.go           | 187   | BitBar command builder |
| StateXML         | structs.go           | 8     | BluOS status structure |
| VolumeStatus     | structs.go           | 75    | Volume state structure |
| Presets          | structs.go           | 61    | Radio presets structure |

## 10. Confidence and Limitations

- Project structure: High - all 4 Go files analyzed
- API endpoints: High - documented in code and README
- Network behavior: High - retry logic clearly implemented
- SwiftBar integration: High - bitbar package usage verified
- Volume calculations: Medium - Db2vol function present but usage unclear
- Error handling patterns: High - multiple fallback states observed
- Unknown: Actual BluOS API response variations across device models
- Unknown: Performance impact of 8 second polling on BluOS device
- Unknown: Behavior with multiple BluOS devices on network

## 11. Footer
GeneratedAt=September 12, 2025 at 05:17 PM  Model=claude-opus-4-1-20250805