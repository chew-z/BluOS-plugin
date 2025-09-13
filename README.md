# BluOS-plugin

Simple bitbar (but rather modern [Swiftbar](https://github.com/swiftbar/SwiftBar) versions) plugin for managing [BlueOS](https://bluos.net/) devices on MacOS.

[BluOS Controller for Mac](https://www.bluesound.com/downloads/) is Electron based, slow and buggy app in my opinion. This plugin is using small part of [BlueOS API](http://bluos.net/wp-content/uploads/2022/07/BluOS-Custom-Integration-API-v1.5.pdf) in order to limit interaction with BlueOS Controller.

## What is does?

First plugin checks if you are on the same network as your BlueOS device. If your device is reachable plugin displays a state of your player (playing, paused, stoped) and what it is playing (radio, tracks).

You can toggle play/pause. You can also see a list of your presets in dropdown menu and start one by clicking on it. If Option key (⌥) is pressed it will reveal current stream quality.

That's it at the moment.

## Some remarks

This project is currently work in progress. Yet.

And it is being optimized for my personal use case of [Bluesound Node](https://www.bluesound.com/products/node/)

-   streaming radio (mostly SomaFM stations) with TuneIn
-   streaming radio with Radio Paradise
-   playing albums with Amazon Music Unlimited
-   audio output level set to fixed

BlueOS can do much more. So some use cases may have not been tested or optimized for.

I am optimizing for later versions of [SwiftBar](https://github.com/swiftbar/SwiftBar) (1.44, 1.5) and [SF Symbols 4](https://developer.apple.com/sf-symbols/)

BlueOS is not the fastest and the plugin is updated every 15 seconds, not everything refreshes instantly so please be patient.

## Location of BlueOS device

The plugin now supports **automatic device discovery** using mDNS/Bonjour. It will automatically find BluOS devices on your local network.

### Configuration Options:

1. **Automatic Discovery (Recommended)**: The plugin will automatically discover BluOS devices on your network using mDNS/Bonjour protocol. No manual configuration required.

2. **Manual Configuration (Fallback)**: If automatic discovery fails, the plugin falls back to manually configured settings in `.env` file at `SWIFTBAR_PLUGINS_PATH`:
   - `BLUE_WIFI` - Your WiFi network name (for display purposes)
   - `BLUE_URL` - Manual IP address of your BluOS device (e.g., `http://192.168.1.101:11000`)

### How Discovery Works:

The plugin searches for BluOS service types (`_musc._tcp`, `_musp._tcp`, `_mush._tcp`) on the local network and automatically connects to the first working device found. This eliminates the need to manually configure IP addresses and handles dynamic IP changes automatically.

## Troubleshooting

### No BluOS device found (with automatic discovery)

If the plugin shows "BluOS Not Found" even though your device is online:

1. **Check network connectivity**: Ensure your Mac and BluOS device are on the same network/subnet.

2. **Multicast traffic**: Ensure your router/firewall allows multicast traffic (UDP port 5353). Many enterprise/guest WiFi networks block mDNS.

3. **Manual test**: You can test mDNS discovery manually using: `dns-sd -B _musc._tcp`

4. **Fallback to manual config**: Add `BLUE_URL` to your `.env` file as a backup if discovery consistently fails.

### Device shows as disconnected but is actually online

If the plugin shows the device as disconnected even though you can access it with curl or the BluOS app:

1. **Restart the plugin**: Click the plugin icon in SwiftBar menu bar and select "Refresh".

2. **Check your `.env` file**: Make sure the `BLUE_URL` setting is correct and has the proper format (e.g., `http://192.168.1.101:11000`).

3. **Check your network**: Ensure your computer is on the same network as the BluOS device. Some WiFi networks may segment devices.

4. **Device may be busy**: The BluOS device may temporarily stop responding to API calls if it's performing updates or processing heavy tasks. The plugin now includes better retry logic and will show different status messages for different connection issues.

5. **Discovery vs manual**: The plugin first tries automatic discovery, then falls back to `BLUE_URL` if configured. Check logs to see which method is being used.

6. **Manual testing**: You can test connectivity manually with curl:
   ```bash
   curl -v http://YOUR_DEVICE_IP:11000/Status
   ```
   
Versions 1.2.0+ include improved error handling and retry logic that should help with intermittent connectivity issues.
