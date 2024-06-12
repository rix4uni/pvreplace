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
  -single-replace    Replacing one By one parameter value
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
▶ cat urls.txt | pvreplace -param-value
https://example.com/path?one=FUZZ&two=FUZZ
https://example.com/path?two=FUZZ&one=FUZZ
https://example.com/pathtwo?two=FUZZ&one=FUZZ
https://example.net/a/path?two=FUZZ&one=FUZZ
http://testphp.vulnweb.com/artists.php?artist=FUZZ&id=FUZZ
```

### Replacing one By one parameter value
```
▶ cat urls.txt | pvreplace -param-value -single-replace
https://example.com/path?one=FUZZ&two=2
https://example.com/path?one=1&two=FUZZ
https://example.com/path?two=FUZZ&one=1
https://example.com/path?two=2&one=FUZZ
https://example.com/pathtwo?two=FUZZ&one=1
https://example.com/pathtwo?two=2&one=FUZZ
https://example.net/a/path?two=FUZZ&one=1
https://example.net/a/path?two=2&one=FUZZ
http://testphp.vulnweb.com/artists.php?artist=FUZZ&id=2
http://testphp.vulnweb.com/artists.php?artist=1&id=FUZZ
```

### Replace query string with custom payloads
```
▶ cat urls.txt | pvreplace "<script>alert(1)</script>" -param-value
https://example.com/path?one=%3Cscript%3Ealert%281%29%3C%2Fscript%3E&two=%3Cscript%3Ealert%281%29%3C%2Fscript%3E
https://example.com/path?two=%3Cscript%3Ealert%281%29%3C%2Fscript%3E&one=%3Cscript%3Ealert%281%29%3C%2Fscript%3E
https://example.com/pathtwo?two=%3Cscript%3Ealert%281%29%3C%2Fscript%3E&one=%3Cscript%3Ealert%281%29%3C%2Fscript%3E
https://example.net/a/path?two=%3Cscript%3Ealert%281%29%3C%2Fscript%3E&one=%3Cscript%3Ealert%281%29%3C%2Fscript%3E
http://testphp.vulnweb.com/artists.php?artist=%3Cscript%3Ealert%281%29%3C%2Fscript%3E&id=%3Cscript%3Ealert%281%29%3C%2Fscript%3E
```

### Replace query string with custom payloads and without encode the payload
```
▶ cat urls.txt | pvreplace "<script>alert(1)</script>" -param-value -without-encode
https://example.com/path?one=<script>alert(1)</script>&two=<script>alert(1)</script>
https://example.com/path?two=<script>alert(1)</script>&one=<script>alert(1)</script>
https://example.com/pathtwo?two=<script>alert(1)</script>&one=<script>alert(1)</script>
https://example.net/a/path?two=<script>alert(1)</script>&one=<script>alert(1)</script>
http://testphp.vulnweb.com/artists.php?artist=<script>alert(1)</script>&id=<script>alert(1)</script>
```

### Replace multiple payloads separated by commas
```
# urls.txt file content
▶ cat urls.txt
https://example.com/path?one=1&two=2
https://example.com/path?two=2&one=1
https://example.com/pathtwo?two=2&one=1
https://example.net/a/path?two=2&one=1
http://testphp.vulnweb.com/artists.php?artist=1&id=2
http://testphp.vulnweb.com/listproducts.php?artist=dfgsdftgerer
http://testphp.vulnweb.com/listproducts.php?artist=dfgsdftgerer&asdf=yry4tytr&cat=fgfgh


# without encode
▶ cat urls.txt | pvreplace '"><script>confirm(1)</script>, "<image/src/onerror=confirm(1)>' -param-value -without-encode
https://example.com/path?one="><script>confirm(1)</script>&two="><script>confirm(1)</script>
https://example.com/path?one="<image/src/onerror=confirm(1)>&two="<image/src/onerror=confirm(1)>
https://example.com/path?two="><script>confirm(1)</script>&one="><script>confirm(1)</script>
https://example.com/path?two="<image/src/onerror=confirm(1)>&one="<image/src/onerror=confirm(1)>
https://example.com/pathtwo?two="><script>confirm(1)</script>&one="><script>confirm(1)</script>
https://example.com/pathtwo?two="<image/src/onerror=confirm(1)>&one="<image/src/onerror=confirm(1)>
https://example.net/a/path?two="><script>confirm(1)</script>&one="><script>confirm(1)</script>
https://example.net/a/path?two="<image/src/onerror=confirm(1)>&one="<image/src/onerror=confirm(1)>
http://testphp.vulnweb.com/artists.php?artist="><script>confirm(1)</script>&id="><script>confirm(1)</script>
http://testphp.vulnweb.com/artists.php?artist="<image/src/onerror=confirm(1)>&id="<image/src/onerror=confirm(1)>
http://testphp.vulnweb.com/listproducts.php?artist="><script>confirm(1)</script>
http://testphp.vulnweb.com/listproducts.php?artist="<image/src/onerror=confirm(1)>
http://testphp.vulnweb.com/listproducts.php?artist="><script>confirm(1)</script>&asdf="><script>confirm(1)</script>&cat="><script>confirm(1)</script>
http://testphp.vulnweb.com/listproducts.php?artist="<image/src/onerror=confirm(1)>&asdf="<image/src/onerror=confirm(1)>&cat="<image/src/onerror=confirm(1)>

# without encode with one by one parameter value replace
▶ cat urls.txt | pvreplace '"><script>confirm(1)</script>, "<image/src/onerror=confirm(1)>' -param-value -without-encode -single-replace
https://example.com/path?one="><script>confirm(1)</script>&two=2
https://example.com/path?one=1&two="><script>confirm(1)</script>
https://example.com/path?one="<image/src/onerror=confirm(1)>&two=2
https://example.com/path?one=1&two="<image/src/onerror=confirm(1)>
https://example.com/path?two="><script>confirm(1)</script>&one=1
https://example.com/path?two=2&one="><script>confirm(1)</script>
https://example.com/path?two="<image/src/onerror=confirm(1)>&one=1
https://example.com/path?two=2&one="<image/src/onerror=confirm(1)>
https://example.com/pathtwo?two="><script>confirm(1)</script>&one=1
https://example.com/pathtwo?two=2&one="><script>confirm(1)</script>
https://example.com/pathtwo?two="<image/src/onerror=confirm(1)>&one=1
https://example.com/pathtwo?two=2&one="<image/src/onerror=confirm(1)>
https://example.net/a/path?two="><script>confirm(1)</script>&one=1
https://example.net/a/path?two=2&one="><script>confirm(1)</script>
https://example.net/a/path?two="<image/src/onerror=confirm(1)>&one=1
https://example.net/a/path?two=2&one="<image/src/onerror=confirm(1)>
http://testphp.vulnweb.com/artists.php?artist="><script>confirm(1)</script>&id=2
http://testphp.vulnweb.com/artists.php?artist=1&id="><script>confirm(1)</script>
http://testphp.vulnweb.com/artists.php?artist="<image/src/onerror=confirm(1)>&id=2
http://testphp.vulnweb.com/artists.php?artist=1&id="<image/src/onerror=confirm(1)>
http://testphp.vulnweb.com/listproducts.php?artist="><script>confirm(1)</script>
http://testphp.vulnweb.com/listproducts.php?artist="<image/src/onerror=confirm(1)>
http://testphp.vulnweb.com/listproducts.php?artist="><script>confirm(1)</script>&asdf=yry4tytr&cat=fgfgh
http://testphp.vulnweb.com/listproducts.php?artist=dfgsdftgerer&asdf="><script>confirm(1)</script>&cat=fgfgh
http://testphp.vulnweb.com/listproducts.php?artist=dfgsdftgerer&asdf=yry4tytr&cat="><script>confirm(1)</script>
http://testphp.vulnweb.com/listproducts.php?artist="<image/src/onerror=confirm(1)>&asdf=yry4tytr&cat=fgfgh
http://testphp.vulnweb.com/listproducts.php?artist=dfgsdftgerer&asdf="<image/src/onerror=confirm(1)>&cat=fgfgh
http://testphp.vulnweb.com/listproducts.php?artist=dfgsdftgerer&asdf=yry4tytr&cat="<image/src/onerror=confirm(1)>
```

### Other flags
```
# Test replacing all parameter names with a specific string
echo "http://testphp.vulnweb.com/artists.php?artist=1&id=2" | python3 pvreplace -param-name "456"
http://testphp.vulnweb.com/artists.php?456=1&456=2

echo "http://testphp.vulnweb.com/artists.php?artist=1&id=2" | python3 pvreplace -param-name "456" -single-replace
# Expected output: http://testphp.vulnweb.com/artists.php?456=1&id=2
# Expected output: http://testphp.vulnweb.com/artists.php?artist=1&456=2

# Test replacing parameter values
echo "http://testphp.vulnweb.com/artists.php?artist=1&id=2" | python3 pvreplace -param-value
http://testphp.vulnweb.com/artists.php?artist=456&id=456

echo "http://testphp.vulnweb.com/artists.php?artist=1&id=2" | python3 pvreplace -param-value -single-replace
# Expected output: http://testphp.vulnweb.com/artists.php?artist=FUZZ&id=2
# Expected output: http://testphp.vulnweb.com/artists.php?artist=1&id=FUZZ

# Test adding a suffix to the path
echo "http://testphp.vulnweb.com/artists.php?artist=1&id=2" | python3 pvreplace -path-suffix
# Expected output: http://testphp.vulnweb.com/artists.php/FUZZ?artist=1&id=2

# Test modifying the path and adding a parameter
echo "http://testphp.vulnweb.com/artists.php?artist=1&id=2" | python3 pvreplace -path-param
# Expected output: http://testphp.vulnweb.com/artists.php/FUZZ?artist=1&id=2
# Another possible output: http://testphp.vulnweb.com/artists.php?FUZZ&artist=1&id=2

# Test replacing the filename with FUZZ
echo "http://testphp.vulnweb.com/artists.php?artist=1&id=2" | python3 pvreplace -ext-filename
# Expected output: http://testphp.vulnweb.com/FUZZ.php?artist=1&id=2
```

### Comparsion
```
## qsreplace
▶ echo "http://fakedomain.com/fakefile.jsp;jsessionid=2ed4262dbe69850d25bc7c6424ba59db?hardwareid=14&tarifid=9998" | qsreplace "FUZZ"
http://fakedomain.com/fakefile.jsp;jsessionid=2ed4262dbe69850d25bc7c6424ba59db?hardwareid=FUZZ&tarifid=FUZZ

## pvreplace
▶ echo "http://fakedomain.com/fakefile.jsp;jsessionid=2ed4262dbe69850d25bc7c6424ba59db?hardwareid=14&tarifid=9998" | pvreplace "FUZZ" -param-value
http://fakedomain.com/fakefile.jsp;jsessionid=FUZZ&tarifid=FUZZ

▶ echo "http://fakedomain.com/fakefile.jsp;jsessionid=2ed4262dbe69850d25bc7c6424ba59db?hardwareid=14&tarifid=9998" | pvreplace "FUZZ" -param-value -single-replace
http://fakedomain.com/fakefile.jsp;jsessionid=FUZZ&tarifid=9998
http://fakedomain.com/fakefile.jsp;jsessionid=2ed4262dbe69850d25bc7c6424ba59db?hardwareid=14&tarifid=FUZZ
```

## Credit
This tool was inspired by @R0X4R's [bhedak](https://github.com/R0X4R/bhedak) tool. Thanks to them for the great idea!
