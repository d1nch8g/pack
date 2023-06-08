// 2023 FMNX team.
// Use of this code is governed by GNU General Public License.
// Official web page: https://fmnx.su/core/pack
// Contact email: help@fmnx.su

package server

import (
	"os"
	"path"

	"fmnx.su/core/pack/pacman"
)

// This implementation can be used to create local directory for packages and
// add packages to database with a function.
type LocalDirDb struct {
	Dir    string
	DbName string
}

// Parameters required to add package to pacman database.
type AddPkgParameters struct {
	Package  []byte
	Sign     []byte
	Filename string
	Force    bool
}

func (d *LocalDirDb) AddPkg(p AddPkgParameters) error {
	pkgpath := path.Join(d.Dir, p.Filename)
	err := os.WriteFile(pkgpath, p.Package, 0600)
	if err != nil {
		return err
	}

	err = os.WriteFile(path.Join(d.Dir, p.Filename+".sig"), p.Sign, 0600)
	if err != nil {
		return err
	}

	return pacman.RepoAdd(path.Join(d.Dir, d.DbName+".db.tar.gz"), pkgpath)
}