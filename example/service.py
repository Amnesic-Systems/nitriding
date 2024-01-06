#!/usr/bin/env python3

import urllib.request

nitriding_url = "http://127.0.0.1:8080/enclave/ready"


def fetch_addr():
    url = "https://raw.githubusercontent.com/Amnesic-Systems/nitriding/master/README.md"
    with urllib.request.urlopen(url) as f:
        print(
            "[py] Successfully fetched %d bytes of README.md from within enclave!"
            % len(f.read(100))
        )


if __name__ == "__main__":
    fetch_addr()
    print("[py] Made Web request to the outside world.")
