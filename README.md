<p>
  <h3 align="center">Rift</h3>
  <p align="center">LoL Esports in your terminal.</p>
</p>

---
<p align="center">
  <a href="https://github.com/matthieugusmini/lolesport/releases"><img src="https://img.shields.io/github/release/matthieugusmini/lolesport.svg" alt="Latest Release"></a>
  <a href="https://pkg.go.dev/github.com/matthieugusmini/lolesport?tab=doc"><img src="https://godoc.org/github.com/golang/gddo?status.svg" alt="GoDoc"></a>
  <a href="https://github.com/charmbracelet/glow/actions"><img src="https://github.com/charmbracelet/glow/workflows/build/badge.svg" alt="Build Status"></a>

</p>

<p align="center">Never miss a match again, keep track of the results and follow your favorite leagues and players from your terminal!</p>
<p align="center">
  <img src="https://vhs.charm.sh/vhs-159DB9Zm1KES7xyOzKE44E.gif" alt="Made with VHS" width=700>
</p>

## Installation
> [!IMPORTANT]
> For the best experience a [Nerd Font](https://www.nerdfonts.com/) installed and enabled is required.

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

### yum

```bash
# Fedora/RHEL
echo '[fury]
name=Gemfury Private Repo
baseurl=https://yum.fury.io/matthieugusmini/
enabled=1
gpgcheck=0' | sudo tee /etc/yum.repos.d/fury.repo
sudo yum install rift
```

### Go

```bash
go install github.com/matthieugusmini/rift@latest
```

### Docker

```bash
docker run --rm -it ghcr.io/matthieugusmini/rift
```

### Linux packages(`.deb`, `.apk`, `.rpm`)

Download one of the `.deb`, `.rpm` or `.apk` file from the [releases page](https://github.com/matthieugusmini/rift/releases) and install it using your tool of choice.

### Build

```bash
git clone https://github.com/matthieugusmini/rift.git
cd rift
go build
```

## Supported terminals

| Terminal          | Supported | Issues                                             |
|:-----------------:|:---------:|:--------------------------------------------------:|
| WezTerm           | Yes       | None                                               |
| iTerm2            | Yes       | None                                               |
| macOS Terminal    | Yes       | Match borders doesn't render properly              |
| Windows Terminal  | Yes       | Country flags emojis are not supported on Windows  |

## License

[MIT](https://github.com/charmbracelet/glow/raw/master/LICENSE)
