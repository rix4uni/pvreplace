package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/rix4uni/pvreplace/banner"
	"gopkg.in/yaml.v3"
)

// Config represents the root structure of the YAML config file
type Config struct {
	Configurations []FuzzingConfig `yaml:"configurations"`
}

// FuzzingConfig represents a single fuzzing configuration
type FuzzingConfig struct {
	FuzzingPart string `yaml:"fuzzing-part"`
	FuzzingType string `yaml:"fuzzing-type"`
	FuzzingMode string `yaml:"fuzzing-mode"`
	Ignore      bool   `yaml:"ignore,omitempty"`
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
	fuzzingPart := flag.String("fuzzing-part", "param-value", "Fuzzing part: param-value, param-name, path-suffix, path-suffix-slash, path-segment, path-ext, headers, all")
	config := flag.String("config", "", "Path to YAML config file with fuzzing configurations")
	output := flag.String("output", "", "Directory to save modified requests (default: ~/.config/pvreplace/modified_request)")
	silent := flag.Bool("silent", false, "Silent mode.")
	version := flag.Bool("version", false, "Print the version of the tool and exit.")
	verbose := flag.Bool("verbose", false, "Show detailed information about what's being processed.")
	flag.Parse()

	// Define all available fuzzing parts
	allFuzzingParts := []string{"param-value", "param-name", "path-suffix", "path-suffix-slash", "path-segment", "path-ext", "headers"}

	// Function to load and parse config file
	loadConfig := func(configPath string) ([]FuzzingConfig, error) {
		file, err := os.Open(configPath)
		if err != nil {
			return nil, fmt.Errorf("error opening config file: %v", err)
		}
		defer file.Close()

		var config Config
		decoder := yaml.NewDecoder(file)
		if err := decoder.Decode(&config); err != nil {
			return nil, fmt.Errorf("error parsing config file: %v", err)
		}

		// Filter out ignored configurations
		var activeConfigs []FuzzingConfig
		for _, cfg := range config.Configurations {
			if !cfg.Ignore {
				activeConfigs = append(activeConfigs, cfg)
			}
		}

		return activeConfigs, nil
	}

	// Print version and exit if -version flag is provided
	if *version {
		banner.PrintBanner()
		banner.PrintVersion()
		return
	}

	// Don't Print banner if -silnet flag is provided
	if !*silent {
		banner.PrintBanner()
	}

	// Validate that -ignore-lines can only be used with -raw flag
	if *ignoreLines != "" && *raw == "" {
		fmt.Fprintf(os.Stderr, "Error: -ignore-lines flag can only be used with -raw flag\n")
		os.Exit(1)
	}

	// Validate that -output can only be used with -raw flag
	if *output != "" && *raw == "" {
		fmt.Fprintf(os.Stderr, "Error: -output flag can only be used with -raw flag\n")
		os.Exit(1)
	}

	// Validate that -config cannot be used with -fuzzing-mode, -fuzzing-type, or -fuzzing-part
	if *config != "" {
		var conflictingFlags []string
		flag.Visit(func(f *flag.Flag) {
			if f.Name == "fuzzing-mode" || f.Name == "fuzzing-type" || f.Name == "fuzzing-part" {
				conflictingFlags = append(conflictingFlags, "-"+f.Name)
			}
		})
		if len(conflictingFlags) > 0 {
			fmt.Fprintf(os.Stderr, "Error: -config flag cannot be used with %s flags\n", strings.Join(conflictingFlags, ", "))
			os.Exit(1)
		}
	}

	// Regular expressions for different fuzzing parts
	reValue := regexp.MustCompile(`=[^&\s]*`)                                                                 // For parameter values
	reName := regexp.MustCompile(`([?&])([^&=]+)=`)                                                           // For parameter names
	rePathSuffix := regexp.MustCompile(`/([^/]+\.(php|asp|aspx|jsp|jspx|xml))`)                               // For URL paths
	rePathSegment := regexp.MustCompile(`(https?://(?:[^/]+/)+)([^/]+)/([^/]+\.(php|aspx|asp|jsp|jspx|xml))`) // For path segment
	rePathExt := regexp.MustCompile(`/([^/]+)\.(php|aspx|asp|jsp|jspx|xml)`)                                  // For file extensions in paths
	reUserAgent := regexp.MustCompile(`^(User-Agent:\s)(.*)$`)                                                // For matching headers
	reHeader := regexp.MustCompile(`^(User-Agent|Referer|Cookie|X-Forwarded-For|X-Real-IP):\s*(.*)$`)         // For matching injectable headers

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

	// Function to get the default ignore-lines.txt path
	getDefaultIgnoreLinesPath := func() (string, error) {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("error getting home directory: %v", err)
		}
		configDir := filepath.Join(homeDir, ".config", "pvreplace")
		return filepath.Join(configDir, "ignore-lines.txt"), nil
	}

	// Function to get the default config.yaml path
	getDefaultConfigPath := func() (string, error) {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("error getting home directory: %v", err)
		}
		configDir := filepath.Join(homeDir, ".config", "pvreplace")
		return filepath.Join(configDir, "config.yaml"), nil
	}

	// Function to download ignore-lines.txt from GitHub
	downloadIgnoreLines := func(filePath string) error {
		url := "https://raw.githubusercontent.com/rix4uni/pvreplace/refs/heads/main/ignore-lines.txt"

		resp, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("error downloading ignore-lines.txt: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("error downloading ignore-lines.txt: HTTP %d", resp.StatusCode)
		}

		// Create directory if it doesn't exist
		dir := filepath.Dir(filePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("error creating config directory: %v", err)
		}

		// Create the file
		file, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("error creating ignore-lines.txt: %v", err)
		}
		defer file.Close()

		// Write the downloaded content to file
		_, err = io.Copy(file, resp.Body)
		if err != nil {
			return fmt.Errorf("error writing ignore-lines.txt: %v", err)
		}

		if *verbose {
			fmt.Fprintf(os.Stderr, "[+] Downloaded ignore-lines.txt to: %s\n", filePath)
		}
		return nil
	}

	// Function to download config.yaml from GitHub
	downloadConfig := func(filePath string) error {
		url := "https://raw.githubusercontent.com/rix4uni/pvreplace/refs/heads/main/config.yaml"

		resp, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("error downloading config.yaml: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("error downloading config.yaml: HTTP %d", resp.StatusCode)
		}

		// Create directory if it doesn't exist
		dir := filepath.Dir(filePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("error creating config directory: %v", err)
		}

		// Create the file
		file, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("error creating config.yaml: %v", err)
		}
		defer file.Close()

		// Write the downloaded content to file
		_, err = io.Copy(file, resp.Body)
		if err != nil {
			return fmt.Errorf("error writing config.yaml: %v", err)
		}

		if *verbose {
			fmt.Fprintf(os.Stderr, "[+] Downloaded config.yaml to: %s\n", filePath)
		}
		return nil
	}

	// Function to ensure default ignore-lines.txt exists
	ensureDefaultIgnoreLines := func() (string, error) {
		defaultPath, err := getDefaultIgnoreLinesPath()
		if err != nil {
			return "", err
		}

		// Check if file exists
		if _, err := os.Stat(defaultPath); os.IsNotExist(err) {
			// Download the file from GitHub
			if err := downloadIgnoreLines(defaultPath); err != nil {
				return "", err
			}
		}

		return defaultPath, nil
	}

	// Function to ensure default config.yaml exists
	ensureDefaultConfig := func() (string, error) {
		defaultPath, err := getDefaultConfigPath()
		if err != nil {
			return "", err
		}

		// Check if file exists
		if _, err := os.Stat(defaultPath); os.IsNotExist(err) {
			// Download the file from GitHub
			if err := downloadConfig(defaultPath); err != nil {
				return "", err
			}
		}

		return defaultPath, nil
	}

	// Function to get the default output directory path
	getDefaultOutputPath := func() (string, error) {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("error getting home directory: %v", err)
		}
		return filepath.Join(homeDir, ".config", "pvreplace", "modified_request"), nil
	}

	// Function to ensure output directory exists
	ensureOutputDir := func(outputPath string) error {
		if err := os.MkdirAll(outputPath, 0755); err != nil {
			return fmt.Errorf("error creating output directory: %v", err)
		}
		return nil
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

		case "path-suffix-slash":
			if mode == "multiple" {
				switch ftype {
				case "replace":
					modifiedURL = rePathSuffix.ReplaceAllString(url, "${0}/"+payload)
				default:
					fmt.Fprintf(os.Stderr, "Invalid fuzzing type: %s (path-suffix-slash only supports replace)\n", ftype)
					return
				}
				fmt.Println(modifiedURL)
			} else if mode == "single" {
				for _, match := range rePathSuffix.FindAllStringIndex(url, -1) {
					switch ftype {
					case "replace":
						modifiedURL = url[:match[1]] + "/" + payload + url[match[1]:]
					default:
						fmt.Fprintf(os.Stderr, "Invalid fuzzing type: %s (path-suffix-slash only supports replace)\n", ftype)
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
				if *verbose {
					fmt.Println("You cannot use -fuzzing-mode single with -fuzzing-part path-segment")
				}
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
				if *verbose {
					fmt.Println("You cannot use -fuzzing-mode single with -fuzzing-part path-segment")
				}
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
				if *verbose {
					fmt.Println("You cannot use -fuzzing-mode single with -fuzzing-part headers")
				}
			}

		default:
			fmt.Fprintf(os.Stderr, "Invalid fuzzing part: %s\n", part)
			return
		}
	}

	// Function to check if a line is an injectable header
	isInjectableHeader := func(line string) bool {
		return reHeader.MatchString(line)
	}

	// Function to add payload to injectable headers
	fuzzHeader := func(line, payload string) string {
		if reHeader.MatchString(line) {
			return reHeader.ReplaceAllString(line, "${1}: ${2}"+payload)
		}
		return line
	}

	// Handle URL passed via the `-u` flag
	if *url != "" {
		payloads, err := getPayloads(*payload)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return
		}

		// Load config if provided or use default
		var configs []FuzzingConfig
		configPath := *config
		if configPath == "" {
			// Use default config path
			defaultPath, err := ensureDefaultConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Could not use default config.yaml: %v\n", err)
			} else {
				configPath = defaultPath
			}
		}

		if configPath != "" {
			loadedConfigs, err := loadConfig(configPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				return
			}
			configs = loadedConfigs
		}

		for _, p := range payloads {
			if len(configs) > 0 {
				// Process with each config from file
				for _, cfg := range configs {
					if cfg.FuzzingPart == "all" {
						for _, part := range allFuzzingParts {
							processURL(*url, strings.TrimSpace(p), cfg.FuzzingMode, cfg.FuzzingType, part)
						}
					} else {
						processURL(*url, strings.TrimSpace(p), cfg.FuzzingMode, cfg.FuzzingType, cfg.FuzzingPart)
					}
				}
			} else {
				// Use flag-based configuration
				if *fuzzingPart == "all" {
					for _, part := range allFuzzingParts {
						processURL(*url, strings.TrimSpace(p), *fuzzingMode, *fuzzingType, part)
					}
				} else {
					processURL(*url, strings.TrimSpace(p), *fuzzingMode, *fuzzingType, *fuzzingPart)
				}
			}
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

		// Load config if provided or use default
		var configs []FuzzingConfig
		configPath := *config
		if configPath == "" {
			// Use default config path
			defaultPath, err := ensureDefaultConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Could not use default config.yaml: %v\n", err)
			} else {
				configPath = defaultPath
			}
		}

		if configPath != "" {
			loadedConfigs, err := loadConfig(configPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				return
			}
			configs = loadedConfigs
		}

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			url := scanner.Text()
			for _, p := range payloads {
				if len(configs) > 0 {
					// Process with each config from file
					for _, cfg := range configs {
						if cfg.FuzzingPart == "all" {
							for _, part := range allFuzzingParts {
								processURL(url, strings.TrimSpace(p), cfg.FuzzingMode, cfg.FuzzingType, part)
							}
						} else {
							processURL(url, strings.TrimSpace(p), cfg.FuzzingMode, cfg.FuzzingType, cfg.FuzzingPart)
						}
					}
				} else {
					// Use flag-based configuration
					if *fuzzingPart == "all" {
						for _, part := range allFuzzingParts {
							processURL(url, strings.TrimSpace(p), *fuzzingMode, *fuzzingType, part)
						}
					} else {
						processURL(url, strings.TrimSpace(p), *fuzzingMode, *fuzzingType, *fuzzingPart)
					}
				}
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		}
		return
	}

	// Handle Burp Suite raw request data passed via the `-raw` flag
	if *raw != "" {
		// Check if the path is a directory or a file
		fileInfo, err := os.Stat(*raw)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error accessing raw request path: %v\n", err)
			return
		}

		var filesToProcess []string

		if fileInfo.IsDir() {
			// Read all files from the directory
			entries, err := os.ReadDir(*raw)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading directory: %v\n", err)
				return
			}

			for _, entry := range entries {
				if !entry.IsDir() {
					filesToProcess = append(filesToProcess, filepath.Join(*raw, entry.Name()))
				}
			}

			if len(filesToProcess) == 0 {
				fmt.Fprintf(os.Stderr, "Error: No files found in directory: %s\n", *raw)
				return
			}
		} else {
			// Single file
			filesToProcess = append(filesToProcess, *raw)
		}

		payloads, err := getPayloads(*payload)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return
		}

		// Prepare the ignore lines set
		ignoreSet := make(map[string]bool)

		// Determine which ignore-lines file to use
		ignoreLinesPath := *ignoreLines
		if ignoreLinesPath == "" {
			// Use default path if --ignore-lines is not provided
			defaultPath, err := ensureDefaultIgnoreLines()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Could not use default ignore-lines.txt: %v\n", err)
			} else {
				ignoreLinesPath = defaultPath
			}
		}

		// Load ignore patterns
		if ignoreLinesPath != "" {
			lines, err := getIgnoreLines(ignoreLinesPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				return
			}
			for _, line := range lines {
				ignoreSet[strings.TrimSpace(line)] = true
			}
		}

		// Determine output directory
		var outputDir string
		var useOutputDir bool
		if *output != "" {
			outputDir = *output
			useOutputDir = true
		} else {
			// Use default output directory
			defaultOutput, err := getDefaultOutputPath()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Could not get default output path: %v\n", err)
			} else {
				outputDir = defaultOutput
				useOutputDir = true
			}
		}

		// Create output directory if needed
		if useOutputDir {
			if err := ensureOutputDir(outputDir); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				return
			}
		}

		// Process each file
		for _, filePath := range filesToProcess {
			content, err := os.ReadFile(filePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", filePath, err)
				continue
			}

			// Prepare output file if output directory is specified
			var outputFile *os.File
			if useOutputDir {
				baseFileName := filepath.Base(filePath)
				outputFilePath := filepath.Join(outputDir, baseFileName)
				outputFile, err = os.Create(outputFilePath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error creating output file %s: %v\n", outputFilePath, err)
					continue
				}
				defer outputFile.Close()

				if *verbose {
					fmt.Fprintf(os.Stderr, "[+] Saving modified request to: %s\n", outputFilePath)
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

					var modifiedLine string
					// If the line is ignored, print it as-is without fuzzing
					if ignore {
						modifiedLine = line
					} else if isInjectableHeader(line) {
						// If it's an injectable header, append payload to the end
						modifiedLine = fuzzHeader(line, strings.TrimSpace(payload))
					} else {
						// Check if line contains parameters or needs fuzzing
						if reValue.MatchString(line) {
							modifiedLine = reValue.ReplaceAllString(line, "="+strings.TrimSpace(payload))
						} else {
							modifiedLine = line
						}
					}

					// Print to stdout
					fmt.Println(modifiedLine)

					// Write to output file if specified
					if outputFile != nil {
						fmt.Fprintln(outputFile, modifiedLine)
					}
				}

				if err := scanner.Err(); err != nil {
					fmt.Fprintf(os.Stderr, "Error reading raw data: %v\n", err)
				}

				// Separate different payload outputs with a newline
				fmt.Println()
				if outputFile != nil {
					fmt.Fprintln(outputFile)
				}
			}
		}
		return
	}

	// Handle URLs passed via standard input (pipe)
	payloads, err := getPayloads(*payload)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}

	// Load config if provided or use default
	var configs []FuzzingConfig
	configPath := *config
	if configPath == "" {
		// Use default config path
		defaultPath, err := ensureDefaultConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not use default config.yaml: %v\n", err)
		} else {
			configPath = defaultPath
		}
	}

	if configPath != "" {
		loadedConfigs, err := loadConfig(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return
		}
		configs = loadedConfigs
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		url := scanner.Text()
		for _, p := range payloads {
			if len(configs) > 0 {
				// Process with each config from file
				for _, cfg := range configs {
					if cfg.FuzzingPart == "all" {
						for _, part := range allFuzzingParts {
							processURL(url, strings.TrimSpace(p), cfg.FuzzingMode, cfg.FuzzingType, part)
						}
					} else {
						processURL(url, strings.TrimSpace(p), cfg.FuzzingMode, cfg.FuzzingType, cfg.FuzzingPart)
					}
				}
			} else {
				// Use flag-based configuration
				if *fuzzingPart == "all" {
					for _, part := range allFuzzingParts {
						processURL(url, strings.TrimSpace(p), *fuzzingMode, *fuzzingType, part)
					}
				} else {
					processURL(url, strings.TrimSpace(p), *fuzzingMode, *fuzzingType, *fuzzingPart)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
	}
}
