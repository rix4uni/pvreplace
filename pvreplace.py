import re
import urllib.parse as ul
from sys import stdin, stdout, argv, exit

# Check if the "-without-encode" argument is provided
encode = True
if "-without-encode" in argv:
    encode = False

# Check if help argument is provided
if len(argv) > 1 and (argv[1] == "-h" or argv[1] == "--help"):
    print("Usage: python3 pvreplace.py [string] [-without-encode]")
    print("Arguments:")
    print("  [string]           The string(s) to be replaced in URLs")
    print("  -without-encode    Optional argument to disable URL encoding (default: enabled)")
    exit(0)

try:
    strings_arg = argv[1] if len(argv) > 1 else "FUZZ"
    strings = [s.strip() for s in strings_arg.split(",")]

    encoded_strings = []
    for string in strings:
        encoded_strings.append(ul.quote(str(string), safe='') if encode else str(string))

    for url in stdin.readlines():
        domain = str(url.strip())
        for encoded in encoded_strings:
            modified_url = re.sub(r"=[^?\|&]*", '=' + str(encoded), domain)
            stdout.write(modified_url + '\n')

except KeyboardInterrupt:
    exit(0)
except Exception as e:
    print("Error:", e)
    exit(127)
