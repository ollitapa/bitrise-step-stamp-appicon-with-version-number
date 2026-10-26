Release notes

This file is used as the body of the GitHub release created by the
`github-release-this-step` workflow. Update it before running
`bitrise run release-this-step`.

## What's changed

### 2.0.0

This is a major release: the Step now renders stamps in pure Go with
an embedded font, removes all external tool dependencies, and works
for any project type out of the box.

**Rendering (breaking)**
- Replaced ImageMagick + Ghostscript with a pure Go implementation
  (`fogleman/gg` + an embedded Go Bold font): no `deps` installs, no
  system font probing, identical output on every OS
- Fixed the stamp text rendering at a tiny size (the font-fitting
  measured in fixed-point units instead of pixels) and hanging during
  fitting (non-terminating search)
- Fixed the stamp text sitting at the top of the banner instead of
  being centered: the baseline is now positioned using the font's
  actual ascent/descent
- New default stamp format: `1.2 (123)` instead of `- 1.2(123)`
- Colors accept `#RGB`, `#RRGGBB`, `#RRGGBBAA`, and common color names

**New features**
- Icon directory auto-detection: with an empty `stamp_path_to_icons`
  the Step finds iOS `*.appiconset`, Android `mipmap-*`/`drawable-*`,
  and macOS `*.iconset` directories — iOS, Android, Flutter, and
  React Native projects work with zero configuration
- `stamp_text` input: custom stamp template with `%version%` and
  `%build%` placeholders
- `stamp_position` input: `north`, `south` (default), or `center`
  banner placement
- `stamp_banner_height_ratio` input: banner height as a ratio of the
  icon height (default `0.3`, previously hardcoded)
- New outputs: `stamped_icon_count` and `stamped_icon_paths`

**CI and maintenance**
- Unit tests for color parsing, stamping, template rendering, font
  fitting, text placement, and directory detection
- Bitrise-native CI: lint (build/vet/gofmt), unit tests exported to
  the build's Tests tab via JUnit XML, stamped sample icons deployed
  as build artifacts (default and custom-variant)
- Dependabot for Go modules and GitHub Actions

### 1.2.0

- Font file input (`stamp_font`) with per-OS system font detection
  (macOS/Linux), used as the ImageMagick caption font
- `run-example.sh` local runner script with per-run overridable
  configuration
- README rewrite: local build/run instructions, inputs table, release
  flow documentation
- Bitrise CI runs the step test and deploys stamped icons as build
  artifacts

### 1.1.0 and earlier

- Initial ImageMagick-based implementation stamping
  `version (build number)` on all `.png` icons in a directory
