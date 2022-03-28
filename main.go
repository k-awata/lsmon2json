package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Feature struct {
	Host     string              `json:"Host Name"`
	Info     map[string]string   `json:"Feature Information"`
	Licenses []map[string]string `json:"License Information"`
	Clients  []map[string]string `json:"Client Information"`
}

func main() {
	if len(os.Args) != 1 {
		fmt.Fprintln(os.Stderr, "Convert lsmon.exe output to JSON through standard I/O")
		os.Exit(1)
	}
	type scanmode int
	const (
		NONE scanmode = iota
		INFO
		LICENSE
		CLIENT
	)
	mode := NONE
	features := []Feature{}
	temphost := ""
	var tempinfo map[string]string
	var templics []map[string]string
	var tempclis []map[string]string
	var temp map[string]string
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		// Check host name
		if strings.HasPrefix(strings.TrimSpace(s.Text()), `[Contacting Sentinel RMS Development Kit server on host "`) &&
			strings.HasSuffix(strings.TrimSpace(s.Text()), `"]`) {
			h := strings.SplitN(s.Text(), `"`, 3)
			if len(h) == 3 {
				temphost = h[1]
			}
			continue
		}
		_, buf, f := strings.Cut(s.Text(), " |- ")
		if !f {
			continue
		}
		// Change data hierarchy
		if strings.HasSuffix(buf, " Information") {
			if mode == INFO {
				tempinfo = temp
			} else if mode == LICENSE {
				templics = append(templics, temp)
			} else if mode == CLIENT {
				tempclis = append(tempclis, temp)
			}
			temp = map[string]string{}
			if strings.HasPrefix(buf, "Feature ") {
				if mode != NONE {
					features = append(features, Feature{Host: temphost, Info: tempinfo, Licenses: templics, Clients: tempclis})
				}
				tempinfo = map[string]string{}
				templics = []map[string]string{}
				tempclis = []map[string]string{}
				mode = INFO
			} else if strings.HasPrefix(buf, "License ") {
				mode = LICENSE
			} else if strings.HasPrefix(buf, "Client ") {
				mode = CLIENT
			}
			continue
		}
		// Set key and value
		k, v, f := strings.Cut(buf, ": ")
		if !f {
			continue
		}
		temp[strings.TrimSpace(k)] = strings.Trim(strings.TrimSpace(v), `"`)
	}
	if mode != NONE {
		if mode == INFO {
			tempinfo = temp
		} else if mode == LICENSE {
			templics = append(templics, temp)
		} else if mode == CLIENT {
			tempclis = append(tempclis, temp)
		}
		features = append(features, Feature{Host: temphost, Info: tempinfo, Licenses: templics, Clients: tempclis})
	}
	// Convert to JSON
	json, err := json.Marshal(features)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Print(string(json))
}
