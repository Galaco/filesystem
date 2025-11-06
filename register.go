package filesystem

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/galaco/KeyValues"
)

// CreateFilesystemFromGameInfoDefinitions Reads game resource data paths
// from gameinfo.txt
// All games should ship with a gameinfo.txt, but it isn't actually mandatory.
// GameInfo definitions are quite unreliable, there are often bad entries;
// allowInvalidLocations will skip over bad paths, and an error collection
// will be returned will all paths that are invalid.
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

	searchPathsNode, _ := fsNode.Find("SearchPaths")
	searchPaths, _ := searchPathsNode.Children()
	basePath, _ = filepath.Abs(basePath)
	basePath = strings.Replace(basePath, "\\", "/", -1)

	// @NOTE: Kind of hacky.
	// extract the game directory from the base path (e.g. cstrike)
	// This is assumed to be where the gameinfo is located.
	// Will be used to avoid paths defined relative to the game dir.
	basePathLastFolderIndex := strings.LastIndex(basePath, "/")
	gameDir := ""
	gameRootPath := basePath
	if basePathLastFolderIndex != -1 {
		gameDir = basePath[basePathLastFolderIndex+1:]
		gameRootPath = strings.TrimSuffix(basePath, "/"+gameDir)
	}

	badPathErrorCollection := NewInvalidResourcePathCollectionError()

	for _, searchPath := range searchPaths {
		kv := searchPath
		path, _ := kv.AsString()
		path = strings.Trim(path, " ")

		// Current directory
		gameInfoPathRegex := regexp.MustCompile(`(?i)\|gameinfo_path\|`)
		if gameInfoPathRegex.MatchString(path) {
			path = gameInfoPathRegex.ReplaceAllString(path, basePath+"/")

			// Search for vpk directories in the top directory. Cannot confirm if this is actually accurate behaviour,
			// but CS:GO doesn't include any explicit vpk definitions in it's gameinfo.txt
			vpkDirectories, _ := filepath.Glob(basePath + "/*_dir.vpk")
			for _, key := range vpkDirectories {
				vpkHandle, err := openVPK(strings.TrimRight(key, "_dir.vpk"))
				if err != nil {
					if !allowInvalidLocations {
						return nil, err
					}
					badPathErrorCollection.AddPath(path)
					continue
				}
				fs.RegisterVpk(key, vpkHandle)
			}
		}

		// Executable directory
		allSourceEnginePathsRegex := regexp.MustCompile(`(?i)\|all_source_engine_paths\|`)
		if allSourceEnginePathsRegex.MatchString(path) {
			path = allSourceEnginePathsRegex.ReplaceAllString(path, gameRootPath+"/")
		}

		path = strings.TrimPrefix(path, gameDir+"/")

		if strings.Contains(strings.ToLower(kv.Key()), "mod") && !strings.HasPrefix(path, basePath) {
			path = basePath + "/" + path
		}

		path = strings.ReplaceAll(path, "//", "/")

		// Strip vpk extension, then load it
		path = strings.Trim(strings.Trim(path, " "), "\"")
		if strings.HasSuffix(path, ".vpk") {
			path = strings.Replace(path, ".vpk", "", 1)
			vpkHandle, err := openVPK(path)
			if err != nil {
				if !allowInvalidLocations {
					return nil, err
				}
				badPathErrorCollection.AddPath(path)
				continue
			}
			fs.RegisterVpk(path, vpkHandle)
		} else {
			// wildcard suffixes not useful
			if strings.HasSuffix(path, "/*") {
				path = strings.Replace(path, "/*", "", -1)
			}
			// @TODO: handle relative paths properly
			// For now, just assume non-absolute paths are relative to basePath
			// Absolute paths start with / or \, or with a drive letter (X:/ or X:\) on Windows
			if !((strings.HasPrefix(path, "/")) || (strings.HasPrefix(path, "\\")) || (len(path) > 3 && (path[1:3] == ":/" || path[1:3] == ":\\"))) {
				path = basePath + "/" + path
			}

			// @TODO: Hack to fix the trailing '/.' current dir
			path = strings.ReplaceAll(path, "/./", "/")
			path = strings.TrimSuffix(path, "/.")

			fs.RegisterLocalDirectory(path)
		}
	}

	// A filesystem can be valid, even if some GameInfo defined locations
	// were not.
	if allowInvalidLocations && len(badPathErrorCollection.paths) > 0 {
		return fs, badPathErrorCollection
	}

	return fs, nil
}
