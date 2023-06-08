// 2023 FMNX team.
// Use of this code is governed by GNU General Public License.
// Official web page: https://fmnx.su/core/pack
// Contact email: help@fmnx.su

package pack

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"fmnx.su/core/pack/tmpl"
	"github.com/mitchellh/ioprogress"
)

// Parameters that will be used to execute push command.
type PushParameters struct {
	// Directory to read package files and signatures.
	Directory string
	// Which protocol to use for connection.
	Protocol string
	// Custom endpoint for package push
	Endpoint string
	// Owerwrite package with same version if exists.
	Force bool
	// Where command will write output text.
	Stdout io.Writer
	// Where command will write output text.
	Stderr io.Writer
	// Stdin from user is command will ask for something.
	Stdin io.Reader
}

func pushdefault() *PushParameters {
	return &PushParameters{
		Protocol:  "https",
		Endpoint:  "/api/pack/push",
		Directory: "/var/cache/pacman/pkg",
	}
}

// Push your package to registry.
func Push(args []string, prms ...PushParameters) error {
	p := formOptions(prms, pushdefault)

	email, err := gnupgEmail()
	if err != nil {
		return err
	}

	pkgs, _, err := formatpkgs(args)
	if err != nil {
		return err
	}

	err = checkRegistries(pkgs)
	if err != nil {
		return err
	}

	filenames, err := listPkgFilenames(p.Directory)
	if err != nil {
		return err
	}

	pprms, err := fillfileinfo(fillparams{
		filenames: filenames,
		packages:  pkgs,
		directory: p.Directory,
	})
	if err != nil {
		return err
	}

	for _, pp := range pprms {
		pp.Protocol = p.Protocol
		err = push(pp, email)
		if err != nil {
			return err
		}
	}

	return nil
}

// This function will be used to get email from user's GnuPG identitry.
func gnupgEmail() (string, error) {
	gnupgident, err := gnuPGIdentity()
	if err != nil {
		return ``, err
	}
	return strings.ReplaceAll(strings.Split(gnupgident, "<")[1], ">", ""), nil
}

// Check if all packages have registries where they will be pushed to.
func checkRegistries(pkgs []registrypkg) error {
	for _, pkg := range pkgs {
		if pkg.Registry == "" {
			return errors.New("provide registry to push package: " + pkg.Name)
		}
	}
	return nil
}

// Structure including base registry parameters and information about file
// pathes requied to push packages.
type pushpkg struct {
	registrypkg
	PushParameters
	// Name of the file which will be pushed.
	Filename string
	// Path to file which will be read and pushed.
	PkgPath string
	// Signature encoded to base64 string to check.
	Signature string
}

// List file names in provided cache directory.
func listPkgFilenames(dir string) ([]string, error) {
	des, err := os.ReadDir(dir)
	if err != nil {
		return nil, errors.New(
			tmpl.Err + " unable to get directory contents: " +
				dir + " " + err.Error(),
		)
	}
	var fns []string
	for _, de := range des {
		fn := de.Name()
		if strings.HasSuffix(fn, ".pkg.tar.zst") {
			fns = append(fns, fn)
		}
	}
	return fns, nil
}

type fillparams struct {
	filenames []string
	packages  []registrypkg
	directory string
}

// Create array of package arguements, that will be pushed to registry.
func fillfileinfo(p fillparams) ([]pushpkg, error) {
	var ppkgs []pushpkg
	for _, pkg := range p.packages {
		for i := len(p.filenames) - 1; i >= 0; i-- {
			filename := p.filenames[i]
			if !strings.Contains(filename, pkg.Name) {
				continue
			}
			pkgname, err := ejectpkgname(filename)
			if err != nil {
				return nil, err
			}
			if pkgname == pkg.Name {
				pkgpath := path.Join(p.directory, filename)
				sigbase64, err := readpkgsign(pkgpath + ".sig")
				if err != nil {
					return nil, err
				}
				ppkgs = append(ppkgs, pushpkg{
					registrypkg: pkg,
					Filename:    filename,
					PkgPath:     pkgpath,
					Signature:   sigbase64,
				})
				break
			}
			if i == 0 {
				return nil, errors.New("unable to find package: " + pkg.Name)
			}
		}
	}
	return ppkgs, nil
}

// Eject package name from file name.
func ejectpkgname(filename string) (string, error) {
	pkgsplt := strings.Split(filename, "-")
	if len(pkgsplt) < 4 {
		return ``, errors.New("not valid package file name: " + filename)
	}
	return strings.Join(pkgsplt[:len(pkgsplt)-3], "-"), nil
}

// Read package signature and encode to base64.
func readpkgsign(path string) (string, error) {
	err := exec.Command("bash", "-c", "sudo chmod 0777 "+path).Run()
	if err != nil {
		return ``, errors.New("unable to read signature: " + path)
	}
	sigbytes, err := os.ReadFile(path)
	if err != nil {
		return ``, errors.New("unable to read signature: " + path)
	}
	return base64.StdEncoding.EncodeToString(sigbytes), nil
}

// This function pushes package to registry via http.
func push(p pushpkg, email string) error {
	packagefile, err := os.Open(p.PkgPath)
	if err != nil {
		return err
	}
	fi, err := os.Stat(p.PkgPath)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		http.MethodPut,
		p.Protocol+"://"+p.Registry+p.Endpoint,
		&ioprogress.Reader{
			Reader:       packagefile,
			Size:         fi.Size(),
			DrawFunc:     tmpl.Loader(p.Registry, p.Owner, p.Name),
			DrawInterval: time.Nanosecond * 1000,
		},
	)
	if err != nil {
		return err
	}

	req.Header.Add("file", p.Filename)
	req.Header.Add("email", email)
	req.Header.Add("sign", p.Signature)
	if p.Owner != "" {
		req.Header.Add("owner", p.Owner)
	}
	if p.Force {
		req.Header.Add("force", "true")
	}

	var client http.Client
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return errors.Join(err, errors.New(resp.Status))
		}
		return fmt.Errorf("%s, %s - %s", resp.Status, string(b), p.Filename)
	}
	return nil
}
