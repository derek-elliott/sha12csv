# sha12csv

## Overview

This is a small utility used to write the sha1 sum of all files in a directory tree to a CSV.

## Install
Download one of the binary files on the release page.  You can also `go get` and `go install` yourself.
```
go get github.com/derek-elliott/sha12csv
```
There's a Makefile included that will help you build and install.

## Useage
```
sha1sum .
```
This will output a list of each file in the directory(including sub-directories) with their sha1 sums.  You might run into the open file limit.  If you do, you can increase it temporarly with `ulimit -n 90000`, or perminantly by editing `/etc/security/limits.conf`.

## License

MIT.
