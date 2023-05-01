package cmd

import (
	"fmt"
	"os"
	"strings"

	"fmnx.io/dev/pack/core"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type PackageInfo struct {
	FullName  string
	ShortName string
	HttpsLink string
	Version   string
	Owner     string
	IsPacman  bool
}

type PackYml struct {
	RunDeps      []string          `yaml:"run-deps"`
	BuildDeps    []string          `yaml:"build-deps"`
	BuildScripts []string          `yaml:"scripts"`
	PackMap      map[string]string `yaml:"mapping"`
}

type PackMap map[string]string

var (
	depsTmpl     = "depends=(\n  \"%s\"\n)"
	makedepsTmpl = "makedepends=(\n  \"%s\"\n)"
	pkgbuildTmpl = `# PKGBUILD generated by pack.
# More info at: https://fmnx.io/dev/pack

pkgname="%s"
pkgver="%s"
pkgrel="1"
arch=('i686' 'pentium4' 'x86_64' 'arm' 'armv7h' 'armv6h' 'aarch64' 'riscv64')
url="%s"
%s
%s
package() {
  cd ..
  %s
}`
)

func init() {
	rootCmd.AddCommand(getCmd)
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "📥 install new packages",
	Run:   Get,
}

func Get(cmd *cobra.Command, pkgs []string) {
	if len(pkgs) != 0 {
		err := core.PrepareDir(cfg.RepoCacheDir)
		CheckErr(err)
		err = core.PrepareDir(cfg.PackageCacheDir)
		CheckErr(err)
	}
	for _, pkg := range pkgs {
		info := EjectInfo(pkg)
		if CheckIfInstalled(info) {
			YellowPrint("Package installed, skipping: ", info.FullName)
			continue
		}
		if info.IsPacman {
			BluePrint("Installing package with pacman: ", info.FullName)
			out, err := core.SystemCall("sudo pacman --noconfirm -Sy " + pkg)
			if err != nil {
				fmt.Println("Pacman output: ", out)
			}
			CheckErr(err)
			GreenPrint("Installed: ", info.FullName+" - OK")
			continue
		}
		PrepareRepo(info)
		packyml := ReadPackYml(info)
		allDeps := append(packyml.RunDeps, packyml.BuildDeps...)
		pacmanPkgs, packPkgs := SplitDependencies(allDeps)
		ResolvePacmanDeps(pacmanPkgs)
		Get(cmd, packPkgs)
		BuildPackage(info, packyml)
		GeneratePkgbuild(info, packyml)
		InstallPackage()
		AddToMapping(info)
		CleanGitDir()
		GreenPrint("Package installed: ", info.FullName)
	}
	lf.Unlock()
}

func EjectInfo(pkg string) PackageInfo {
	if !strings.Contains(pkg, ".") {
		return PackageInfo{
			FullName:  pkg,
			ShortName: pkg,
			IsPacman:  true,
		}
	}
	fullName := strings.Split(pkg, "@")[0]
	httpslink := "https://" + fullName
	split := strings.Split(httpslink, "/")
	shortname := split[len(split)-1]
	owner := strings.Join(split[0:len(split)-1], "/")
	version := ""
	if len(strings.Split(pkg, "@")) == 1 {
		version = GetDefaultBranch(httpslink)
		if strings.Contains(version, "redirecting to") {
			RedPrint("adress mismatch (redirected): ", httpslink)
			os.Exit(1)
		}
	} else {
		version = strings.Split(pkg, "@")[1]
	}
	return PackageInfo{
		FullName:  fullName,
		ShortName: shortname,
		HttpsLink: httpslink,
		Version:   version,
		Owner:     owner,
		IsPacman:  false,
	}
}

func GetDefaultBranch(link string) string {
	out, err := core.SystemCallf("git remote show %s | sed -n '/HEAD branch/s/.*: //p'", link)
	CheckErr(err)
	return strings.Trim(out, "\n")
}

func CheckIfInstalled(i PackageInfo) bool {
	mp := ReadMapping()
	if _, packageExists := mp[i.FullName]; packageExists {
		return true
	}
	_, err := core.SystemCall("pacman -Q " + i.ShortName)
	return err == nil
}

func ReadMapping() PackMap {
	_, err := os.Stat(cfg.MapFile)
	if err != nil {
		core.AppendToFile(cfg.MapFile, "")
		return PackMap{}
	}
	b, err := os.ReadFile(cfg.MapFile)
	CheckErr(err)
	var mapping PackMap
	err = yaml.Unmarshal(b, &mapping)
	CheckErr(err)
	return mapping
}

func PrepareRepo(i PackageInfo) {
	CheckErr(os.Chdir(cfg.RepoCacheDir))
	BluePrint("Cloning repository: ", i.HttpsLink)
	out, err := core.SystemCallf("git clone %s", i.HttpsLink)
	CheckErr(os.Chdir(i.ShortName))
	if strings.Contains(out, "already exists and is not an empty directory") {
		YellowPrint("Repository exists: ", "pulling changes...")
		ExecuteCheck("git pull")
		GreenPrint("Changes pulled: ", "success")
		err = nil
	}
	CheckErr(err)
	BluePrint("Switching repo to version: ", i.Version)
	ExecuteCheck("git checkout " + i.Version)
}

func ReadPackYml(i PackageInfo) PackYml {
	b, err := os.ReadFile(".pack.yml")
	CheckErr(err)
	var packyml PackYml
	err = yaml.Unmarshal(b, &packyml)
	CheckErr(err)
	return packyml
}

func SplitDependencies(deps []string) ([]string, []string) {
	var pacmandeps []string
	var packdeps []string
	for _, dep := range deps {
		if strings.Contains(dep, ".") {
			packdeps = append(packdeps, dep)
			continue
		}
		pacmandeps = append(pacmandeps, dep)
	}
	return pacmandeps, packdeps
}

func ResolvePacmanDeps(pkgs []string) {
	for _, pkg := range pkgs {
		_, err := core.SystemCall("pacman -Q " + pkg)
		if err != nil {
			BluePrint("Gettings dependecy package: ", pkg)
			ExecuteCheck("sudo pacman --noconfirm -Sy " + pkg)
		}
	}
}

func BuildPackage(i PackageInfo, y PackYml) {
	CheckErr(os.Chdir(cfg.RepoCacheDir + "/" + i.ShortName))
	for _, script := range y.BuildScripts {
		BluePrint("Executing build script: ", script)
		ExecuteCheck(script)
	}
}

func GeneratePkgbuild(i PackageInfo, y PackYml) {
	CheckErr(os.Chdir(cfg.RepoCacheDir + "/" + i.ShortName))
	deps := fmt.Sprintf(depsTmpl, strings.Join(y.RunDeps, "\"\n  \""))
	if len(y.RunDeps) == 0 {
		deps = ""
	}
	makedeps := fmt.Sprintf(makedepsTmpl, strings.Join(y.BuildDeps, "\"\n  \""))
	if len(makedeps) == 0 {
		makedeps = ""
	}
	var installScripts []string
	for src, dst := range y.PackMap {
		installScripts = append(installScripts, FormatInstallSrc(src, dst))
	}
	install := strings.Join(installScripts, "\n  ")
	pkgb := fmt.Sprintf(pkgbuildTmpl, i.ShortName, i.Version, i.HttpsLink, deps, makedeps, install)
	CheckErr(core.WriteFile("PKGBUILD", pkgb))
}

func FormatInstallSrc(src string, dst string) string {
	i, err := os.Stat(src)
	CheckErr(err)
	if i.IsDir() {
		return fmt.Sprintf(`cd %s && find . -type f -exec install -Dm755 {} "${pkgdir}%s/{}" \; && cd $srcdir/..`, src, dst)
	}
	return fmt.Sprintf(`install -Dm755 %s "${pkgdir}%s"`, src, dst)
}

func InstallPackage() {
	BluePrint("Building and installing package: ", "makepkg")
	ExecuteCheck("makepkg --noconfirm -sfri")
}

func AddToMapping(i PackageInfo) {
	err := core.AppendToFile(cfg.MapFile, fmt.Sprintf("%s: %s", i.FullName, i.ShortName))
	CheckErr(err)
}

func CleanGitDir() {
	ExecuteCheck("git clean -fd")
	ExecuteCheck("git reset --hard")
}
