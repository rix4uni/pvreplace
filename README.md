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
Usage: python3 pvreplace.py [string] [-without-encode] [-flags]
Arguments:
  [string]           The string(s) to be replaced in URLs (default: FUZZ)
  -without-encode    Optional argument to disable URL encoding (default: enabled)
  -flags             Additional flags for URL modification patterns
  -param-value       Replacing parameter values
  -param-name        Replacing parameter names
  -path-suffix       Adding a suffix to the path
  -path-param        Modifying the path and adding a parameter
  -ext-filename      Replacing the filenames
  -without-encode    Prints URLs without encode the payload
  -single-replace    Replacing one By one
  -v, --version      Prints current version
  -h, --help         Prints Help
```

### Example input file:
```
▶ cat urls.txt
https://example.com/path?one=1&two=2
https://example.com/path?two=2&one=1
https://example.com/pathtwo?two=2&one=1
https://example.net/a/path?two=2&one=1
http://testphp.vulnweb.com/artists.php?artist=1&id=2
```

### If you not passed any `payload` by default replace with `FUZZ`
```
▶ cat urls.txt | pvreplace
https://example.net/a/path?two=FUZZ&one=FUZZ
https://example.com/pathtwo?two=FUZZ&one=FUZZ
https://example.com/path?two=FUZZ&one=FUZZ
https://example.com/path?one=FUZZ&two=FUZZ
```

### Replace query string with custom payloads
```
▶ cat urls.txt | pvreplace "<script>alert(1)</script>"
https://example.com/path?two=%3Cscript%3Ealert%281%29%3C%2Fscript%3E&one=%3Cscript%3Ealert%281%29%3C%2Fscript%3E
https://example.com/path?one=%3Cscript%3Ealert%281%29%3C%2Fscript%3E&two=%3Cscript%3Ealert%281%29%3C%2Fscript%3E
https://example.net/a/path?two=%3Cscript%3Ealert%281%29%3C%2Fscript%3E&one=%3Cscript%3Ealert%281%29%3C%2Fscript%3E
https://example.com/pathtwo?two=%3Cscript%3Ealert%281%29%3C%2Fscript%3E&one=%3Cscript%3Ealert%281%29%3C%2Fscript%3E
```

### Replace query string with custom payloads and without encode the payload
```
▶ cat urls.txt | pvreplace "<script>alert(1)</script>" -without-encode
https://example.com/pathtwo?two=<script>alert(1)</script>&one=<script>alert(1)</script>
https://example.net/a/path?two=<script>alert(1)</script>&one=<script>alert(1)</script>
https://example.com/path?one=<script>alert(1)</script>&two=<script>alert(1)</script>
https://example.com/path?two=<script>alert(1)</script>&one=<script>alert(1)</script>
```

### Replace multiple payloads separated by commas
```
# urls.txt file content
http://testphp.vulnweb.com/listproducts.php?artist=dfgsdftgerer
http://testphp.vulnweb.com/listproducts.php?artist=dfgsdftgerer&asdf=yry4tytr&cat=fgfgh


# Command
cat urls.txt | pvreplace '"><script>confirm(1)</script>, "<image/src/onerror=confirm(1)>' -without-encode
http://testphp.vulnweb.com/listproducts.php?artist="><script>confirm(1)</script>
http://testphp.vulnweb.com/listproducts.php?artist="<image/src/onerror=confirm(1)>
http://testphp.vulnweb.com/listproducts.php?artist="><script>confirm(1)</script>&asdf="><script>confirm(1)</script>&cat="><script>confirm(1)</script>
http://testphp.vulnweb.com/listproducts.php?artist="<image/src/onerror=confirm(1)>&asdf="<image/src/onerror=confirm(1)>&cat="<image/src/onerror=confirm(1)>
```

### Comparsion
```
## qsreplace
echo "http://fakedomain.com/fakefile.jsp;jsessionid=2ed4262dbe69850d25bc7c6424ba59db?hardwareid=14&tarifid=9998" | qsreplace "FUZZ"
http://fakedomain.com/fakefile.jsp;jsessionid=2ed4262dbe69850d25bc7c6424ba59db?hardwareid=FUZZ&tarifid=FUZZ

## pvreplace
echo "http://fakedomain.com/fakefile.jsp;jsessionid=2ed4262dbe69850d25bc7c6424ba59db?hardwareid=14&tarifid=9998" | pvreplace "FUZZ"
http://fakedomain.com/fakefile.jsp;jsessionid=FUZZ?hardwareid=FUZZ&tarifid=FUZZ
```

## Credit
This tool was inspired by @R0X4R's [bhedak](https://github.com/R0X4R/bhedak) tool. Thanks to them for the great idea!
