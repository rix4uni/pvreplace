## pvreplace

A powerful URL parameter and request fuzzing tool that processes URLs or Burp Suite raw requests, replacing values with custom payloads while maintaining unique parameter combinations.

## 🚀 Quick Start

### Installation

**Using Go:**
```
go install github.com/rix4uni/pvreplace@latest
```

**Pre-built Binaries:**
```
wget https://github.com/rix4uni/pvreplace/releases/download/v0.0.7/pvreplace-linux-amd64-0.0.7.tgz
tar -xvzf pvreplace-linux-amd64-0.0.7.tgz
mv pvreplace ~/go/bin/
```

**From Source:**
```
git clone --depth 1 https://github.com/rix4uni/pvreplace.git
cd pvreplace; go install
```

## 📖 Usage

```console
Usage: pvreplace [OPTIONS]

Basic Options:
  -u string          Target URL to process
  -list string       File containing URLs to process
  -raw string        File/directory with Burp Suite raw requests
  -payload string    Payload(s) to use (default: "FUZZ")
  -silent            Suppress banner output
  -verbose           Show detailed processing information
  -version           Display version information

Fuzzing Options:
  -fuzzing-mode string    Fuzzing mode: single, multiple (default: "multiple")
  -fuzzing-part string    Fuzzing target: param-value, param-name, path-suffix, 
                          path-segment, path-ext, headers (default: "param-value")
  -fuzzing-type string    Fuzzing method: replace, prefix, postfix (default: "replace")

Advanced Options:
  -ignore-lines string   Lines to ignore in raw requests (comma-separated or file)
  -output string         Output directory for modified requests
```

## 🎯 Fuzzing Capabilities

### Fuzzing Parts

| Part | Description | Example |
|------|-------------|---------|
| **param-value** (default) | Fuzz parameter values | `?id=1` → `?id=FUZZ` |
| **param-name** | Fuzz parameter names | `?id=1` → `?FUZZ=1` |
| **path-suffix** | Fuzz path endings | `/page.php` → `/page.phpFUZZ` |
| **path-segment** | Fuzz path segments | `/admin/page` → `/adminFUZZ/page` |
| **path-ext** | Fuzz file extensions | `/script.php` → `/script.FUZZ` |
| **headers** | Fuzz HTTP headers | `User-Agent: Mozilla` → `User-Agent: MozillaFUZZ` |

### Fuzzing Types

| Type | Description | Example |
|------|-------------|---------|
| **replace** (default) | Replace entire value | `?id=123` → `?id=FUZZ` |
| **prefix** | Add payload before value | `?id=123` → `?id=FUZZ123` |
| **postfix** | Add payload after value | `?id=123` → `?id=123FUZZ` |

### Fuzzing Modes

| Mode | Description | Compatibility |
|------|-------------|---------------|
| **multiple** (default) | Replace all targets at once | All fuzzing parts |
| **single** | Replace one target at a time | Not compatible with: path-segment, path-ext, headers |

## 💡 Examples

### Basic URL Processing

```yaml
# Replace all parameter values
echo "http://example.com/page.php?id=1&name=test" | pvreplace
# Output: http://example.com/page.php?id=FUZZ&name=FUZZ

# Single mode - one parameter at a time
echo "http://example.com/page.php?id=1&name=test" | pvreplace -fuzzing-mode single
# Output:
# http://example.com/page.php?id=FUZZ&name=test
# http://example.com/page.php?id=1&name=FUZZ

# Custom payload
echo "http://example.com/page.php?id=1" | pvreplace -payload "' OR '1'='1"
```

### Path Fuzzing

```yaml
# Fuzz path segments
echo "http://example.com/admin/dashboard.php" | pvreplace -fuzzing-part path-segment
# Output: http://example.com/adminFUZZ/dashboard.php

# Fuzz file extensions
echo "http://example.com/script.php" | pvreplace -fuzzing-part path-ext
# Output: http://example.com/script.FUZZ
```

### Header Fuzzing

```yaml
echo "User-Agent: Mozilla/5.0" | pvreplace -fuzzing-part headers -fuzzing-type postfix
# Output: User-Agent: Mozilla/5.0FUZZ
```

## 🔧 Burp Suite Integration

### Process Raw Requests

```yaml
# Single file
pvreplace -raw request.txt

# Directory of requests
pvreplace -raw ./burp-requests/

# With custom output directory
pvreplace -raw request.txt -output ./modified-requests/

# Verbose mode to see processing details
pvreplace -raw request.txt -verbose
```

### Example Burp Request Processing

**Input (`burp-request.txt`):**
```yaml
POST /userinfo.php HTTP/1.1
Host: testphp.vulnweb.com
Content-Length: 20
Cache-Control: max-age=0
Origin: http://testphp.vulnweb.com
DNT: 1
Upgrade-Insecure-Requests: 1
Content-Type: application/x-www-form-urlencoded
User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36
Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7
Referer: http://testphp.vulnweb.com/login.php
Accept-Encoding: gzip, deflate
Accept-Language: en-US,en;q=0.9,hi;q=0.8,en-IN;q=0.7
Connection: close

uname=test&pass=test
```

**Command:**
```yaml
pvreplace -silent -raw burp-request.txt
```

**Output:**
```yaml
POST /userinfo.php HTTP/1.1
Host: testphp.vulnweb.com
Content-Length: 20
Cache-Control: max-age=0
Origin: http://testphp.vulnweb.com
DNT: 1
Upgrade-Insecure-Requests: 1
Content-Type: application/x-www-form-urlencoded
User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36FUZZ
Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7
Referer: http://testphp.vulnweb.com/login.php
Accept-Encoding: gzip, deflate
Accept-Language: en-US,en;q=0.9,hi;q=0.8,en-IN;q=0.7
Connection: close

uname=FUZZ&pass=FUZZ
```

### Ignore Lines Configuration

```yaml
# Auto-download default ignore list
pvreplace -raw request.txt

# Custom ignore lines
pvreplace -raw request.txt -ignore-lines "Host,Accept-Encoding,Connection"

# From file
pvreplace -raw request.txt -ignore-lines ignore-list.txt
```

## ⚙️ Advanced Usage

### Multiple Payloads

```yaml
# Comma-separated payloads
pvreplace -raw request.txt -payload "payload1,payload2,payload3"

# From file
pvreplace -raw request.txt -payload payloads.txt
```

### Batch Processing

```yaml
# Process URL list from file
pvreplace -list urls.txt

# Combined with custom fuzzing
pvreplace -list urls.txt -fuzzing-part param-name -fuzzing-type prefix
```

## 📋 Supported Features

### File Extensions
- `.php`, `.asp`, `.aspx`
- `.jsp`, `.jspx`, `.xml`

### Injectable Headers
- `User-Agent`, `Referer`, `Cookie`
- `X-Forwarded-For`, `X-Real-IP`

## ⚠️ Important Notes

- **Single mode limitations**: Not compatible with `path-segment`, `path-ext`, or `headers` fuzzing parts
- **Flag dependencies**: 
  - `-ignore-lines` and `-output` only work with `-raw` flag
  - Auto-downloads ignore list when using `-raw` without `-ignore-lines`
- **Output directory**: Defaults to `~/.config/pvreplace/modified_request/`
- The tool ensures unique parameter combinations per host and path

## 🔍 Verbose Output

Use `-verbose` to see detailed processing information:

```yaml
pvreplace -raw request.txt -verbose
# Shows: download confirmations, file save locations, processing stats
```
