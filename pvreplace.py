import re
import urllib.parse as ul
from sys import stdin, stdout, argv, exit

def print_help():
    print("Usage: python3 pvreplace.py [strings] [-without-encode] [-part] [-type] [-mode] [-payload [strings or filepath]]")
    print("\npositional arguments:")
    print("  strings             The string(s) to be replaced in URLs (default: FUZZ)")
    print("\noptions:")
    print("  -without-encode    Optional argument to disable URL encoding (default: enabled)")
    print("  -part              Specify which part of the URL to modify Options: param-value, param-name, path-suffix, path-param, ext-filename (default: param-value)")
    print("  -type              Specify the type of modification Options: replace, prefix, postfix (default: replace)")
    print("  -mode              Specify the mode of replacement Options: multiple, single (default: multiple)")
    print("  -payload           Specify payload(s) directly or from a file")
    print("  -v, --version      Prints current version")
    print("  -h, --help         Prints Help")
    exit(0)

def parse_args():
    encode = True
    part = "param-value"
    replace_type = "replace"
    mode = "multiple"
    strings_arg = "FUZZ"

    if "-h" in argv or "--help" in argv:
        print_help()

    if "-v" in argv or "--version" in argv:
        print("pvreplace version: v0.0.2")
        exit(0)
    
    if "-without-encode" in argv:
        encode = False
    
    payload_flag = False
    payload_file = False

    for arg in argv[1:]:
        if arg == "-without-encode":
            encode = False
        elif arg == "-part":
            part_index = argv.index(arg) + 1
            if part_index < len(argv):
                part = argv[part_index]
        elif arg == "-type":
            type_index = argv.index(arg) + 1
            if type_index < len(argv):
                replace_type = argv[type_index]
        elif arg == "-mode":
            mode_index = argv.index(arg) + 1
            if mode_index < len(argv):
                mode = argv[mode_index]
        elif arg == "-payload":
            payload_flag = True
            payload_index = argv.index(arg) + 1
            if payload_index < len(argv):
                strings_arg = argv[payload_index]
                if strings_arg.endswith(".txt"):
                    payload_file = True
        elif arg not in ["-without-encode", "-part", part, "-type", replace_type, "-mode", mode, "-payload"]:
            strings_arg = arg

    if payload_flag and payload_file:
        with open(strings_arg, 'r') as file:
            strings_arg = file.read().strip().replace("\n", ",")

    return strings_arg, encode, part, replace_type, mode

def modify_url(url, encoded_strings, part, replace_type, mode):
    domain = url.strip()
    modified_urls = []

    for encoded in encoded_strings:
        if part == "param-value":
            if mode == "single":
                params = re.findall(r"=[^&]*", domain)
                for param in params:
                    if replace_type == "replace":
                        modified_urls.append(domain.replace(param, f"={encoded}", 1))
                    elif replace_type == "prefix":
                        modified_urls.append(domain.replace(param, f"={encoded}{param[1:]}", 1))
                    elif replace_type == "postfix":
                        modified_urls.append(domain.replace(param, f"={param[1:]}{encoded}", 1))
            else:
                if replace_type == "replace":
                    modified_urls.append(re.sub(r"=([^&]*)", f"={encoded}", domain))
                elif replace_type == "prefix":
                    modified_urls.append(re.sub(r"=([^&]*)", f"={encoded}\\1", domain))
                elif replace_type == "postfix":
                    modified_urls.append(re.sub(r"=([^&]*)", f"=\\1{encoded}", domain))
        elif part == "param-name":
            if mode == "single":
                params = re.findall(r"([?&])([^&=]+)=", domain)
                for param, name in params:
                    if replace_type == "replace":
                        modified_urls.append(domain.replace(f"{param}{name}=", f"{param}{encoded}=", 1))
                    elif replace_type == "prefix":
                        modified_urls.append(domain.replace(f"{param}{name}=", f"{param}{encoded}{name}=", 1))
                    elif replace_type == "postfix":
                        modified_urls.append(domain.replace(f"{param}{name}=", f"{param}{name}{encoded}=", 1))
            else:
                def replace_param_name(match):
                    if replace_type == "replace":
                        return match.group(1) + encoded + "="
                    elif replace_type == "prefix":
                        return match.group(1) + encoded + match.group(2) + "="
                    elif replace_type == "postfix":
                        return match.group(1) + match.group(2) + encoded + "="
                modified_urls.append(re.sub(r"([?&])([^&=]+)=", replace_param_name, domain))
        elif part == "path-suffix":
            if "?" in domain:
                base, params = domain.split("?", 1)
                if replace_type == "replace":
                    modified_urls.append(f"{base}/{encoded}?{params}")
                elif replace_type == "prefix":
                    modified_urls.append(f"{base}/{encoded}{base.split('/')[-1]}?{params}")
                elif replace_type == "postfix":
                    modified_urls.append(f"{base}/{base.split('/')[-1]}{encoded}?{params}")
            else:
                if replace_type == "replace":
                    modified_urls.append(domain + '/' + encoded)
                elif replace_type == "prefix":
                    modified_urls.append(domain + '/' + encoded + domain.split('/')[-1])
                elif replace_type == "postfix":
                    modified_urls.append(domain + '/' + domain.split('/')[-1] + encoded)
        elif part == "path-param":
            if "?" in domain:
                base, params = domain.split("?", 1)
                if replace_type == "replace":
                    modified_urls.append(f"{base}/{encoded}?{params}")
                    modified_urls.append(f"{base}?{encoded}&{params}")
                elif replace_type == "prefix":
                    modified_urls.append(f"{base}/{encoded}{base.split('/')[-1]}?{params}")
                    modified_urls.append(f"{base}?{encoded}{base.split('/')[-1]}&{params}")
                elif replace_type == "postfix":
                    modified_urls.append(f"{base}/{base.split('/')[-1]}{encoded}?{params}")
                    modified_urls.append(f"{base}?{base.split('/')[-1]}{encoded}&{params}")
            else:
                if replace_type == "replace":
                    modified_urls.append(domain + '/' + encoded)
                    modified_urls.append(domain + '?' + encoded)
                elif replace_type == "prefix":
                    modified_urls.append(domain + '/' + encoded + domain.split('/')[-1])
                    modified_urls.append(domain + '?' + encoded + domain.split('/')[-1])
                elif replace_type == "postfix":
                    modified_urls.append(domain + '/' + domain.split('/')[-1] + encoded)
                    modified_urls.append(domain + '?' + domain.split('/')[-1] + encoded)
        elif part == "ext-filename":
            if replace_type == "replace":
                modified_urls.append(re.sub(r"/([^/]+)\.(php|aspx|asp|jsp|jspx|xml)", f"/{encoded}.\\2", domain))
            elif replace_type == "prefix":
                modified_urls.append(re.sub(r"/([^/]+)\.(php|aspx|asp|jsp|jspx|xml)", f"/{encoded}\\1.\\2", domain))
            elif replace_type == "postfix":
                modified_urls.append(re.sub(r"/([^/]+)\.(php|aspx|asp|jsp|jspx|xml)", f"/\\1{encoded}.\\2", domain))

    return modified_urls

def main():
    strings_arg, encode, part, replace_type, mode = parse_args()
    strings = [s.strip() for s in strings_arg.split(",")]

    encoded_strings = [ul.quote(str(string), safe='') if encode else str(string) for string in strings]

    try:
        for url in stdin.readlines():
            modified_urls = modify_url(url, encoded_strings, part, replace_type, mode)
            for modified_url in modified_urls:
                stdout.write(modified_url + '\n')

    except KeyboardInterrupt:
        exit(0)
    except Exception as e:
        print("Error:", e)
        exit(127)

if __name__ == "__main__":
    main()
