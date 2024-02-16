package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type lsmon struct {
	Host string    `json:"host"`
	Time time.Time `json:"time"`
	Data []data    `json:"data"`
}

type data struct {
	Feature  map[string]string   `json:"Feature Information"`
	Licenses []map[string]string `json:"License Information"`
	Clients  []map[string]string `json:"Client Information"`
}

func main() {
	if len(os.Args) != 1 {
		fmt.Fprintln(os.Stderr, "Convert lsmon.exe output to JSON through standard I/O")
		os.Exit(1)
	}

	json, err := json.Marshal(loadLsmon(os.Stdin))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Print(string(json))
}

func loadLsmon(i io.Reader) lsmon {
	type scanmode int
	const (
		NONE scanmode = iota
		FEATURE
		LICENSE
		CLIENT
	)
	mode := NONE

	var lsmon lsmon
	var newdata data
	var kvs map[string]string

	s := bufio.NewScanner(i)
	for s.Scan() {
		// Check host name
		if strings.HasPrefix(strings.TrimSpace(s.Text()), `[Contacting Sentinel RMS Development Kit server on host "`) &&
			strings.HasSuffix(strings.TrimSpace(s.Text()), `"]`) {
			h := strings.SplitN(s.Text(), `"`, 3)
			if len(h) == 3 {
				lsmon.Host = h[1]
			}
			continue
		}
		// Load buffer data
		_, buf, f := strings.Cut(s.Text(), " |- ")
		if !f {
			continue
		}
		// Switch scan mode
		if strings.HasSuffix(buf, " Information") {
			if mode == FEATURE {
				newdata = data{Feature: kvs}
			} else if mode == LICENSE {
				newdata.Licenses = append(newdata.Licenses, kvs)
			} else if mode == CLIENT {
				newdata.Clients = append(newdata.Clients, kvs)
			}
			kvs = map[string]string{}
			if strings.HasPrefix(buf, "Feature ") {
				if mode != NONE {
					lsmon.Data = append(lsmon.Data, newdata)
				}
				mode = FEATURE
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
		kvs[strings.TrimSpace(k)] = strings.Trim(strings.TrimSpace(v), `"`)
	}

	if mode == FEATURE {
		newdata = data{Feature: kvs}
	} else if mode == LICENSE {
		newdata.Licenses = append(newdata.Licenses, kvs)
	} else if mode == CLIENT {
		newdata.Clients = append(newdata.Clients, kvs)
	}
	if mode != NONE {
		lsmon.Data = append(lsmon.Data, newdata)
	}

	lsmon.Time = time.Now()
	return lsmon
}
