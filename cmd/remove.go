// Copyright 2023 FMNX team.
// Use of this code is governed by GNU General Public License.
// Additional information can be found on official web page: https://fmnx.io/
// Contact email: help@fmnx.io

package cmd

import (
	"os"
	"strings"

	"fmnx.io/core/pack/database"
	"fmnx.io/core/pack/print"
	"fmnx.io/core/pack/system"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(removeCmd)
}

var removeCmd = &cobra.Command{
	Use:     "remove",
	Aliases: []string{"rm", "r"},
	Short:   "❌ remove packages",
	Long: `❌ remove packages

Use this command to remove packages from system. You can specify both pacman 
packages and pack links.

Example:
pack rm fmnx.io/core/ainst`,
	Run: Remove,
}

// Cli command removing packages from system.
func Remove(cmd *cobra.Command, pkgs []string) {
	groups := SplitPackages(pkgs)
	restPacmanPkgs := GetPacmanPackagesFromPackNames(groups.PackPackages)
	groups.PacmanPackages = append(groups.PacmanPackages, restPacmanPkgs...)
	RemovePacmanPackages(groups.PacmanPackages)
	for _, pkg := range groups.PacmanPackages {
		database.Remove(pkg, database.PACMAN)
	}
}

// Try to remove all packages at once.
func RemovePacmanPackages(pkgs []string) {
	pkgsStr := strings.Join(pkgs, " ")
	o, err := system.Callf("sudo pacman --noconfirm -R %s", pkgsStr)
	if err != nil {
		PrintNotFoundPackages(o)
		os.Exit(1)
	}
	print.Yellow("Packages removed: ", pkgsStr)
}

// Get pacman packages from parsed removal command.
func PrintNotFoundPackages(o string) {
	o = strings.ReplaceAll(o, "\n", " ")
	o = strings.ReplaceAll(o, `error: target not found: `, "")
	print.Red("Packages not found: ", o)
}

// Get pacman packages related to pack names.
func GetPacmanPackagesFromPackNames(pkgs []string) []string {
	var out []string
	for _, pkg := range pkgs {
		pkgInfo, err := database.Get(pkg, database.PACK)
		if err != nil {
			print.Red("Unable to find package: ", pkg)
			os.Exit(1)
		}
		out = append(out, pkgInfo.PacmanName)
	}
	return out
}
