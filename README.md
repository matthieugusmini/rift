<p align="center">
  <img src="https://github.com/user-attachments/assets/ce2e7963-420f-4d7a-9264-fd382ff54048" height=240>
  <p align="center">LoL Esports in your terminal.</p>
</p>

---
<p align="center">
  <a href="https://github.com/matthieugusmini/lolesport/releases"><img src="https://img.shields.io/github/release/matthieugusmini/lolesport.svg" alt="Latest Release"></a>
  <a href="https://pkg.go.dev/github.com/matthieugusmini/rift?tab=doc"><img src="https://godoc.org/github.com/golang/gddo?status.svg" alt="GoDoc"></a>
  <a href="https://github.com/charmbracelet/glow/actions"><img src="https://github.com/charmbracelet/glow/workflows/build/badge.svg" alt="Build Status"></a>

</p>

<p align="center">Never miss a match again, keep track of the results and follow your favorite leagues and players from your terminal!</p>
<p align="center">
  <img src="https://vhs.charm.sh/vhs-159DB9Zm1KES7xyOzKE44E.gif" alt="Made with VHS" width=700>
</p>

## Installation
> [!IMPORTANT]
> For the best experience, a [Nerd Font](https://www.nerdfonts.com/) should be installed and enabled, along with a terminal that supports emojis.

### Homebrew tap

```bash
# macOS
brew install matthieugusmini/tap/rift
```

### apt

```bash
# Debian/Ubuntu
echo 'deb [trusted=yes] https://apt.fury.io/matthieugusmini/ /' | sudo tee /etc/apt/sources.list.d/fury.list
sudo apt update
sudo apt install rift
```

### dnf

```bash
# Fedora/RHEL
echo '[fury]
name=Gemfury Private Repo
baseurl=https://yum.fury.io/matthieugusmini/
enabled=1
gpgcheck=0' | sudo tee /etc/yum.repos.d/fury.repo
sudo dnf install rift
```

### Go

```bash
go install github.com/matthieugusmini/rift@latest
```

### Docker

```bash
docker run --rm -it ghcr.io/matthieugusmini/rift
```

### Build

```bash
git clone https://github.com/matthieugusmini/rift.git
cd rift
go build
```

### Linux packages (`.deb`, `.apk`, `.rpm`)

Download one of the `.deb`, `.rpm` or `.apk` file from the [releases page](https://github.com/matthieugusmini/rift/releases) and install it using your tool of choice.

## Supported terminals

| Terminal          | Supported | Issues                                                                                                                                                     |
|:-----------------:|:---------:|:----------------------------------------------------------------------------------------------------------------------------------------------------------:|
| WezTerm           | Yes       | None                                                                                                                                                       |
| iTerm2            | Yes       | None                                                                                                                                                       |
| Alacritty         | Partially | [Doesn't display emojis](https://github.com/alacritty/alacritty/issues/153)                                                                                |
| macOS Terminal    | Partially | Emojis rendering messes up with the UI                                                                                                                     |
| Windows Terminal  | Partially | [Country flags emojis are not supported on Windows](https://answers.microsoft.com/en-us/windows/forum/all/flag-emoji/85b163bc-786a-4918-9042-763ccf4b6c05) |

If your terminal does not appear in the above list, feel free to open a pull request to add it.

## License

[MIT](https://github.com/charmbracelet/glow/raw/master/LICENSE)
