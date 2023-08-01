package core

import (
	"bufio"
	"os"
)

func LoadBlockList() {
	file, err := os.Open("blocklist")
	if err != nil {
		CreateErrorLog("loader", "Unable to open blocklist: ", err)
		return
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		BlockedDomainMap[scanner.Text()] = true
	}

	if err := scanner.Err(); err != nil {
		CreateErrorLog("loader", "Error reading blocklist")
	}
}
