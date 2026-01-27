#!/bin/bash

# Architecture detection script
# Returns the current system architecture in a format compatible with Go's GOARCH

set -e

# Get the machine architecture using uname
ARCH=$(uname -m)

# Map to Go architecture names
case $ARCH in
    "x86_64")
        echo "amd64"
        ;;
    "aarch64")
        echo "arm64"
        ;;
    "armv7l"|"armv7")
        echo "arm"
        ;;
    "i386"|"i686")
        echo "386"
        ;;
    *)
        echo "amd64"  # Default fallback
        ;;
esac