package filesystem

import (
	"bytes"
	"io"
	"os"
	"strings"

	"github.com/galaco/vpk2"
)

type PakFile interface {
	GetFile(filePath string) ([]byte, error)
}

// FileSystem implements a Source Engine style filesystem. A file is read from the first place that has it: the
// pakfile, then each search path in the order it was registered, which is the order gameinfo.txt lists them.
type FileSystem struct {
	searchPaths []searchPath
	pakFile     PakFile
}

// searchPath is a vpk, or a directory of loose files
type searchPath struct {
	path string
	// vpk is nil for a directory
	vpk *vpk.VPK
}

// NewFileSystem returns a new filesystem
func NewFileSystem() *FileSystem {
	return &FileSystem{
		searchPaths: make([]searchPath, 0),
		pakFile:     nil,
	}
}

// PakFile returns loaded pakfile
// There can only be 1 registered pakfile at once.
func (fs *FileSystem) PakFile() PakFile {
	return fs.pakFile
}

// RegisterVpk adds a vpk package to the end of the search paths. Registering a path again replaces its vpk
// without changing its place.
func (fs *FileSystem) RegisterVpk(path string, vpkFile *vpk.VPK) {
	for idx := range fs.searchPaths {
		if fs.searchPaths[idx].vpk != nil && fs.searchPaths[idx].path == path {
			fs.searchPaths[idx].vpk = vpkFile
			return
		}
	}
	fs.searchPaths = append(fs.searchPaths, searchPath{path: path, vpk: vpkFile})
}

func (fs *FileSystem) UnregisterVpk(path string) {
	fs.unregister(path, true)
}

// RegisterLocalDirectory adds a directory to the end of the search paths. Registering a directory again keeps its
// place.
func (fs *FileSystem) RegisterLocalDirectory(directory string) {
	realpath := strings.TrimRight(directory, "/")
	for _, p := range fs.searchPaths {
		if p.vpk == nil && p.path == realpath {
			return
		}
	}
	fs.searchPaths = append(fs.searchPaths, searchPath{path: realpath})
}

func (fs *FileSystem) UnregisterLocalDirectory(directory string) {
	fs.unregister(strings.TrimRight(directory, "/"), false)
}

func (fs *FileSystem) unregister(path string, isVpk bool) {
	for idx, p := range fs.searchPaths {
		if p.path == path && (p.vpk != nil) == isVpk {
			fs.searchPaths = append(fs.searchPaths[:idx], fs.searchPaths[idx+1:]...)
			return
		}
	}
}

// RegisterPakFile Set a pakfile to be used as an asset directory.
// This would normally be called during each map load
func (fs *FileSystem) RegisterPakFile(pakFile PakFile) {
	fs.pakFile = pakFile
}

// UnregisterPakFile removes the current pakfile from
// available search locations
func (fs *FileSystem) UnregisterPakFile() {
	fs.pakFile = nil
}

// EnumerateResourcePaths returns all registered resource paths, in the order they are searched.
// PakFile is excluded.
func (fs *FileSystem) EnumerateResourcePaths() []string {
	list := make([]string, 0, len(fs.searchPaths))
	for _, p := range fs.searchPaths {
		list = append(list, p.path)
	}
	return list
}

// GetFile returns a file from the first place that has it: the pakfile, then each search path in order.
func (fs *FileSystem) GetFile(filename string) (io.Reader, error) {
	// sanitise file
	searchPath := NormalisePath(strings.ToLower(filename))

	// try to read from pakfile first
	if fs.pakFile != nil {
		f, err := fs.pakFile.GetFile(searchPath)
		if err == nil && f != nil && len(f) != 0 {
			return bytes.NewReader(f), nil
		}
	}

	for _, p := range fs.searchPaths {
		if p.vpk != nil {
			if entry := p.vpk.Entry(searchPath); entry != nil {
				return entry.Open()
			}
			continue
		}

		realpath := p.path + "/" + searchPath
		if _, err := os.Stat(realpath); err != nil {
			continue
		}
		file, err := os.ReadFile(realpath)
		if err != nil {
			return nil, err
		}
		return bytes.NewBuffer(file), nil
	}

	return nil, NewFileNotFoundError(filename)
}
