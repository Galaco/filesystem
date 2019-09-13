package filesystem

import (
	"github.com/galaco/vpk2"
)

// openVPK Basic wrapper around vpk library.
// Just opens a multi-part vpk (ver 2 only)
func openVPK(filepath string) (*vpk.VPK, error) {
	return vpk.Open(vpk.MultiVPK(filepath))
}
