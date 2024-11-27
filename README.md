## GoLinkFinder

A minimal JS endpoint extractor, To extract endpoints in both HTML source and embedded javascript files. Useful for bug hunters, red teamers, infosec ninjas.

## Installation
```
go install github.com/rix4uni/GoLinkFinder@latest
```

## Download prebuilt binaries
```
wget https://github.com/rix4uni/GoLinkFinder/releases/download/v0.0.1/GoLinkFinder-linux-amd64-0.0.1.tgz
tar -xvzf GoLinkFinder-linux-amd64-0.0.1.tgz
rm -rf GoLinkFinder-linux-amd64-0.0.1.tgz
mv GoLinkFinder ~/go/bin/GoLinkFinder
```
Or download [binary release](https://github.com/rix4uni/GoLinkFinder/releases) for your platform.

## Compile from source
```
git clone --depth 1 github.com/rix4uni/GoLinkFinder.git
cd GoLinkFinder; go install
```

## Usage
```
Usage of GoLinkFinder:
      --H string          Set custom User-Agent. (default "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
      --complete-url      Add the domain to relative URLs.
  -c, --concurrency int   Concurrency level. (default 10)
      --delay int         Delay between requests in milliseconds.
  -d, --domain string     Input a URL.
  -l, --list string       Input file containing a list of live subdomains to process.
      --only-complete     Show only complete URLs starting with http:// or https://.
  -o, --output string     File to write output results.
      --silent            silent mode.
      --timeout int       HTTP timeout in seconds. (default 10)
      --verbose           Enable verbose mode.
      --version           Print the version of the tool and exit.
```

## Examples
Single Target:
```
▶ echo "http://testphp.vulnweb.com" | GoLinkFinder -silent
```

Multiple Targets:
```
cat targets.txt | GoLinkFinder -silent
```