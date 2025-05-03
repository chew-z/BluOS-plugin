#!/bin/bash
echo "Testing BluOS plugin..."
export SWIFTBAR_PLUGINS_PATH="/Users/rrj/Projekty/Go/src/BlueOS"
./bin/blueos.10s.gobin 2>&1 | tee debug.log
