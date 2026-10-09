package filesystem

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	keyvalues "github.com/galaco/KeyValues"
)

// writeVPK creates a version 2 vpk holding files, with each file's data stored in the directory tree. A path ending
// in _dir.vpk also gets an empty numbered archive beside it, as a multi-part vpk has.
func writeVPK(t *testing.T, filePath string, files map[string]string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		t.Fatal(err)
	}

	// The tree groups files by extension, then directory
	tree := map[string]map[string]map[string]string{}
	for rel, data := range files {
		dir, file := path.Split(rel)
		ext := path.Ext(file)
		base := strings.TrimSuffix(file, ext)
		dir, ext = strings.TrimSuffix(dir, "/"), strings.TrimPrefix(ext, ".")
		if tree[ext] == nil {
			tree[ext] = map[string]map[string]string{}
		}
		if tree[ext][dir] == nil {
			tree[ext][dir] = map[string]string{}
		}
		tree[ext][dir][base] = data
	}
	var body bytes.Buffer
	for ext, dirs := range tree {
		body.WriteString(ext + "\x00")
		for dir, bases := range dirs {
			body.WriteString(dir + "\x00")
			for base, data := range bases {
				body.WriteString(base + "\x00")
				binary.Write(&body, binary.LittleEndian, struct {
					CRC          uint32
					PreloadBytes uint16
					ArchiveIndex int16
					Offset       uint32
					Length       uint32
					Terminator   uint16
				}{crc32.ChecksumIEEE([]byte(data)), uint16(len(data)), 0x7fff, 0, 0, 0xffff})
				body.WriteString(data)
			}
			body.WriteString("\x00")
		}
		body.WriteString("\x00")
	}
	body.WriteString("\x00")

	var out bytes.Buffer
	binary.Write(&out, binary.LittleEndian, [7]uint32{0x55aa1234, 2, uint32(body.Len()), 0, 0, 0, 0})
	out.Write(body.Bytes())
	writeFile(t, filePath, out.String())
	if strings.HasSuffix(filePath, "_dir.vpk") {
		writeFile(t, strings.TrimSuffix(filePath, "_dir.vpk")+"_000.vpk", "")
	}
}

func writeFile(t *testing.T, path string, data string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}
}

type testPakFile map[string]string

func (p testPakFile) GetFile(path string) ([]byte, error) {
	if data, ok := p[path]; ok {
		return []byte(data), nil
	}
	return nil, errors.New("not found")
}

func TestCreateFilesystemFromGameInfoDefinitions(t *testing.T) {
	root := t.TempDir()
	const dir = "materials/order/"

	// Every file below is in two or more places; which one is read shows which place is searched first
	writeVPK(t, root+"/game/custom/a_addon.vpk", map[string]string{dir + "d.txt": "custom a vpk"})
	writeFile(t, root+"/game/custom/b_addon/"+dir+"b.txt", "custom b dir")
	writeFile(t, root+"/game/custom/b_addon/"+dir+"d.txt", "custom b dir")
	writeVPK(t, root+"/game/custom/c_addon_dir.vpk", map[string]string{dir + "c.txt": "custom c vpk"})
	writeVPK(t, root+"/game/game_pak_dir.vpk", map[string]string{
		dir + "a.txt": "game vpk", dir + "b.txt": "game vpk", dir + "c.txt": "game vpk",
	})
	writeVPK(t, root+"/shared/shared_misc_dir.vpk", map[string]string{dir + "e.txt": "shared vpk"})
	writeFile(t, root+"/game/"+dir+"a.txt", "game dir")
	writeFile(t, root+"/game/"+dir+"e.txt", "game dir")
	writeFile(t, root+"/game/bin/"+dir+"f.txt", "gamebin")
	writeVPK(t, root+"/shared/pak01_dir.vpk", map[string]string{dir + "h.txt": "shared pak01"})
	writeFile(t, root+"/shared/"+dir+"g.txt", "shared dir")
	writeFile(t, root+"/shared/"+dir+"h.txt", "shared dir")
	writeFile(t, root+"/game/download/"+dir+"g.txt", "download")

	gameInfo := `"GameInfo"
{
	game	"Test"
	FileSystem
	{
		SearchPaths
		{
			game+mod			game/custom/*
			game+mod			game/game_pak.vpk
			game				|all_source_engine_paths|shared/shared_misc.vpk
			game				|all_source_engine_paths|shared/missing.vpk
			platform			|all_source_engine_paths|platform/platform_misc.vpk
			mod+mod_write+default_write_path		|gameinfo_path|.
			game+game_write		game
			gamebin				game/bin
			game				|all_source_engine_paths|shared
			game+download		game/download
		}
	}
}`
	reader := keyvalues.NewReader(bytes.NewBufferString(gameInfo))
	kv, err := reader.Read()
	if err != nil {
		t.Fatal(err)
	}

	fs, err := CreateFilesystemFromGameInfoDefinitions(root+"/game", &kv, true)
	if fs == nil {
		t.Fatalf("no filesystem created: %v", err)
	}
	badPaths, ok := err.(*InvalidResourcePathCollectionError)
	if !ok || !reflect.DeepEqual(badPaths.paths, []string{root + "/shared/missing.vpk"}) {
		t.Errorf("expected the missing vpk to be reported, got %v", err)
	}

	expectedOrder := []string{
		root + "/game/custom/a_addon.vpk",
		root + "/game/custom/b_addon",
		root + "/game/custom/c_addon_dir.vpk",
		root + "/game/game_pak.vpk",
		root + "/shared/shared_misc.vpk",
		root + "/game",
		root + "/shared/pak01_dir.vpk",
		root + "/shared",
		root + "/game/download",
	}
	if order := fs.EnumerateResourcePaths(); !reflect.DeepEqual(order, expectedOrder) {
		t.Errorf("unexpected search order:\n got  %v\n want %v", order, expectedOrder)
	}

	read := func(name string) string {
		f, err := fs.GetFile(dir + name)
		if err != nil {
			return "not found"
		}
		data, _ := io.ReadAll(f)
		return string(data)
	}
	cases := []struct {
		file, expected, reason string
	}{
		{"a.txt", "game vpk", "vpks listed before the game directory are searched first"},
		{"b.txt", "custom b dir", "custom directories are listed before the vpks"},
		{"c.txt", "custom c vpk", "custom vpks are listed before the game's vpks"},
		{"d.txt", "custom a vpk", "custom content is mounted in alphabetical order"},
		{"e.txt", "shared vpk", "every vpk listed before the game directory is searched first"},
		{"f.txt", "not found", "gamebin paths do not hold game content"},
		{"g.txt", "shared dir", "downloads are listed last"},
		{"h.txt", "shared pak01", "a directory's pak01 vpk is searched before its loose files"},
	}
	for _, c := range cases {
		if got := read(c.file); got != c.expected {
			t.Errorf("%s: got %q, want %q: %s", c.file, got, c.expected, c.reason)
		}
	}

	// A map's pakfile is searched before everything else
	fs.RegisterPakFile(testPakFile{dir + "a.txt": "pakfile"})
	if got := read("a.txt"); got != "pakfile" {
		t.Errorf("a.txt: got %q, want %q: the pakfile is searched first", got, "pakfile")
	}
}
