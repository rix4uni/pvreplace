import re
import urllib.parse as ul
from sys import stdin, stdout, argv, exit

def print_help():
    print("Usage: python3 pvreplace.py [string] [-without-encode] [-flags]")
    print("Arguments:")
    print("  [string]           The string(s) to be replaced in URLs (default: FUZZ)")
    print("  -without-encode    Optional argument to disable URL encoding (default: enabled)")
    print("  -flags             Additional flags for URL modification patterns")
    print("  -param-value       Replacing parameter values")
    print("  -param-name        Replacing parameter names")
    print("  -path-suffix       Adding a suffix to the path")
    print("  -path-param        Modifying the path and adding a parameter")
    print("  -ext-filename      Replacing the filenames")
    print("  -without-encode    Prints URLs without encode the payload")
    print("  -single-replace    Replacing one By one")
    print("  -v, --version      Prints current version")
    print("  -h, --help         Prints Help")
    exit(0)

def parse_args():
    encode = True
    flags = []
    strings_arg = "FUZZ"

    if "-h" in argv or "--help" in argv:
        print_help()

    if "-v" in argv or "--version" in argv:
        print("pvreplace version: v0.0.2")
        exit(0)
    
    if "-without-encode" in argv:
        encode = False
    
    for arg in argv[1:]:
        if arg not in ["-without-encode", "-param-value", "-param-name", "-path-suffix", "-path-param", "-ext-filename", "-single-replace"]:
            strings_arg = arg
        if arg in ["-param-value", "-param-name", "-path-suffix", "-path-param", "-ext-filename", "-single-replace"]:
            flags.append(arg)

    return strings_arg, encode, flags

def modify_url(url, encoded_strings, flags):
    domain = url.strip()
    modified_urls = []

    for encoded in encoded_strings:
        if "-param-value" in flags:
            if "-single-replace" in flags:
                params = re.findall(r"=[^&]*", domain)
                for param in params:
                    modified_urls.append(domain.replace(param, f"={encoded}", 1))
            else:
                modified_urls.append(re.sub(r"=([^&]*)", f"={encoded}", domain))
        if "-param-name" in flags:
            if "-single-replace" in flags:
                params = re.findall(r"([?&])([^&=]+)=", domain)
                for param, name in params:
                    modified_urls.append(domain.replace(f"{param}{name}=", f"{param}{encoded}="))
            else:
                def replace_param_name(match):
                    return match.group(1) + encoded + "="
                modified_urls.append(re.sub(r"([?&])([^&=]+)=", replace_param_name, domain))
        if "-path-suffix" in flags:
            if "?" in domain:
                base, params = domain.split("?", 1)
                modified_urls.append(f"{base}/{encoded}?{params}")
            else:
                modified_urls.append(domain + '/' + encoded)
        if "-path-param" in flags:
            if "?" in domain:
                base, params = domain.split("?", 1)
                modified_urls.append(f"{base}/{encoded}?{params}")
                modified_urls.append(f"{base}?{encoded}&{params}")
            else:
                modified_urls.append(domain + '/' + encoded)
                modified_urls.append(domain + '?' + encoded)
        if "-ext-filename" in flags:
            modified_urls.append(re.sub(r"/([^/]+)\.(php|aspx|asp|jsp|jspx|xml)", f"/{encoded}.\\2", domain))

    return modified_urls

def main():
    strings_arg, encode, flags = parse_args()
    strings = [s.strip() for s in strings_arg.split(",")]

    encoded_strings = [ul.quote(str(string), safe='') if encode else str(string) for string in strings]

    try:
        for url in stdin.readlines():
            modified_urls = modify_url(url, encoded_strings, flags)
            for modified_url in modified_urls:
                stdout.write(modified_url + '\n')

    except KeyboardInterrupt:
        exit(0)
    except Exception as e:
        print("Error:", e)
        exit(127)

if __name__ == "__main__":
    main()
