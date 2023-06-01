import re
import urllib.parse as ul
from sys import stdin, stdout, argv, exit

# Check if the "-without-encode" argument is provided
encode = True
if len(argv) > 2 and argv[2] == "-without-encode":
    encode = False

# Check if help argument is provided
if len(argv) > 1 and (argv[1] == "-h" or argv[1] == "--help"):
    print("Usage: python3 pvreplace.py [string] [-without-encode]")
    print("Arguments:")
    print("  [string]           The string to be encoded and replaced in URLs")
    print("  -without-encode    Optional argument to disable URL encoding (default: enabled)")
    exit(0)

try:
    encoded = ul.quote(str(argv[1]), safe='') if encode else str(argv[1])
except IndexError:
    encoded = ul.quote("FUZZ", safe='')

try:
    unique_urls = set()  # Set to store unique URLs

    for url in stdin.readlines():
        domain = str(url.strip())
        modified_url = re.sub(r"=[^?\|&]*", '=' + str(encoded), str(domain))
        unique_urls.add(modified_url)  # Add modified URL to the set

    # Print unique URLs
    for url in unique_urls:
        stdout.write(url + '\n')
except KeyboardInterrupt:
    exit(0)
except:
    exit(127)