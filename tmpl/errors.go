// 2023 FMNX team.
// Use of this code is governed by GNU General Public License.
// Official web page: https://fmnx.su/core/pack
// Contact email: help@fmnx.su

package tmpl

import "github.com/fatih/color"

var GnuPGprivkeyNotFound = color.RedString("error:") + ` GnuPG private key not found.
It is required for package signing, run:

1) Install gnupg:
` + color.YellowString("pack -S gnupg") + `

2) Generate a key:
` + color.YellowString("gpg --gen-key") + `

3) Get KEY-ID, paste it to next command:
` + color.YellowString("gpg -k") + `

4) Send it to key server:
` + color.YellowString("gpg --send-keys KEY-ID") + `

5) Edit ` + color.BlueString("PACKAGER") + ` variable in ` + color.CyanString("/etc/makepkg.conf") + `
Name and email should match with name and email in GnuPG authority for pack to work properly.
`

var NoPackager = color.RedString("error:") + ` packager not found.

Add PACKAGER variable matching your GnuPG authority in ` + color.CyanString("/etc/makepkg.conf") + `

PACKAGER="John Doe <john@doe.com>"
`

var SignerMissmatch = color.RedString("error:") + ` signer and packager are different.

Authority you defined in GnuPG is not matching with PACKAGER variable in 

` + color.BlueString("PACKAGER") + `=` + color.HiMagentaString("John Doe <john@doe.com>\n")