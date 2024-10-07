package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// prints the version message
const version = "0.0.6"

func printVersion() {
	fmt.Printf("Current pvreplace version %s\n", version)
}

// Prints the Colorful banner
func printBanner() {
	banner := `
                                     __                 
    ____  _   __ _____ ___   ____   / /____ _ _____ ___ 
   / __ \| | / // ___// _ \ / __ \ / // __  // ___// _ \
  / /_/ /| |/ // /   /  __// /_/ // // /_/ // /__ /  __/
 / .___/ |___//_/    \___// .___//_/ \__,_/ \___/ \___/ 
/_/                      /_/                            
`
fmt.Printf("%s\n%70s\n\n", banner, "Current pvreplace version "+version)

}

func main() {
	// Define command-line flags
	payload := flag.String("payload", "FUZZ", "Comma-separated list of payloads or a file with payloads")
	url := flag.String("u", "", "The URL to process")
	list := flag.String("list", "", "File containing URLs to process")
	raw := flag.String("raw", "", "File containing Burp Suite raw request data to process")
	ignoreLines := flag.String("ignore-lines", "", "Comma-separated list or file of lines to ignore in raw data")
	fuzzingMode := flag.String("fuzzing-mode", "multiple", "Fuzzing mode: single, multiple")
	fuzzingType := flag.String("fuzzing-type", "replace", "Fuzzing type: replace, prefix, postfix")
	fuzzingPart := flag.String("fuzzing-part", "param-value", "Fuzzing part: param-value, param-name, path-suffix, path-segment")
	silent := flag.Bool("silent", false, "silent mode.")
	version := flag.Bool("version", false, "Print the version of the tool and exit.")
	flag.Parse()

	// Print version and exit if -version flag is provided
	if *version {
		printBanner()
		printVersion()
		return
	}

	// Don't Print banner if -silnet flag is provided
	if !*silent {
		printBanner()
	}

	// Regular expressions for different fuzzing parts
	reValue := regexp.MustCompile(`=[^&\s]*`)                                                                 // For parameter values
	reName := regexp.MustCompile(`([?&])([^&=]+)=`)                                                           // For parameter names
	rePathSuffix := regexp.MustCompile(`/([^/]+\.(php|asp|aspx|jsp|jspx|xml))`)                               // For URL paths
	rePathSegment := regexp.MustCompile(`(https?://(?:[^/]+/)+)([^/]+)/([^/]+\.(php|aspx|asp|jsp|jspx|xml))`) // For path segment
	rePathExt := regexp.MustCompile(`/([^/]+)\.(php|aspx|asp|jsp|jspx|xml)`)                                  // For file extensions in paths
	reUserAgent := regexp.MustCompile(`^(User-Agent:\s)(.*)$`)                                                // For matching headers

	// Function to read payloads from a file or comma-separated list
	getPayloads := func(input string) ([]string, error) {
		if strings.HasSuffix(input, ".txt") {
			file, err := os.Open(input)
			if err != nil {
				return nil, fmt.Errorf("error opening payload file: %v", err)
			}
			defer file.Close()

			var payloads []string
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				payload := strings.TrimSpace(scanner.Text())
				if payload != "" {
					payloads = append(payloads, payload)
				}
			}

			if err := scanner.Err(); err != nil {
				return nil, fmt.Errorf("error reading payload file: %v", err)
			}
			return payloads, nil
		}

		// Split comma-separated values
		return strings.Split(input, ","), nil
	}

	// Function to read ignore lines from a file or a comma-separated list
	getIgnoreLines := func(input string) ([]string, error) {
		if strings.HasSuffix(input, ".txt") {
			file, err := os.Open(input)
			if err != nil {
				return nil, fmt.Errorf("error opening ignore lines file: %v", err)
			}
			defer file.Close()

			var lines []string
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line != "" {
					lines = append(lines, line)
				}
			}

			if err := scanner.Err(); err != nil {
				return nil, fmt.Errorf("error reading ignore lines file: %v", err)
			}
			return lines, nil
		}

		// Split comma-separated values
		return strings.Split(input, ","), nil
	}

	// Function to process and replace parts of a URL based on fuzzing mode, type, and part
	processURL := func(url, payload, mode, ftype, part string) {
		var modifiedURL string

		switch part {
		case "param-value":
			if mode == "multiple" {
				switch ftype {
				case "replace":
					modifiedURL = reValue.ReplaceAllString(url, "="+payload)
				case "prefix":
					modifiedURL = reValue.ReplaceAllString(url, "="+payload+"${0}")
				case "postfix":
					modifiedURL = reValue.ReplaceAllString(url, "${0}"+payload)
				default:
					fmt.Fprintf(os.Stderr, "Invalid fuzzing type: %s\n", ftype)
					return
				}
				fmt.Println(modifiedURL)
			} else if mode == "single" {
				for _, match := range reValue.FindAllStringIndex(url, -1) {
					switch ftype {
					case "replace":
						modifiedURL = url[:match[0]] + "=" + payload + url[match[1]:]
					case "prefix":
						modifiedURL = url[:match[0]] + "=" + payload + url[match[0]+1:]
					case "postfix":
						modifiedURL = url[:match[0]] + url[match[0]:match[1]] + payload + url[match[1]:]
					default:
						fmt.Fprintf(os.Stderr, "Invalid fuzzing type: %s\n", ftype)
						return
					}
					fmt.Println(modifiedURL)
				}
			}

		case "param-name":
			if mode == "multiple" {
				switch ftype {
				case "replace":
					modifiedURL = reName.ReplaceAllString(url, "${1}"+payload+"=")
				case "prefix":
					modifiedURL = reName.ReplaceAllString(url, "${1}"+payload+"${2}=")
				case "postfix":
					modifiedURL = reName.ReplaceAllString(url, "${1}${2}"+payload+"=")
				default:
					fmt.Fprintf(os.Stderr, "Invalid fuzzing type: %s\n", ftype)
					return
				}
				fmt.Println(modifiedURL)
			} else if mode == "single" {
				for _, match := range reName.FindAllStringSubmatchIndex(url, -1) {
					switch ftype {
					case "replace":
						modifiedURL = url[:match[2]] + url[match[2]:match[3]] + payload + url[match[5]:]
					case "prefix":
						modifiedURL = url[:match[2]] + url[match[2]:match[4]] + payload + url[match[4]:]
					case "postfix":
						modifiedURL = url[:match[2]] + url[match[2]:match[5]] + payload + url[match[5]:]
					default:
						fmt.Fprintf(os.Stderr, "Invalid fuzzing type: %s\n", ftype)
						return
					}
					fmt.Println(modifiedURL)
				}
			}

		case "path-suffix":
			if mode == "multiple" {
				switch ftype {
				case "replace":
					modifiedURL = rePathSuffix.ReplaceAllString(url, "/"+payload)
				case "prefix":
					modifiedURL = rePathSuffix.ReplaceAllString(url, "/"+payload+"${1}")
				case "postfix":
					modifiedURL = rePathSuffix.ReplaceAllString(url, "${0}"+payload)
				default:
					fmt.Fprintf(os.Stderr, "Invalid fuzzing type: %s\n", ftype)
					return
				}
				fmt.Println(modifiedURL)
			} else if mode == "single" {
				for _, match := range rePathSuffix.FindAllStringIndex(url, -1) {
					switch ftype {
					case "replace":
						modifiedURL = url[:match[0]] + "/" + payload + url[match[1]:]
					case "prefix":
						modifiedURL = url[:match[0]+1] + payload + url[match[0]+1:]
					case "postfix":
						modifiedURL = url[:match[0]] + url[match[0]:match[1]] + payload + url[match[1]:]
					default:
						fmt.Fprintf(os.Stderr, "Invalid fuzzing type: %s\n", ftype)
						return
					}
					fmt.Println(modifiedURL)
				}
			}

		case "path-segment":
			if mode == "multiple" {
				switch ftype {
				case "replace":
					modifiedURL = rePathSegment.ReplaceAllString(url, "${1}"+payload+"/${3}")
				case "prefix":
					modifiedURL = rePathSegment.ReplaceAllString(url, "${1}"+payload+"${2}/${3}")
				case "postfix":
					modifiedURL = rePathSegment.ReplaceAllString(url, "${1}${2}"+payload+"/${3}")
				default:
					fmt.Fprintf(os.Stderr, "Invalid fuzzing type: %s\n", ftype)
					return
				}
				fmt.Println(modifiedURL)
			} else if mode == "single" {
				fmt.Println("You cannot use -fuzzing-mode single with -fuzzing-part path-segment")
			}

		case "path-ext":
			if mode == "multiple" {
				switch ftype {
				case "replace":
					modifiedURL = rePathExt.ReplaceAllString(url, "/${1}."+payload)
				case "prefix":
					modifiedURL = rePathExt.ReplaceAllString(url, "/${1}."+payload+"${2}")
				case "postfix":
					modifiedURL = rePathExt.ReplaceAllString(url, "${0}"+payload)
				default:
					fmt.Fprintf(os.Stderr, "Invalid fuzzing type: %s\n", ftype)
					return
				}
				fmt.Println(modifiedURL)
			} else if mode == "single" {
				fmt.Println("You cannot use -fuzzing-mode single with -fuzzing-part path-segment")
			}

		case "headers":
			if mode == "multiple" {
				switch ftype {
				case "replace":
					modifiedURL = reUserAgent.ReplaceAllString(url, "${1}"+payload)
				case "prefix":
					modifiedURL = reUserAgent.ReplaceAllString(url, "${1}"+payload+"${2}")
				case "postfix":
					modifiedURL = reUserAgent.ReplaceAllString(url, "${1}${2}"+payload)
				default:
					fmt.Fprintf(os.Stderr, "Invalid fuzzing type: %s\n", ftype)
					return
				}
				fmt.Println(modifiedURL)
			} else if mode == "single" {
				// Handle single replacement mode if needed
				fmt.Println("You cannot use -fuzzing-mode single with -fuzzing-part path-segment")
			}

		default:
			fmt.Fprintf(os.Stderr, "Invalid fuzzing part: %s\n", part)
			return
		}
	}

	// Handle URL passed via the `-u` flag
	if *url != "" {
		payloads, err := getPayloads(*payload)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return
		}

		for _, p := range payloads {
			processURL(*url, strings.TrimSpace(p), *fuzzingMode, *fuzzingType, *fuzzingPart)
		}
		return
	}

	// Handle URLs passed via the `-list` flag (file input)
	if *list != "" {
		file, err := os.Open(*list)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
			return
		}
		defer file.Close()

		payloads, err := getPayloads(*payload)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return
		}

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			url := scanner.Text()
			for _, p := range payloads {
				processURL(url, strings.TrimSpace(p), *fuzzingMode, *fuzzingType, *fuzzingPart)
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		}
		return
	}

	// Handle Burp Suite raw request data passed via the `-raw` flag
	if *raw != "" {
		content, err := os.ReadFile(*raw)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading raw request file: %v\n", err)
			return
		}

		payloads, err := getPayloads(*payload)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return
		}

		// Prepare the ignore lines set
		ignoreSet := make(map[string]bool)
		if *ignoreLines != "" {
			lines, err := getIgnoreLines(*ignoreLines)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				return
			}
			for _, line := range lines {
				ignoreSet[strings.TrimSpace(line)] = true
			}
		}

		// Process the content line by line
		for _, payload := range payloads {
			scanner := bufio.NewScanner(strings.NewReader(string(content)))
			for scanner.Scan() {
				line := scanner.Text()

				// Check if the line should be ignored
				ignore := false
				for prefix := range ignoreSet {
					if strings.HasPrefix(line, prefix) {
						ignore = true
						break
					}
				}

				// If the line is not ignored, replace parameter values
				if !ignore {
					processURL(line, strings.TrimSpace(payload), *fuzzingMode, *fuzzingType, *fuzzingPart)
				}
			}

			if err := scanner.Err(); err != nil {
				fmt.Fprintf(os.Stderr, "Error reading raw data: %v\n", err)
			}

			// Separate different payload outputs with a newline
			fmt.Println()
		}
		return
	}

	// Handle URLs passed via standard input (pipe)
	payloads, err := getPayloads(*payload)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		url := scanner.Text()
		for _, p := range payloads {
			processURL(url, strings.TrimSpace(p), *fuzzingMode, *fuzzingType, *fuzzingPart)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
	}
}
