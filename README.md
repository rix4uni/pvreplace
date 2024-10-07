## pvreplace
 
pvreplace accept URLs on stdin, replace all query string values with a user-supplied value, only output each combination of query string parameters once per host and path.

## Installation
```
go install github.com/rix4uni/pvreplace@latest
```

## Download prebuilt binaries
```
wget https://github.com/rix4uni/pvreplace/releases/download/v0.0.1/pvreplace-linux-amd64-0.0.1.tgz
tar -xvzf pvreplace-linux-amd64-0.0.1.tgz
rm -rf pvreplace-linux-amd64-0.0.1.tgz
mv pvreplace ~/go/bin/pvreplace
```
Or download [binary release](https://github.com/rix4uni/pvreplace/releases) for your platform.

## Compile from source
```
git clone --depth 1 github.com/rix4uni/pvreplace.git
cd pvreplace; go install
```

## Usage
```console
Usage of pvreplace:
  -fuzzing-mode string
        Fuzzing mode: single, multiple (default "multiple")
  -fuzzing-part string
        Fuzzing part: param-value, param-name, path-suffix, path-segment (default "param-value")
  -fuzzing-type string
        Fuzzing type: replace, prefix, postfix (default "replace")
  -ignore-lines string
        Comma-separated list or file of lines to ignore in raw data
  -list string
        File containing URLs to process
  -payload string
        Comma-separated list of payloads or a file with payloads (default "FUZZ")
  -raw string
        File containing Burp Suite raw request data to process
  -u string
        The URL to process
```

## Fuzzing-Part
`param-value (default) - fuzz param-value for URL`
```
▶ echo "http://testphp.vulnweb.com/artists.php?artist=1&id=2" | pvreplace -fuzzing-part param-value
http://testphp.vulnweb.com/artists.php?artist=FUZZ&id=FUZZ
```

`param-name - fuzz param-name for URL`
```
▶ echo "http://testphp.vulnweb.com/artists.php?artist=1&id=2" | pvreplace -fuzzing-part param-name
http://testphp.vulnweb.com/artists.php?FUZZ=1&FUZZ=2
```

`path-suffix - fuzz path-suffix for URL`
```
▶ echo "http://testphp.vulnweb.com/artists.php?artist=1&id=2" | pvreplace -fuzzing-part path-suffix
http://testphp.vulnweb.com/artists.phpFUZZ?artist=1&id=2
```

`path-segment - fuzz path-segment for URL`
```
▶ echo "http://testphp.vulnweb.com/wp-admin/admin-ajax.php" | pvreplace -fuzzing-part path-segment
http://testphp.vulnweb.com/wp-adminFUZZ/admin-ajax.php
```

`ext-filename - fuzz ext-filename for URL`
```
▶ echo "http://testphp.vulnweb.com/artists.php?artist=1&id=2" | pvreplace -fuzzing-part ext-filename
http://testphp.vulnweb.com/FUZZ.php?artist=1&id=2
```

## Fuzzing-Type
`replace (default) - replace the value with payload`
```
▶ echo "http://testphp.vulnweb.com/artists.php?artist=1&id=2" | pvreplace -fuzzing-type replace
http://testphp.vulnweb.com/artists.php?artist=FUZZ&id=FUZZ
```

`prefix - prefix the value with payload`
```
▶ echo "http://testphp.vulnweb.com/artists.php?artist=1&id=2" | pvreplace -fuzzing-type prefix
http://testphp.vulnweb.com/artists.php?artist=FUZZ1&id=FUZZ2
```

`postfix - postfix the value with payload`
```
▶ echo "http://testphp.vulnweb.com/artists.php?artist=1&id=2" | pvreplace -fuzzing-type postfix
http://testphp.vulnweb.com/artists.php?artist=1FUZZ&id=2FUZZ
```

## Fuzzing-Mode
`multiple (default) - replace all values at once`
```
▶ echo "http://testphp.vulnweb.com/artists.php?artist=1&id=2" | pvreplace -fuzzing-mode multiple
http://testphp.vulnweb.com/artists.php?artist=FUZZ&id=FUZZ
```

`single - replace one value at a time`
```
▶ echo "http://testphp.vulnweb.com/artists.php?artist=1&id=2" | pvreplace -fuzzing-mode single
http://testphp.vulnweb.com/artists.php?artist=FUZZ&id=2
http://testphp.vulnweb.com/artists.php?artist=1&id=FUZZ
```

## TODO
- use "github.com/projectdiscovery/goflags"