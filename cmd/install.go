// 2023 FMNX team.
// Use of this code is governed by GNU General Public License.
// Additional information can be found on official web page: https://fmnx.io/
// Contact email: help@fmnx.io

package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"fmnx.io/core/pack/config"
	"fmnx.io/core/pack/database"
	"fmnx.io/core/pack/print"
	"fmnx.io/core/pack/system"
	"fmnx.io/core/pack/tmpl"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

func init() {
	rootCmd.AddCommand(installCmd)
}

var installCmd = &cobra.Command{
	Use:     "install",
	Example: "pack install fmnx.io/core/ainst fmnx.io/core/keks@main",
	Aliases: []string{"i"},
	Short:   tmpl.InstallShort,
	Long:    tmpl.InstallLong,
	Run:     Install,
}

// Cli command installing packages into system.
func Install(cmd *cobra.Command, upkgs []string) {
	PrepareForInstallation(upkgs)
	pkgs := SplitPackages(upkgs)
	CheckUnreachablePacmanPackages(pkgs.PacmanPackages)
	CheckUnreachablePackPackages(pkgs.PackPackages)
	InstallPacmanPackages(pkgs.PacmanPackages)
	InstallPackPackages(pkgs.PackPackages)
}

// Exis if there is no target packages, prepare cache directories.
func PrepareForInstallation(pkgs []string) {
	if len(pkgs) == 0 {
		return
	}
	if !Updating {
		print.Blue("Installing packages: ", strings.Join(pkgs, " "))
	}
}

type PackageGroups struct {
	PacmanPackages []string
	PackPackages   []string
}

// Split packages into pacman and pack to resolve dependencies differently.
func SplitPackages(pkgs []string) PackageGroups {
	var pacmanPackages []string
	var packPackages []string
	for _, pkg := range pkgs {
		if strings.Contains(pkg, "/") {
			packPackages = append(packPackages, pkg)
			continue
		}
		pacmanPackages = append(pacmanPackages, pkg)
	}
	return PackageGroups{
		PacmanPackages: pacmanPackages,
		PackPackages:   packPackages,
	}
}

// Check if some pacman packages could not be installed.
func CheckUnreachablePacmanPackages(pkgs []string) {
	deps := GetDependeciesResolvableByPacman()
	var unreachable []string
	for _, pkg := range pkgs {
		if _, ok := deps[pkg]; !ok {
			unreachable = append(unreachable, pkg)
		}
	}
	if len(unreachable) != 0 {
		pkgs := strings.Join(unreachable, " ")
		print.Red("Unable to resolve those pacman packages: ", pkgs)
		os.Exit(1)
	}
}

// Fill struct that shows which packages could be resolved with pacman
// (packages that you can load from pacman servers).
func GetDependeciesResolvableByPacman() map[string]struct{} {
	o, err := system.Call("pacman -Ssq")
	CheckErr(err)
	deps := map[string]struct{}{}
	for _, pkg := range strings.Split(o, "\n") {
		deps[pkg] = struct{}{}
	}
	return deps
}

// Check if some pack packages could not be installed.
func CheckUnreachablePackPackages(pkgs []string) {
	g, _ := errgroup.WithContext(context.Background())
	var unreachable []string
	for _, pkg := range pkgs {
		syncpkg := pkg
		g.Go(func() error {
			err := CheckPackPackage(syncpkg)
			if err != nil {
				unreachable = append(unreachable, syncpkg)
			}
			return err
		})
	}
	err := g.Wait()
	if err != nil {
		out := strings.Join(unreachable, " ")
		print.Red("Some pack packages are unreachable: ", out)
		os.Exit(1)
	}
}

// Validate pack package to be reachable via network.
func CheckPackPackage(pkg string) error {
	i := EjectInfoFromPackLink(pkg)
	out, err := system.Callf("git clone %s %s", i.HttpsLink, i.Directory)
	if err != nil {
		if !strings.Contains(out, "already exists and is not an empty dir") {
			print.Red("Unable to reach git for: ", pkg)
			return err
		}
	}
	_, err = os.Stat(i.Pkgbuild)
	if err != nil {
		print.Red("Unable to find PKGBUILD for: ", pkg)
	}
	return err
}

// Info formed from pack link about all information related to that package.
type PackInfo struct {
	PacmanName string
	PackName   string
	Directory  string
	Version    string
	Pkgbuild   string
	HttpsLink  string
}

// Eject pack information for provided pack link.
func EjectInfoFromPackLink(pkg string) PackInfo {
	rez := PackInfo{}
	versplt := strings.Split(pkg, "@")
	rez.PackName = versplt[0]
	rez.HttpsLink = "https://" + versplt[0]
	if len(versplt) > 1 {
		rez.Version = versplt[1]
	}
	dashsplt := strings.Split(rez.PackName, "/")
	rez.PacmanName = dashsplt[len(dashsplt)-1]
	rez.Directory = config.RepoCacheDir + "/" + rez.PacmanName
	rez.Pkgbuild = rez.Directory + "/PKGBUILD"
	return rez
}

// Install pacman packages.
func InstallPacmanPackages(pkgs []string) {
	uninstalled := CleanAlreadyInstalled(pkgs)
	if len(uninstalled) == 0 {
		return
	}
	joined := strings.Join(uninstalled, " ")
	o, err := system.Callf("sudo pacman --noconfirm -S %s", joined)
	if err != nil {
		print.Red("Unable to get pacman packages: ", joined)
		fmt.Println(o)
		os.Exit(1)
	}
	print.Green("Pacman packages installed: ", joined)
}

// Removes pacman packages that are already installed in the system.
func CleanAlreadyInstalled(pkgs []string) []string {
	var uninstalledPkgs []string
	for _, pkg := range pkgs {
		_, err := system.Callf("pacman -Q %s", pkg)
		if err != nil {
			uninstalledPkgs = append(uninstalledPkgs, pkg)
		}
	}
	return uninstalledPkgs
}

// Checks if packages are not installed and installing them.
func InstallPackPackages(pkgs []string) {
	for _, pkg := range pkgs {
		_, err := database.Get(pkg, database.PACK)
		if err == nil {
			continue
		}
		InstallPackPackage(EjectInfoFromPackLink(pkg))
	}
	if len(pkgs) > 0 {
		pkglist := strings.Join(pkgs, " ")
		if !Updating {
			print.Green("Installed: ", pkglist)
		}
	}
}

// Install pack package.
func InstallPackPackage(i PackInfo) {
	CleanRepository(i)
	branch, version := SetPackageVersion(i)
	packDeps := EjectPackDependencies(i.Pkgbuild)
	Install(nil, packDeps)
	SwapPackDependencies(i.Pkgbuild, packDeps)
	InstallPackageWithMakepkg(i)
	database.Add(database.Package{
		PacmanName: i.PacmanName,
		PackName:   i.PackName,
		Version:    version,
		Branch:     branch,
	})
	CachePackage(i.Directory)
	CleanRepository(i)
}

// Checkout repository with pack package to some version. And return applied
// branch and version for this repo.
func SetPackageVersion(i PackInfo) (string, string) {
	branch := GetDefaultGitBranch(i.Directory)
	GetDirCheckout(i.Directory, branch)
	GitDirPull(i.Directory)
	if i.Version == `` {
		i.Version = GetLastCommitHash(i.Directory, branch)
	}
	GetDirCheckout(i.Directory, i.Version)
	return branch, i.Version
}

// Pull changes for specified directory with git repository.
func GitDirPull(dir string) {
	o, err := system.Callf("git -C %s pull", dir)
	if err != nil {
		print.Red("Unable to git pull: ", dir)
		fmt.Println(o)
		os.Exit(1)
	}
}

// Perform git checkout for specific branch/commit/tag on dir.
func GetDirCheckout(dir string, branch string) {
	o, err := system.Callf("git -C %s checkout %s ", dir, branch)
	if err != nil {
		if !strings.HasPrefix(o, "Already on ") {
			print.Red("Unable to set pack version for: ", dir)
			fmt.Println(o)
			os.Exit(1)
		}
	}
}

// Returns default branch for git repository located in git directory.
func GetDefaultGitBranch(dir string) string {
	origin, err := system.Callf("git -C %s remote show", dir)
	CheckErr(err)
	origin = strings.Trim(origin, "\n")
	remoteInfo, err := system.Callf("git -C %s remote show %s", dir, origin)
	CheckErr(err)
	rawInfo := strings.Split(remoteInfo, "HEAD branch: ")[1]
	return strings.Split(rawInfo, "\n")[0]
}

// Get last commit hash for provided git branch in git directory.
func GetLastCommitHash(dir string, branch string) string {
	command := `git -C ` + dir + ` log -n 1 --pretty=format:"%H" ` + branch
	o, err := system.Call(command)
	if err != nil {
		print.Red("Unable to get last git commit sha: ", dir)
		fmt.Println(o)
		os.Exit(1)
	}
	return strings.Trim(o, "\n")
}

// Get dependencies and make dependencies related to pack from PKGBUILD file.
func EjectPackDependencies(pkgbuild string) []string {
	deps, err := system.GetShellList(pkgbuild, "depends")
	CheckErr(err)
	makedeps, err := system.GetShellList(pkgbuild, "makedepends")
	CheckErr(err)
	alldeps := append(deps, makedeps...)
	groups := SplitPackages(alldeps)
	return groups.PackPackages
}

// Temporarily swap pack dependencies in PKGBUILD to pacman package name for
// installation process.
func SwapPackDependencies(pkgbuild string, deps []string) {
	b, err := os.ReadFile(pkgbuild)
	CheckErr(err)
	var rez = string(b)
	for _, dep := range deps {
		dashsplt := strings.Split(dep, "/")
		shortname := dashsplt[len(dashsplt)-1]
		rez = strings.ReplaceAll(rez, dep, shortname)
	}
	err = os.WriteFile(pkgbuild, []byte(rez), 0o600)
	CheckErr(err)
}

// Install package with makepkg.
func InstallPackageWithMakepkg(i PackInfo) {
	CheckErr(os.Chdir(i.Directory))
	print.Yellow("Building package: ", i.PackName)
	out, err := system.Call("makepkg -sfi --noconfirm")
	if err != nil {
		print.Red("Unable to build and install package: ", i.PackName)
		fmt.Println(out)
		os.Exit(1)
	}
}

// Move prepared .pkg.tar.zst package into pacman cache.
func CachePackage(dir string) {
	if !config.RemoveBuiltPackages {
		const command = "sudo mv %s/*.pkg.tar.zst %s"
		_, err := system.Callf(command, dir, config.PackageCacheDir)
		CheckErr(err)
	}
}

// Clean or remove git directory after installation depending on configuration.
func CleanRepository(i PackInfo) {
	if config.RemoveGitRepos {
		CheckErr(os.RemoveAll(i.Directory))
		return
	}
	_, err := system.Callf("sudo rm -rf %s/*.tar.gz", i.Directory)
	CheckErr(err)
	CheckErr(os.RemoveAll(i.Directory + "/pkg"))
	CheckErr(os.RemoveAll(i.Directory + "/src"))
	_, err = system.Callf("git -C %s clean -fd", i.Directory)
	CheckErr(err)
	_, err = system.Callf("git -C %s reset --hard", i.Directory)
	CheckErr(err)
}
