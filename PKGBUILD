# PKGBUILD generated by pack.
# More info at: https://fmnx.io/core/pack

pkgname="pack"
pkgver="1"
pkgrel="1"
arch=('i686' 'pentium4' 'x86_64' 'arm' 'armv7h' 'armv6h' 'aarch64' 'riscv64')
url="https://fmnx.io/core/pack"
depends=(
  "git"
  "pacman"
)
makedepends=(
  "go"
)

build() {
  cd ..
  go build -buildvcs=false -o pack .
}

package() {
  cd ..
  install -Dm755 pack $pkgdir/usr/bin/pack
}
