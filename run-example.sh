#!/usr/bin/env bash
#
# Example script for running the Step locally.
#
# Edit the variables below and run:  ./run-example.sh
# Icon files are modified in place, so run this against a copy of your icons.
#
set -euo pipefail

cd "$(dirname "$0")"

# ---------------------------------------------------------------------------
# Configuration — edit these values for your project
# ---------------------------------------------------------------------------

# Directory containing the .png icons to stamp.
# For an iOS app this is typically something like:
#   MyApp/Assets.xcassets/AppIcon.appiconset
ICON_PATH="${ICON_PATH:-./testing/AppIcon}"

# Version number or string to stamp, e.g. "1.2" or "Dev"
VERSION="${VERSION:-1.2}"

# Build number to stamp
BUILD_NUMBER="${BUILD_NUMBER:-123}"

# Stamp colors — #RGB, #RRGGBB, #RRGGBBAA, or a common color name
FOREGROUND_COLOR="${FOREGROUND_COLOR:-white}"
BACKGROUND_COLOR="${BACKGROUND_COLOR:-#0008}"

# ---------------------------------------------------------------------------
# Run the Step
# ---------------------------------------------------------------------------

stamp_path_to_icons="$ICON_PATH" \
stamp_version="$VERSION" \
stamp_build_number="$BUILD_NUMBER" \
foreground_color="$FOREGROUND_COLOR" \
background_color="$BACKGROUND_COLOR" \
go run .
