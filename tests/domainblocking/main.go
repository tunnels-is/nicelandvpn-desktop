package main

import (
	"bytes"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// go:embed blocklists/*
// var BlockLists embed.FS

func main() {
	// CombineAllLists()
	ZipAndMoveCombinedLists()

	// LoadBlockLists()
	// BuildDomainBlocklist()
	// CleanListV2("originals/fakenews.txt")
	// CleanListV2("originals/porn.txt")
	// CleanListV2("originals/gambling.txt")
	// CleanListV2("originals/social.txt")
}

func ZipAndMoveCombinedLists() {
	log.Println("ZIPPING LISTS")

	filepath.WalkDir("combined", func(path string, d fs.DirEntry, err error) error {

		// cmd := exec.Command("rm", path+".gz")
		// out, err := cmd.CombinedOutput()
		// log.Println(string(out), err)

		cmd := exec.Command("cp", path, path+".bak")
		out, err := cmd.CombinedOutput()
		log.Println(string(out), err)

		cmd = exec.Command("gzip", "-9", path)
		out, err = cmd.CombinedOutput()
		log.Println(string(out), err)

		cmd = exec.Command("mv", path+".bak", path)
		out, err = cmd.CombinedOutput()
		log.Println(string(out), err)

		return nil
	})

}

func CombineAllLists() {
	var currentDirectory = ""
	var fileMap = make(map[string]map[string]bool)

	filepath.WalkDir("combine", func(path string, d fs.DirEntry, err error) error {
		// info, err := d.Info()

		if d.IsDir() {

			if path != currentDirectory {

				if path == "combine" {
					return nil
				}

				splitPath := strings.Split(path, "\\")
				currentDirectory = splitPath[1]

			}

			return nil
		}

		log.Println("Parsing file:", path)
		file, err := os.Open(path)
		fileBytes, err := io.ReadAll(file)

		fileLines := bytes.Split(fileBytes, []byte{10})
		for _, v := range fileLines {
			line := string(v)
			if len(line) < 1 {
				// log.Println("SKIPPED")
				continue
			} else if line[0] == '#' {
				continue
			}
			splitLines := bytes.Split(v, []byte(" "))
			lineIndex := 1
			if len(splitLines) < 2 {
				lineIndex = 0
			}

			_, ok := fileMap[currentDirectory]
			if !ok {
				fileMap[currentDirectory] = make(map[string]bool, 0)
			}

			fileMap[currentDirectory][string(splitLines[lineIndex])] = true
			// currentDomainList[string(splitLines[lineIndex])] = true
		}

		file.Close()

		return nil
	})

	// var currentDomainList = make(map[string]bool)
	os.Remove("combined")
	os.Mkdir("combined", 0777)

	for i := range fileMap {
		if len(fileMap[i]) < 1 {
			log.Println("SKIPPING", i)
			continue
		}

		log.Println("NEW FILE:", "combined\\"+i+"."+strconv.Itoa(len(fileMap[i])))

		newFile, err := os.Create("combined\\" + i + "." + strconv.Itoa(len(fileMap[i])))
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}

		for i := range fileMap[i] {
			newFile.Write([]byte(i))
			newFile.Write([]byte{10})
		}

		newFile.Close()

	}

}

// func BuildDomainBlocklist() {

// 	for i := range BLists {
// 		BLists[i].Enabled = true
// 	}

// 	tempBlockList := make(map[string]bool)

// 	for i := range BLists {
// 		if BLists[i].Enabled {
// 			list, err := LoadFileIntoMap("blocklists/" + BLists[i].FullPath)

// 			if err != nil {
// 				log.Println(err)
// 				return
// 			}

// 			for i := range list {
// 				tempBlockList[i] = true
// 			}

// 		}
// 	}

// 	log.Println("TL:", len(tempBlockList))

// }

// var BLists []*List

// type List struct {
// 	FullPath string
// 	Tag      string
// 	Enabled  bool
// 	Domains  string
// }

// func LoadBlockLists() {
// 	x, err := BlockLists.ReadDir("blocklists")
// 	if err != nil {
// 		CreateErrorLog("", "Unable to read blocklist directory: ", err)
// 		return
// 	}

// 	for _, v := range x {
// 		info, err := v.Info()
// 		if err != nil {
// 			CreateErrorLog("", "Unable to get info for blocklist: ", err)
// 			continue
// 		}

// 		name := strings.Split(info.Name(), ".")
// 		if len(name) < 3 {
// 			CreateErrorLog("", "Blocklist is invalid: ", info.Name())
// 			continue
// 		}

// 		BLists = append(BLists, &List{
// 			FullPath: info.Name(),
// 			Tag:      name[0],
// 			Domains:  name[1],
// 			Enabled:  false,
// 		})

// 	}

// 	for _, v := range BLists {
// 		log.Println(v)
// 	}

// 	return
// }

func CleanListV2(path string) {
	file, _ := os.Open(path)
	fileBytes, _ := io.ReadAll(file)
	fileLines := bytes.Split(fileBytes, []byte{10})

	lineMap := make(map[string]bool)

	for _, v := range fileLines {
		line := string(v)
		if len(line) < 1 {
			// log.Println("SKIPPED")
			continue
		} else if line[0] == '#' {
			continue
		}
		splitLines := bytes.Split(v, []byte(" "))
		if len(splitLines) < 2 {
			continue
		}

		splitDomain := bytes.Split(splitLines[1], []byte("."))
		if string(splitDomain[0]) == "www" {
			lineMap[string(splitLines[1][4:])] = true
			lineMap[string(splitLines[1])] = true
		} else {
			lineMap[string(splitLines[1])] = true
		}

	}

	file.Close()

	newPath := strings.Replace(path, "originals", "blocklists", -1)
	newPath = strings.Replace(newPath, ".txt", "."+strconv.Itoa(len(lineMap)), -1)
	newFile, _ := os.Create(newPath)

	for i := range lineMap {
		newFile.WriteString(i + "\n")
	}
	newFile.Close()

	cmd := exec.Command("gzip", "-9", newPath)
	out, err := cmd.CombinedOutput()
	log.Println(out, err)

}

func CreateErrorLog(x ...any) {
	log.Println(x)
}

// func LoadFileIntoMap(path string) (list map[string]bool, err error) {
// 	// defer RecoverAndLogToFile()
// 	file, err := BlockLists.Open(path)
// 	if err != nil {
// 		CreateErrorLog("", "Unable to open blocklist", path, err)
// 		return nil, err
// 	}

// 	defer func() {
// 		if r := recover(); r != nil {
// 			CreateErrorLog("", "Panic while processing blocklist", path, err)
// 		}
// 		if file != nil {
// 			file.Close()
// 		}
// 	}()

// 	gzipReader, err := gzip.NewReader(file)
// 	if err != nil {
// 		CreateErrorLog("", "Unable to un-zip blocklist", path, err)
// 		return nil, err
// 	}

// 	fileBytes, err := io.ReadAll(gzipReader)
// 	if err != nil {
// 		CreateErrorLog("", "Unable to read blocklist", path, err)
// 		return nil, err
// 	}

// 	list = make(map[string]bool)

// 	fileLines := bytes.Split(fileBytes, []byte{10})
// 	for _, v := range fileLines {
// 		line := string(v)
// 		if len(line) < 1 {
// 			continue
// 		} else if line[0] == '#' {
// 			continue
// 		}
// 		list[line] = true
// 	}

// 	return
// }
