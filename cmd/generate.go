package cmd

import (
	"fmt"
	"strings"

	"fmnx.io/dev/pack/system"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(genCmd)
}

var genCmd = &cobra.Command{
	Use:     "generate",
	Aliases: []string{"gen"},
	Short:   "📋 generate PKGBUILD template for your repository",
	Long: `📋 generate .pack.yml and update README.md

This command will generate .pack.yml template and add some lines to README.md
to provide information about installation with pack.`,
	Run: Gen,
}

const (
	pkgbuildTemplate = `# PKGBUILD generated by pack.
# More info at: https://fmnx.io/dev/pack

pkgname="%s"
pkgver="1"
pkgrel="1"
arch=('i686' 'pentium4' 'x86_64' 'arm' 'armv7h' 'armv6h' 'aarch64' 'riscv64')
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
  # Edit your build scripts after this line
  make build
}

package() {
  cd ..
  # Edit file mapping in end system
  declare -A filemap=(
    ["example.sh"]="/usr/bin/example"
    ["example.desktop"]="/usr/share/applications/example.desktop"
    ["logo.png"]="/usr/share/icons/hicolor/512x512/apps/example.png"
  )
  # Edit directory mapping in end system
  declare -A dirmap=(
    ["build/dir"]="/usr/share/example"
  )
  
  # Process of file installation
  for file in "${!filemap[@]}"
  do 
    install -Dm755 $file $pkgdir${filemap[$file]}
  done
  
  # Process of directory installation
  for dir in "${!dirmap[@]}"
  do 
    cd $dir
    find . -type f -exec install -Dm755 {} $pkgdir${dirmap[$dir]}/{} \;
    cd $srcdir/..
  done
}
`
	readmeTemplate = `

---

### 📦 Install package with [pack](https://fmnx.io/dev/pack):

%s
pack get %s
%s
`
)

func Gen(cmd *cobra.Command, args []string) {
	info := GetInstallLink()
	WritePackageBuild(info)
	ModifyReadmeFile(info.Link)
	GreenPrint("Updated files: ", "PKGBUILD README.md")
}

type PackageInfo struct {
	Name string
	Link string
}

func GetInstallLink() PackageInfo {
	link, err := system.SystemCallf("git config --get remote.origin.url")
	CheckErr(err)
	link = strings.Trim(link, "\n")
	link = strings.ReplaceAll(link, "https://", "")
	link = strings.ReplaceAll(link, "git@", "")
	link = strings.ReplaceAll(link, ":", "/")
	link = strings.ReplaceAll(link, ".git", "")
	splt := strings.Split(link, "/")
	return PackageInfo{
		Name: splt[len(splt)-1],
		Link: "https://" + link,
	}
}

func WritePackageBuild(i PackageInfo) {
	tmpl := fmt.Sprintf(pkgbuildTemplate, i.Name, i.Link)
	err := system.WriteFile("PKGBUILD", tmpl)
	CheckErr(err)
}

func ModifyReadmeFile(link string) {
	insatllMd := fmt.Sprintf(readmeTemplate, "```", link, "```")
	system.AppendToFile("README.md", insatllMd)
}
