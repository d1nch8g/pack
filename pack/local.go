// 2023 FMNX team.
// Use of this code is governed by GNU General Public License.
// Additional information can be found on official web page: https://fmnx.io/
// Contact email: help@fmnx.io

package pack

import (
	"fmt"
	"os"
	"strings"

	"fmnx.io/core/pack/pacman"
	"fmnx.io/core/pack/tmpl"
)

type PackageGroups struct {
	PacmanPackages []string
	PackPackages   []string
}

// Split packages into pacman and pack groups.
func Split(pkgs []string) PackageGroups {
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

// Temporarily swap dependencies in PKGBUILD file from pack to pacman format
// while pacakge is installing.
func SwapDeps(pkgbuild string, deps []string) error {
	b, err := os.ReadFile(pkgbuild)
	if err != nil {
		return err
	}
	var rez = string(b)
	for _, dep := range deps {
		dashsplt := strings.Split(dep, "/")
		shortname := dashsplt[len(dashsplt)-1]
		rez = strings.ReplaceAll(rez, dep, shortname)
	}
	return os.WriteFile(pkgbuild, []byte(rez), 0o600)
}

// Get pack dependecies from PKGBUILD file.
func GetDeps(pkgbuild string) ([]string, error) {
	deps, err := pacman.PkgbuildParams(pkgbuild, "depends")
	if err != nil {
		return nil, err
	}
	makedeps, err := pacman.PkgbuildParams(pkgbuild, "makedepends")
	if err != nil {
		return nil, err
	}
	alldeps := append(deps, makedeps...)
	groups := Split(alldeps)
	return groups.PackPackages, nil
}

// Print package description.
func PrintDescription(d PkgInfo) {
	fmt.Printf(
		tmpl.PrettyDesc,
		d.Name,
		d.Version,
		d.Description,
		d.Size,
		d.Url,
		d.PackName,
		d.PackVersion,
		d.PackBranch,
	)
}
