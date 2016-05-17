package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

const (
	VERSION         = "1.0.0"
	CHAR_UNDERSCORE = byte(95)
	CHAR_UPPER_A    = byte(65)
	CHAR_UPPER_Z    = byte(90)
	CHAR_LOWER_A    = byte(97)
	CHAR_LOWER_Z    = byte(122)
	CHAR_0          = byte(48)
	CHAR_9          = byte(57)
	CASE_DELTA      = 32 // Length between an uppercase letter and its lowercase counterpart in ascii
)

var el = log.New(os.Stderr, "", 0)
var b = make([]byte, 255) // key name buffer

func printVersion() {
	fmt.Fprintf(os.Stderr, "go-rsyslog-pstats %s\n", VERSION)
}

func printHelp() {
	printVersion()
	fmt.Fprintf(os.Stderr, "Parses and forwards rsyslog process stats to a local statsite or statsd process\n\n")
	flag.PrintDefaults()
}

func parseConfig() (port string) {
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

	udpAddr, err := net.ResolveUDPAddr("udp", "localhost:"+outPort)
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
	pos := 0
	last := CHAR_UNDERSCORE

	for _, r := range s {
		c := byte(r)

		// Only allow alpha numberic
		if c < CHAR_0 || (c > CHAR_9 && c < CHAR_UPPER_A) || (c > CHAR_UPPER_Z && c < CHAR_LOWER_A) || c > CHAR_LOWER_Z {
			// Don't have more than 1 underscore in a row
			if last == CHAR_UNDERSCORE {
				continue
			}

			c = CHAR_UNDERSCORE
		}

		// lower case any upper case characters
		if c >= CHAR_UPPER_A && c <= CHAR_UPPER_Z {
			c = byte(c + CASE_DELTA)
		}

		// Put the byte in the array
		b[pos] = c
		last = b[pos]
		pos++
	}

	// Remove a trailing underscore
	if b[pos-1] == CHAR_UNDERSCORE {
		pos--
	}

	return string(b[:pos])
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
