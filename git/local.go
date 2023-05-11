// 2023 FMNX team.
// Use of this code is governed by GNU General Public License.
// Additional information can be found on official web page: https://fmnx.io/
// Contact email: help@fmnx.io

package git

import (
	"errors"
	"strings"

	"fmnx.io/core/pack/system"
)

// Switch repo to branch/tag/commit.
func Checkout(dir string, target string) error {
	o, err := system.Callf("git -C %s checkout %s ", dir, target)
	if err != nil {
		if !strings.HasPrefix(o, "Already on ") {
			return errors.New("git unable to find checkout target:\n" + o)
		}
	}
	return nil
}

// Clean git repository - all changes in tracked files, newly created files and
// files under gitignore.
func Clean(dir string) error {
	o, err := system.Callf("git -C %s clean -xdf", dir)
	if err != nil {
		return errors.New("git unable to clean xdf:\n" + o)
	}
	o, err = system.Callf("git -C %s reset --hard", dir)
	if err != nil {
		return errors.New("git unable to reset -hard:\n" + o)
	}
	return nil
}

// Get default repo url for git repo.
func RepoUrl(dir string) (string, error) {
	return ``, nil
}

// Get last commit hash for git repo in a branch.
func LastCommitDir(dir string, branch string) (string, error) {
	command := `git -C ` + dir + ` log -n 1 --pretty=format:"%H" ` + branch
	o, err := system.Call(command)
	if err != nil {
		return ``, errors.New("git unable to log:\n" + o)
	}
	return strings.Trim(o, "\n"), nil
}