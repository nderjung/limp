# limp

```
limp is a tool which LIMits standard input which has been Piped to it.

Usage:
  limp [OPTION]... [FILE]...

Examples:
  limp can be used to save the last 10 lines of a build output, for
  example:

    $ kraft build |& limp -n 10 -o build/error.log

Flags:
  -h, --help         help for limp
  -i, --in string    file to read or follow
  -n, --lines int    keep the last number of lines (default 10)
  -o, --out string   output file
  -v, --version      display version number
```

## Installation

**Ubuntu/Debian**
```bash
wget https://github.com/nderjung/limp/releases/download/v0.1.1/limp_0.1.1_linux_amd64.deb
dpkg -i ./limp_0.1.1_linux_amd64.deb
```

**Mac**

```bash
brew tap nderjung/homebrew-tap
brew install limp
```

or download the latest Darwin build from the [releases page](https://github.com/nderjung/limp/releases/download/v0.1.1/limp_0.1.1_darwin_amd64.tar.gz).

**Go tools**
Requires Go version 1.10 or higher.

```bash
go get github.com/nderjung/limp
```
*Note*: installing in this way you will not see a proper version when running `limp -v`.

**Docker**
```bash
docker pull ndrjng/limp:v0.1.1
```
