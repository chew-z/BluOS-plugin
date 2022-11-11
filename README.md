# BluOS-plugin

Simple bitbar (but rather modern [Swiftbar](https://github.com/swiftbar/SwiftBar) versions) plugin for managing [BlueOS](https://bluos.net/) devices on MacOS.

[BluOS Controller for Mac](https://www.bluesound.com/downloads/) is Electron based, slow and buggy app in my opinion. This plugin is using small part of [BlueOS API](http://bluos.net/wp-content/uploads/2022/07/BluOS-Custom-Integration-API-v1.5.pdf) in order to limit interaction with BlueOS Controller.

## What is does?

First it check if you are on the same network as your BlueOS device. If you are plugin displays a state of your player (play, paused, stop) and what is playing (radio, tracks).

You can toggle play/pause. You can also see a list of your presets in dropdown menu and start one by clicking on it. If Option key (⌥) is pressed it will reveal current stream quality.

That's it at the moment.

This project is currently work in progress. Yet.

And it is being optimized for my personal use of [BlueNode 2](https://www.bluesound.com/products/node-2/) (streaming radio (mostly SomaFM stations) with TuneIn or Radio Paradise and playing albums wth Amazon Music Unlimited) with fixed audio level.

So some use cases may have not been tested or optimized for.
