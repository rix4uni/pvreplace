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
Usage: pvreplace [string] [-without-encode]
Arguments:
  [string]           The string to be encoded and replaced in URLs
  -without-encode    Optional argument to disable URL encoding (default: enabled)
```

### Example input file:
```
▶ cat urls.txt
https://example.com/path?one=1&two=2
https://example.com/path?two=2&one=1
https://example.com/pathtwo?two=2&one=1
https://example.net/a/path?two=2&one=1
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
