package core

import (
	"bytes"
	"compress/gzip"
	"embed"
	"io"
	"runtime"
	"strings"
)

//go:embed blocklists
var BlockLists embed.FS

func LoadBlockLists() {
	GLOBAL_BLOCK_LIST = make(map[string]bool, 0)

	x, err := BlockLists.ReadDir("blocklists")
	if err != nil {
		CreateErrorLog("", "Unable to read blocklist directory: ", err)
		return
	}

	for _, v := range x {

		info, err := v.Info()
		if err != nil {
			CreateErrorLog("", "Unable to get info for blocklist: ", err)
			continue
		}

		name := strings.Split(info.Name(), ".")
		if len(name) < 3 {
			CreateErrorLog("", "Blocklist is invalid: ", info.Name())
			continue
		}

		CreateLog("", "Blocklist loaded: ", name[0], " with ", name[1], " domains")

		enabled := false
		for i := range C.EnabledBlockLists {
			if C.EnabledBlockLists[i] == name[0] {
				enabled = true
			}
		}

		GLOBAL_STATE.BLists = append(GLOBAL_STATE.BLists, &List{
			FullPath: info.Name(),
			Tag:      name[0],
			Domains:  name[1],
			Enabled:  enabled,
		})

	}

	BuildDomainBlocklist()

	return
}

func LoadFileIntoMap(path string) (list map[string]bool, err error) {
	defer RecoverAndLogToFile()
	file, err := BlockLists.Open(path)
	if err != nil {
		CreateErrorLog("", "Unable to open blocklist", path, err)
		return nil, err
	}

	defer func() {
		if r := recover(); r != nil {
			CreateErrorLog("", "Panic while processing blocklist", path, err)
		}
		if file != nil {
			file.Close()
		}
	}()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		CreateErrorLog("", "Unable to un-zip blocklist", path, err)
		return nil, err
	}

	fileBytes, err := io.ReadAll(gzipReader)
	if err != nil {
		CreateErrorLog("", "Unable to read blocklist", path, err)
		return nil, err
	}

	list = make(map[string]bool)

	fileLines := bytes.Split(fileBytes, []byte{10})
	for _, v := range fileLines {
		line := string(v)
		if len(line) < 1 {
			continue
		} else if line[0] == '#' {
			continue
		}
		list[line] = true
	}

	return
}

func BuildDomainBlocklist() {

	tempBlockList := make(map[string]bool, 0)

	for i := range GLOBAL_STATE.BLists {
		if GLOBAL_STATE.BLists[i].Enabled {
			list, err := LoadFileIntoMap("blocklists/" + GLOBAL_STATE.BLists[i].FullPath)

			if err != nil {
				CreateErrorLog("", "Error parsing blocklist: ", err)
				return
			}

			for i := range list {
				tempBlockList[i] = true
			}

		}
	}

	GLOBAL_BLOCK_LIST = make(map[string]bool, 0)
	GLOBAL_BLOCK_LIST = tempBlockList

	runtime.GC()
	CreateLog("", "New domain blocklist has been created using ", len(tempBlockList), " domains")
}
