# Stamp AppIcon with version number

[![GitHub](https://img.shields.io/github/license/mashape/apistatus.svg)](https://github.com/ollitapa/bitrise-step-stamp-appicon-with-version-number/blob/master/LICENSE)

A [Bitrise](https://www.bitrise.io) Step that stamps a version banner — `version (build number)` — onto the bottom of every `.png` icon in a directory, using [ImageMagick](https://imagemagick.org).

Handy for test, beta, and internal builds: every installed build is instantly recognizable on a device's home screen by its icon.

Although originally built for iOS `AppIcon.appiconset` directories, the Step works with any directory of `.png` icons (Android, web app icons, etc.).

## How to use this Step

Add it to your workflow in the Bitrise Workflow Editor, or reference it directly in your `bitrise.yml`:

```yaml
- git::https://github.com/ollitapa/bitrise-step-stamp-appicon-with-version-number.git@main:
   title: Stamp app icons with version number
   inputs:
   - stamp_path_to_icons: "MyApp/Assets.xcassets/AppIcon.appiconset"
   - stamp_version: "1.2"
   - stamp_build_number: "$BITRISE_BUILD_NUMBER"
```

### Inputs

| Input | Description | Default |
|---|---|---|
| `stamp_path_to_icons` | Path to the directory containing the icons to stamp | (required) |
| `stamp_version` | Version number or string to stamp, e.g. `0.1` or `Dev` | (required) |
| `stamp_build_number` | Build number to stamp | `$BITRISE_BUILD_NUMBER` |
| `foreground_color` | Text color of the stamp (any ImageMagick color) | `white` |
| `background_color` | Banner background color (any ImageMagick color) | `#0008` |
| `stamp_font` | Path to a font file (`.ttf`) for the stamp text; auto-detected if empty | (auto-detect) |

The Step modifies the icon files in place, so run it **before** the build/archive step that consumes the icons.

## How to build and run locally

### Installing Go

The easiest way to install Go on macOS or Linux is with Homebrew (on Linux, [install Homebrew first](https://docs.brew.sh/Homebrew-on-Linux)):

```bash
brew install go
```

Verify the installation:

```bash
go version
```

Alternatively, download an official installer for your platform from [go.dev/dl](https://go.dev/dl/), or follow the [official installation instructions](https://go.dev/doc/install).

### Installing ImageMagick and Ghostscript

```bash
brew install imagemagick ghostscript   # macOS
sudo apt-get install imagemagick ghostscript   # Linux
```

### Running the Step

The easiest way to run the Step locally is the included example script. It has all inputs configured as editable variables at the top — no command-line configuration needed:

```bash
./run-example.sh
```

Open [`run-example.sh`](run-example.sh), edit the values in the *Configuration* section (icon path, version, build number, colors), and run it. It stamps the sample icon in `testing/AppIcon/` with `1.2 (123)` by default. Values can also be overridden per run without editing the file, e.g. `VERSION=2.0 ./run-example.sh`.

Alternatively, build the binary and run it manually:

```bash
go build -o ./bin/step .

stamp_path_to_icons="./testing/AppIcon" \
stamp_version="1.2" \
stamp_build_number="123" \
foreground_color="white" \
background_color="#0008" \
stamp_font="" \
./bin/step
```

Or run the included test workflow, which stamps the sample icon in `testing/AppIcon/` with version `1.2 (123)`:

```bash
bitrise run test
```

Tip: copy the test icon first if you want to keep an unstamped original, as the Step modifies files in place.

## Development

- The Step code is `main.go`; input metadata lives in `step.yml`.
- After changing `step.yml`, audit it with `stepman audit --step-yml ./step.yml`.
- To share a new Step version to the Bitrise Step Library, bump `BITRISE_STEP_VERSION` in `bitrise.yml` and run `bitrise run share-this-step`.

### Contributing

1. Fork this repository and create a branch for your change.
2. Make your changes and test them with `bitrise run test`.
3. Commit and open a Pull Request.

## License

This project is licensed under the terms of the [LICENSE](LICENSE) file.
