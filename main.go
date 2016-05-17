package main

import (
	"os"
	"flag"
	"log"
	"bufio"
	"strings"
	"encoding/json"
	"fmt"
	"regexp"
	"net"
	"io"
)

const VERSION = "1.0.0"

var el = log.New(os.Stderr, "", 0)

// Removes any match, runs before any replacing
var firstRemover = regexp.MustCompile(`\(.+?\)`)
// Replaces any match with an underscore
var _Replacer = regexp.MustCompile(`[^a-z0-9_]|_+`)
// Removes any match, runs after any replacing
var finalRemover = regexp.MustCompile(`\_$`)

func printVersion() {
	fmt.Fprintf(os.Stderr, "go-rsyslog-pstats v%s\n", VERSION)
}

func printHelp() {
	printVersion()
	fmt.Fprintf(os.Stderr, "Parses and forwards rsyslog process stats to a local statsite or statsd process\n\n")
	flag.PrintDefaults()
}

func parseConfig () (port string) {
	flag.Usage = printHelp
	outPort := flag.String("port", "8215", "Statsite udp port to connect to")
	printV := flag.Bool("version", false, "Prints the version string")
	flag.Parse()

	if *printV {
		printVersion()
		os.Exit(0)
	}

	return *outPort
}

func main() {
	outPort := parseConfig()

	in := bufio.NewReader(os.Stdin)

	udpAddr, err := net.ResolveUDPAddr("udp", "localhost:" + outPort)
	if err != nil {
		el.Fatal("Could not resolve address", err)
	}

	out, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		el.Fatal("Could not dial address", err)
	}

	for {
		line, err := in.ReadString('\n')
		if err != nil {
			el.Fatalln("Failed to read line from input", err)
		}
		parseLine(line, out)
	}
}

// Use the global regex's to get the stat name ie key in working order
func sanitizeKey(s string) string {
	b := []byte(strings.ToLower(s))
	b = firstRemover.ReplaceAll(b, []byte(""))
	b = _Replacer.ReplaceAll(b, []byte("_"))
	return string(finalRemover.ReplaceAll(b, []byte("")))
}

// Take the entire json blob and find any key/value pairs whos value is number and formulate a stat entry
func findNums(prefix string, kvs map[string]interface{}, out io.Writer) {
	for k, v := range kvs {
		vf, ok := v.(float64)
		if !ok {
			continue
		}

		_, err := fmt.Fprintf(out, "rsyslog.%v.%v:%d|g\n", prefix, sanitizeKey(k), int(vf))
		if err != nil {
			el.Println("Error while writing", err)
		}
	}
}

// Format all the stats in the line
func parseLine(line string, out io.Writer) {
	var name string
	var ok bool

	jsonStart := strings.IndexByte(line, '{')
	if jsonStart < 0 {
		return
	}

	var values map[string]interface{}
	if err := json.Unmarshal([]byte(line[jsonStart:]), &values); err != nil {
		el.Println("Error while decoding json line", err)
		return
	}

	if _, ok = values["origin"]; !ok {
		el.Println("No origin key in the json blob", line)
		return
	}

	name, ok = values["name"].(string)
	if !ok {
		el.Println("No name key in the json blob", line)
		return
	}

	origin, ok := values["origin"].(string)
	if !ok {
		el.Println("No origin key in the json blob", line)
		return
	}

	name = sanitizeKey(origin) + "." + sanitizeKey(name)

	switch origin {
	case "dynstats":
		if vals, ok := values["values"].(map[string]interface{}); ok {
			findNums("dynstats", vals, out)
		}
	case "impstats":
		findNums("resource_usage", values, out)
	default:
		findNums(name, values, out)
	}
}
