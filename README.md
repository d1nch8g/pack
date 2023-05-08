<p align="center">
<img style="align: center; padding-left: 10px; padding-right: 10px; padding-bottom: 10px;" width="238px" height="238px" src="https://fmnx.io/dancheg97/Pictures/raw/branch/main/pack.png" />
</p>

<h2 align="center">Pack - package manager</h2>

[![Generic badge](https://img.shields.io/badge/LICENSE-GPL-orange.svg)](https://fmnx.io/core/pack/src/branch/main/LICENSE)
[![Generic badge](https://img.shields.io/badge/FMNX-REPO-006db0.svg)](https://fmnx.io/core/pack)
[![Generic badge](https://img.shields.io/badge/CODEBERG-REPO-45a3fb.svg)](https://codeberg.org/fmnx/pack)
[![Generic badge](https://img.shields.io/badge/GITHUB-REPO-white.svg)](https://github.com/fmnx-io/pack)
[![Generic badge](https://img.shields.io/badge/DOCKER-REGISTRY-blue.svg)](https://fmnx.io/core/-/packages/container/pack/latest)
[![Build Status](https://ci.fmnx.io/api/badges/core/pack/status.svg)](https://ci.fmnx.io/core/pack)

Decentralized package manager based on git and pacman. Accumulates power of both `git` and `pacman` to provide easier way to create arch packages and distribute them using git links.

Core features:

- install packages using git links

```sh
pack i aur.archlinux.org/package
```

- use git tags/branches to switch package to some version

```sh
pack i example.com/package@v0.21
```

- verify package installation inside docker container

```sh
docker run --rm -it fmnx.io/core/pack i aur.archlinux.org/yay
```

- generate `PKGBUILD` template with example of file and directory mapping to end system

You can find all commands and description by running `pack -h`.

💿 Single line installation script (for all arch based distributions):

```sh
git clone https://fmnx.io/core/pack && cd pack && makepkg --noconfirm -sfri
```

---

### 🐋 Pack docker:

You can use `pack` with docker:

```
docker run --rm -it fmnx.io/core/pack --help
```

You can use env variables to configure pack behaviour in docker:

- `PACK_REMOVE_GIT_REPOS` - remove git repositories after installation
- `PACK_REMOVE_BUILT_PACKAGES` - don't cache built `.pkg.tar.zst` packages
- `PACK_DISABLE_PRETTYPRINT` - disable colors in CLI output
- `PACK_DEBUG_MODE` - watch every system call execution

Also you can modify `pack` configuration in `~/.pack/config.yml` or mount to container.

---

### 📋 Template example:

Configuration example for flutter project:

```sh
# PKGBUILD generated by pack.
# More info at: https://fmnx.io/core/pack

pkgname="example"
pkgver="1"
pkgrel="1"
arch=('i686' 'pentium4' 'x86_64' 'arm' 'armv7h' 'armv6h' 'aarch64' 'riscv64')
url="https://example.com/owner/repo"

# Edit build dependencies
makedepends=(
  "flutter"
  "clang"
  "cmake"
)

build() {
  cd ..
  # Example build command
  make build
}

package() {
  cd ..
  # Example of file installation
  install -Dm755 example.sh $pkgdir/usr/bin/example
  install -Dm755 example.desktop $pkgdir/usr/share/applications/example.desktop
  install -Dm755 assets/logo.png $pkgdir/usr/share/icons/hicolor/512x512/apps/example.png
  # Example of dir installation
  cd build/linux/x64/release/bundle && find . -type f -exec install -Dm755 {} $pkgdir/usr/share/example/{} \; && cd $srcdir/..
}
```
