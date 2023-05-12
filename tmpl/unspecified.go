// 2023 FMNX team.
// Use of this code is governed by GNU General Public License.
// Additional information can be found on official web page: https://fmnx.io/
// Contact email: help@fmnx.io

package tmpl

// This file contains unspecified string templates.

const Cobra = `{{if gt (len .Aliases) 0}}Aliases:
{{.NameAndAliases}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}Available Commands:{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
{{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
{{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

Additional Commands:{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
{{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}{{end}}
`

const PKGBUILD = `# PKGBUILD generated by pack.
# More info at: https://fmnx.io/core/pack

pkgname="%s"
pkgdesc="Something awesome"
pkgver="1"
pkgrel="1"
arch="any"
url="%s"

# Edit runtime dependencies
depends=(
  "python"
  "vlc"
)

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
  install -Dm755 example $pkgdir/usr/bin/example
  # Example of dir installation
  cd build/bundle && find . -type f -exec install -Dm755 {} $pkgdir/etc/share/example/{} \; && cd $srcdir/..
}
`

const READMEmd = `

---

### 📦 Install package with [pack](https://fmnx.io/core/pack):

%s
pack install %s
%s
`

const SysCallErr = `

============================

System call failed:

%s

----------------------------

Error occured.

%+v

----------------------------

Output:

%s

============================

`

const PrettyDesc = `
Name        : %s
Version     : %s
Description : %s
Size        : %s
Url         : %s
PackName    : %s
PackVersion : %s
PackBranch  : %s
`
