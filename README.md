<p align="center">
<img style="align: center; padding-left: 10px; padding-right: 10px; padding-bottom: 10px;" width="238px" height="238px" src="./logo.png" />
</p>

<h2 align="center">Pack - packge manager</h2>

[![Generic badge](https://img.shields.io/badge/LICENSE-GPL-orange.svg)](https://fmnx.io/dev/pack/src/branch/main/LICENSE)
[![Generic badge](https://img.shields.io/badge/GITEA-REPO-red.svg)](https://fmnx.io/dev/pack)
[![Generic badge](https://img.shields.io/badge/GITHUB-REPO-white.svg)](https://github.com/fmnx-io/repo)
[![Build Status](https://ci.fmnx.io/api/badges/dev/repo/status.svg)](https://ci.fmnx.io/dev/pack)

Git-based pacman-compatible package manager. Since `go` creators started reusing `git` for in go package management system, the value of decentralized systems shined from another perspective. This package manager is trying to reuse the power of both `git` and `pacman` to become new age way of arch package distribution.

## 🚀 Features:

- Install and update packages using git links
- Create packages compatible with all arch-based distros
- One simple to write file to adapt repo for pack - `pack.yml`
- Create project templates from markdown files

## 💾 Installationion

You can install `pack` on any arch-based distribution using go.

- With go

```sh
sudo pacman -S go
go install fmnx.io/dev/pack@latest
```

## 📄 Usage

