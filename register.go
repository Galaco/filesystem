package filesystem

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/galaco/KeyValues"
)

var (
	gameInfoPathToken         = regexp.MustCompile(`(?i)\|gameinfo_path\|`)
	allSourceEnginePathsToken = regexp.MustCompile(`(?i)\|all_source_engine_paths\|`)
	vpkArchiveName            = regexp.MustCompile(`(?i)_[0-9]{3}\.vpk$`)
)

// CreateFilesystemFromGameInfoDefinitions Reads game resource data paths
// from gameinfo.txt
// All games should ship with a gameinfo.txt, but it isn't actually mandatory.
// GameInfo definitions are quite unreliable, there are often bad entries;
// allowInvalidLocations will skip over bad paths, and an error collection
// will be returned will all paths that are invalid.
//
// Search paths are searched in the order gameinfo.txt lists them, as the engine does since SteamPipe: vpks and
// directories are interleaved as listed. Only paths with the "game" path ID, which game content is loaded from, are
// added. A path ending in "/*" adds each subdirectory and vpk of that directory, in alphabetical order.
func CreateFilesystemFromGameInfoDefinitions(basePath string, gameInfo *keyvalues.KeyValue, allowInvalidLocations bool) (*FileSystem, error) {
	fs := NewFileSystem()
	var gameInfoNode *keyvalues.KeyValue
	if gameInfo.Key() != "GameInfo" {
		gameInfoNode, _ = gameInfo.Find("GameInfo")
	} else {
		gameInfoNode = gameInfo
	}
	if gameInfoNode == nil {
		return nil, ErrorInvalidGameInfo
	}
	fsNode, _ := gameInfoNode.Find("FileSystem")
	if fsNode == nil {
		return nil, ErrorInvalidGameInfo
	}
	searchPathsNode, _ := fsNode.Find("SearchPaths")
	if searchPathsNode == nil {
		return nil, ErrorInvalidGameInfo
	}
	searchPaths, _ := searchPathsNode.Children()

	// basePath is the directory holding gameinfo.txt (e.g. cstrike). Relative search paths are relative to the
	// directory above it, which holds every game directory (e.g. cstrike and hl2).
	basePath, _ = filepath.Abs(basePath)
	basePath = NormalisePath(basePath)
	gameRootPath := filepath.ToSlash(filepath.Dir(basePath))

	badPaths := NewInvalidResourcePathCollectionError()
	addVpk := func(path string) error {
		vpkFile, err := openVPK(path)
		if err != nil {
			if !allowInvalidLocations {
				return err
			}
			badPaths.AddPath(path)
			return nil
		}
		fs.RegisterVpk(path, vpkFile)
		return nil
	}
	addDirectory := func(path string) error {
		// The engine mounts a directory's pak01 vpk ahead of its loose files
		if pak := path + "/pak01_dir.vpk"; fileExists(pak) {
			if err := addVpk(pak); err != nil {
				return err
			}
		}
		fs.RegisterLocalDirectory(path)
		return nil
	}

	for _, kv := range searchPaths {
		if !hasPathID(kv.Key(), "game") {
			continue
		}
		path, _ := kv.AsString()
		path = strings.Trim(strings.TrimSpace(path), "\"")
		path = gameInfoPathToken.ReplaceAllString(path, basePath+"/")
		path = allSourceEnginePathsToken.ReplaceAllString(path, gameRootPath+"/")
		path = NormalisePath(path)
		if !isAbsolute(path) {
			path = gameRootPath + "/" + path
		}

		var err error
		switch {
		case strings.HasSuffix(path, "/*"):
			err = addWildcard(filepath.ToSlash(filepath.Clean(strings.TrimSuffix(path, "/*"))), addVpk, addDirectory)
		case strings.HasSuffix(strings.ToLower(path), ".vpk"):
			err = addVpk(filepath.ToSlash(filepath.Clean(path)))
		default:
			err = addDirectory(filepath.ToSlash(filepath.Clean(path)))
		}
		if err != nil {
			return nil, err
		}
	}

	// A filesystem can be valid, even if some GameInfo defined locations
	// were not.
	if len(badPaths.paths) > 0 {
		return fs, badPaths
	}

	return fs, nil
}

// addWildcard adds each subdirectory and vpk of a directory, in alphabetical order. The numbered archives of a
// multi-part vpk are read through its _dir.vpk.
func addWildcard(directory string, addVpk, addDirectory func(string) error) error {
	entries, err := os.ReadDir(directory)
	if err != nil {
		// A missing directory, such as an empty custom folder, has nothing to add
		return nil
	}
	sort.Slice(entries, func(i, j int) bool {
		return strings.ToLower(entries[i].Name()) < strings.ToLower(entries[j].Name())
	})
	for _, entry := range entries {
		path := directory + "/" + entry.Name()
		switch {
		case entry.IsDir():
			err = addDirectory(path)
		case strings.HasSuffix(strings.ToLower(entry.Name()), ".vpk") && !vpkArchiveName.MatchString(entry.Name()):
			err = addVpk(path)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// hasPathID returns whether a search path's key, such as "game+mod", includes a path ID
func hasPathID(key string, pathID string) bool {
	for _, id := range strings.Split(key, "+") {
		if strings.EqualFold(strings.TrimSpace(id), pathID) {
			return true
		}
	}
	return false
}

// isAbsolute returns whether a path starts at a root: "/", "\", or a drive letter such as "C:/"
func isAbsolute(path string) bool {
	return strings.HasPrefix(path, "/") || strings.HasPrefix(path, "\\") ||
		(len(path) > 2 && path[1] == ':' && (path[2] == '/' || path[2] == '\\'))
}
