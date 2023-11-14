package main

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
)

const PATH = "/sys/firmware/memmap/"

type mmapentry struct {
	start uint64
	end   uint64
	size  float64
	kind  string
}

type mmaps []mmapentry

func (m mmaps) Len() int      { return len(m) }
func (m mmaps) Swap(i, j int) { m[i], m[j] = m[j], m[i] }
func (m mmaps) Less(i, j int) bool {
	return m[i].start < m[j].start
}

func processDir(dir fs.FileInfo, entries mmaps) mmaps {
	files, err := ioutil.ReadDir(PATH + dir.Name())
	if err != nil {
		log.Fatal(err)
	}
	var entry mmapentry
	for _, f := range files {
		filepath := PATH + dir.Name() + "/" + f.Name()
		switch f.Name() {
		case "start":
			data, _ := os.ReadFile(filepath)
			entry.start, err = strconv.ParseUint(strings.TrimSpace(string(data)), 0, 64)
			if err != nil {
				log.Fatal(err)
			}
		case "end":
			data, _ := os.ReadFile(filepath)
			entry.end, err = strconv.ParseUint(strings.TrimSpace(string(data)), 0, 64)
			if err != nil {
				log.Fatal(err)
			}
		case "type":
			data, _ := os.ReadFile(filepath)
			entry.kind = strings.TrimSpace(string(data))
		}
	}
	entry.size = float64(entry.end - entry.start)
	return append(entries, entry)
}

func prettyPrint(m mmaps) {
	const format = "%v\t%v\t%.2f\t%v\t\n"
	tw := tabwriter.NewWriter(os.Stdout, 10, 0, 2, ' ', 0)
	fmt.Fprintf(tw, "%v\t%v\t%v\t%v\t\n", "Start", "End", "Size (KiB)", "Type")
	fmt.Fprintf(tw, "%v\t%v\t%v\t%v\t\n", "--------", "--------", "--------", "--------")
	for _, v := range m {
		fmt.Fprintf(tw, format, fmt.Sprintf("0x%x", v.start), fmt.Sprintf("0x%x", v.end), (v.size / 1024), v.kind)
	}
	tw.Flush()
}

func main() {
	filesAndDirs, err := ioutil.ReadDir(PATH)
	var entries mmaps
	if err != nil {
		log.Fatal(err)
	}
	for _, element := range filesAndDirs {
		if element.IsDir() {
			entries = processDir(element, entries)
		}
	}

	sort.Sort(entries)
	prettyPrint(entries)
}
