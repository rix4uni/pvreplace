# pvreplace
 
Accept URLs on stdin, replace all query string values with a user-supplied value, only output
each combination of query string parameters once per host and path.

## Installation
```
git clone https://github.com/rix4uni/pvreplace.git ~/bin/pvreplace
echo "alias pvreplace='python3 ~/bin/pvreplace/pvreplace.py'" >> ~/.bashrc && source ~/.bashrc
```

## Usage
```
Usage: python3 pvreplace.py [strings] [-without-encode] [-part] [-type] [-mode] [-payload [strings or filepath]]

positional arguments:
  strings             The string(s) to be replaced in URLs (default: FUZZ)

options:
  -part              Specify which part of the URL to modify Options: param-value, param-name, path-suffix, path-param, ext-filename (default: param-value)
  -type              Specify the type of modification Options: replace, prefix, postfix (default: replace)
  -mode              Specify the mode of replacement Options: multiple, single (default: multiple)
  -payload           Specify payload(s) directly or from a file
  -without-encode    Optional argument to disable URL encoding (default: enabled)
  -v, --version      Prints current version
  -h, --help         Prints Help
```

## Part
`param-value (default) - fuzz param-value for URL`
```
▶ echo "http://testphp.vulnweb.com/artists.php?artist=1&id=2" | pvreplace -part param-value
http://testphp.vulnweb.com/artists.php?artist=FUZZ&id=FUZZ
```

`param-name - fuzz param-name for URL`
```
▶ echo "http://testphp.vulnweb.com/artists.php?artist=1&id=2" | pvreplace -part param-name
http://testphp.vulnweb.com/artists.php?FUZZ=1&FUZZ=2
```

`path-suffix - fuzz path-suffix for URL`
```
▶ echo "http://testphp.vulnweb.com/artists.php?artist=1&id=2" | pvreplace -part path-suffix
http://testphp.vulnweb.com/artists.phpFUZZ?artist=1&id=2
```

`path-param - fuzz path-param for URL`
```
▶ echo "http://testphp.vulnweb.com/artists.php?artist=1&id=2" | pvreplace -part path-param
http://testphp.vulnweb.com/artists.php/FUZZ?artist=1&id=2
http://testphp.vulnweb.com/artists.php?FUZZ&artist=1&id=2
```

`ext-filename - fuzz ext-filename for URL`
```
▶ echo "http://testphp.vulnweb.com/artists.php?artist=1&id=2" | pvreplace -part ext-filename
http://testphp.vulnweb.com/FUZZ.php?artist=1&id=2
```

## Type
`replace (default) - replace the value with payload`
```
▶ echo "http://testphp.vulnweb.com/artists.php?artist=1&id=2" | pvreplace -type replace
http://testphp.vulnweb.com/artists.php?artist=FUZZ&id=FUZZ
```

`prefix - prefix the value with payload`
```
▶ echo "http://testphp.vulnweb.com/artists.php?artist=1&id=2" | pvreplace -type prefix
http://testphp.vulnweb.com/artists.php?artist=FUZZ1&id=FUZZ2
```

`postfix - postfix the value with payload`
```
▶ echo "http://testphp.vulnweb.com/artists.php?artist=1&id=2" | pvreplace -type postfix
http://testphp.vulnweb.com/artists.php?artist=1FUZZ&id=2FUZZ
```

## Mode
`multiple (default) - replace all values at once`
```
▶ echo "http://testphp.vulnweb.com/artists.php?artist=1&id=2" | pvreplace -mode multiple
http://testphp.vulnweb.com/artists.php?artist=FUZZ&id=FUZZ
```

`single - replace one value at a time`
```
▶ echo "http://testphp.vulnweb.com/artists.php?artist=1&id=2" | pvreplace -mode single
http://testphp.vulnweb.com/artists.php?artist=FUZZ&id=2
http://testphp.vulnweb.com/artists.php?artist=1&id=FUZZ
```

## Payload without encode
```
▶ echo "http://testphp.vulnweb.com/artists.php?artist=1&id=2" | pvreplace -payload '"><script>confirm(1)</script>, "<image/src/onerror=confirm(1)>' -without-encode
http://testphp.vulnweb.com/artists.php?artist="><script>confirm(1)</script>&id="><script>confirm(1)</script>
http://testphp.vulnweb.com/artists.php?artist="<image/src/onerror=confirm(1)>&id="<image/src/onerror=confirm(1)>

or

▶ echo "http://testphp.vulnweb.com/artists.php?artist=1&id=2" | pvreplace -payload payloads.txt -without-encode
http://testphp.vulnweb.com/artists.php?artist="><script>confirm(1)</script>&id="><script>confirm(1)</script>
http://testphp.vulnweb.com/artists.php?artist="<image/src/onerror=confirm(1)>&id="<image/src/onerror=confirm(1)>
```

### Comparsion
```
## qsreplace
▶ echo "http://fakedomain.com/fakefile.jsp;jsessionid=2ed4262dbe69850d25bc7c6424ba59db?hardwareid=14&tarifid=9998" | qsreplace "FUZZ"
http://fakedomain.com/fakefile.jsp;jsessionid=2ed4262dbe69850d25bc7c6424ba59db?hardwareid=FUZZ&tarifid=FUZZ

## pvreplace
▶ echo "http://fakedomain.com/fakefile.jsp;jsessionid=2ed4262dbe69850d25bc7c6424ba59db?hardwareid=14&tarifid=9998" | pvreplace -payload "FUZZ" -part param-value
http://fakedomain.com/fakefile.jsp;jsessionid=FUZZ&tarifid=FUZZ

▶ echo "http://fakedomain.com/fakefile.jsp;jsessionid=2ed4262dbe69850d25bc7c6424ba59db?hardwareid=14&tarifid=9998" | pvreplace -payload "FUZZ" -part param-value -mode single
http://fakedomain.com/fakefile.jsp;jsessionid=FUZZ&tarifid=9998
http://fakedomain.com/fakefile.jsp;jsessionid=2ed4262dbe69850d25bc7c6424ba59db?hardwareid=14&tarifid=FUZZ
```

## Credit
This tool was inspired by @R0X4R's [bhedak](https://github.com/R0X4R/bhedak) tool. Thanks to them for the great idea!
