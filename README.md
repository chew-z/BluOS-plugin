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

You need `.env` file at `SWIFTBAR_PLUGINS_PATH` where you specify WiFi network `BLUE_WIFI`and an IP address `BLUE_URL` where your BlueOS device can be found

## Troubleshooting

### Device shows as disconnected but is actually online

If the plugin shows the device as disconnected even though you can access it with curl or the BluOS app, try these solutions:

1. **Restart the plugin**: Click the plugin icon in SwiftBar menu bar and select "Refresh".

2. **Check your `.env` file**: Make sure the `BLUE_URL` setting is correct and has the proper format (e.g., `http://192.168.1.101:11000`).

3. **Check your network**: Ensure your computer is on the same network as the BluOS device. Some WiFi networks may segment devices.

4. **Device may be busy**: The BluOS device may temporarily stop responding to API calls if it's performing updates or processing heavy tasks. The plugin now includes better retry logic and will show different status messages for different connection issues.

5. **Fixed IP address**: Consider setting a fixed/static IP address for your BluOS device in your router settings to prevent IP address changes.

6. **Manual testing**: You can test connectivity manually with curl:
   ```bash
   curl -v http://YOUR_DEVICE_IP:11000/Status
   ```
   
Versions 1.2.0+ include improved error handling and retry logic that should help with intermittent connectivity issues.
