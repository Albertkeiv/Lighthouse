# Lighthouse

Lighthouse is a cross-platform utility for managing SSH tunnels and local domain proxies.

This repository currently provides a minimal prototype with the following features:

- Data model for profiles and tunnels.
- Loading and saving profiles from the user's configuration directory (e.g.
  `~/.config/lighthouse/profiles.json`) or the current working directory if no
  user configuration directory can be determined.
- Basic graphical interface using [Fyne](https://fyne.io/) that lists configured profiles.
- Ability to create and save new profiles through the interface.

The project is under active development and does not yet implement the full specification.

## Cross-compilation

This project uses [fyne-cross](https://github.com/fyne-io/fyne-cross) to create binaries for multiple platforms.

### Install fyne-cross

```bash
go install github.com/fyne-io/fyne-cross@latest
```

Docker must be installed and running.

### Build

```bash
./build.sh
```

The build outputs will appear under the `fyne-cross` directory:

- `fyne-cross/linux-amd64`
- `fyne-cross/windows-amd64`
- `fyne-cross/darwin-amd64`

