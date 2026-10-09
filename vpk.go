package filesystem

import (
	"os"
	"strings"

	"github.com/galaco/vpk2"
)

// openVPK opens a vpk named in a search path. "name.vpk" is a single file vpk if it exists, otherwise the
// multi-part vpk "name_dir.vpk" with its numbered archives. "name_dir.vpk" is always a multi-part vpk.
func openVPK(path string) (*vpk.VPK, error) {
	lower := strings.ToLower(path)
	if strings.HasSuffix(lower, "_dir.vpk") {
		return vpk.Open(vpk.MultiVPK(path[:len(path)-len("_dir.vpk")]))
	}
	if fileExists(path) {
		return vpk.Open(vpk.SingleVPK(path))
	}
	return vpk.Open(vpk.MultiVPK(path[:len(path)-len(".vpk")]))
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
