// 2023 FMNX team.
// Use of this code is governed by GNU General Public License.
// Additional information can be found on official web page: https://fmnx.su/
// Contact email: help@fmnx.su

package cmd

// This package contains all CLI commands that might be executed by user.
// Each file contains a single command, including root cmd.

import (
	"os"
	"strings"

	"fmnx.su/core/pack/config"
	"fmnx.su/core/pack/git"
	"fmnx.su/core/pack/pack"
	"fmnx.su/core/pack/pacman"
	"fmnx.su/core/pack/prnt"
	"fmnx.su/core/pack/system"
	"fmnx.su/core/pack/tmpl"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(buildCmd)
}

var buildCmd = &cobra.Command{
	Use:     "build",
	Aliases: []string{"b"},
	Short:   tmpl.BuildShort,
	Long:    tmpl.BuildLong,
	Run:     Build,
}

// Cli command preparing package in current directory.
func Build(cmd *cobra.Command, pkgs []string) {
	if len(pkgs) == 0 {
		dir := system.Pwd()
		BuildDirectory(dir, ``, false)
		return
	}
	for _, pkg := range pkgs {
		i := pack.GetPackInfo(pkg)
		err := git.Clone(i.GitUrl, i.Directory)
		CheckErr(err)
		BuildDirectory(i.Directory, i.Version, false)
	}
	prnt.Blue("Build complete, results in: ", config.PackageCacheDir)
}

// Build package in specified directory. Assumes this directory has cloned git
// repository with PKGBUILD in it.
func BuildDirectory(dir string, version string, install bool) string {
	pkgname := ValidateBuildDir(dir)
	prnt.Yellow("Building: ", pkgname)
	if version == `` {
		branch, err := git.DefaultBranch(dir)
		CheckErr(err)
		version, err = git.LastCommitDir(dir, branch)
		CheckErr(err)
	}
	err := git.Checkout(dir, version)
	CheckErr(err)
	ResolvePackDeps(dir)
	err = pacman.Build(dir)
	CheckErr(err)
	if install {
		err = pacman.InstallDir(dir)
		CheckErr(err)
	}
	if !config.RemoveBuiltPackages {
		err = system.MvExt(dir, config.PackageCacheDir, ".pkg.tar.zst")
		CheckErr(err)
	}
	if !config.RemoveGitRepos {
		err = git.Clean(dir)
		CheckErr(err)
		prnt.Green("Complete: ", pkgname)
		return version
	}
	err = os.RemoveAll(dir)
	CheckErr(err)
	return version
}

// Validate directory to be valid pack package - git repo name matching package
// name defined in PKGBUILD.
func ValidateBuildDir(dir string) string {
	url, err := git.Url(dir)
	CheckErr(err)
	pkgname, err := pacman.PkgbuildParam(dir, "pkgname")
	CheckErr(err)
	splt := strings.Split(url, "/")
	if pkgname != splt[len(splt)-1] {
		prnt.Red("package name is not matching git link, can't build: ", dir)
		os.Exit(1)
	}
	return pkgname
}

// Resolve pack dependencies for package in provided directory.
func ResolvePackDeps(dir string) {
	deps, err := pacman.GetDeps(dir + "/PKGBUILD")
	CheckErr(err)
	groups := GroupPackages(deps)
	uninstalled := pack.GetUninstalled(groups.PackPackages)
	if len(uninstalled) > 0 {
		prnt.Blue("Resolving pack deps: ", strings.Join(uninstalled, " "))
		Install(nil, uninstalled)
	}
	err = pack.SwapDeps(dir+"/PKGBUILD", groups.PackPackages)
	CheckErr(err)
}
