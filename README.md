# wikipush

[![Go Report Card](https://goreportcard.com/badge/github.com/thomasheller/wikipush)](https://goreportcard.com/report/github.com/thomasheller/wikipush)

Bulk upload files into a MediaWiki wiki using the MediaWiki API.

Uses [go-mediawiki](https://github.com/sadbox/mediawiki/) by
[sadbox](https://github.com/sadbox/).

## Install

Prerequisites:
  - Git
  - Golang

If you haven't already, install Git and Golang on your system. On
Ubuntu/Debian this would be:

```
sudo apt-get install git golang
```

Then set up Go:
  - Create a directory for your `$GOPATH`, for example `~/gocode`
  - Set the `$GOPATH` environment variable accordingly: `export GOPATH=~/gocode`
  - Add the `bin` directory to your `$PATH`, for example: ` export PATH=$PATH:~/gocode/bin`

Now you can install wikipush using `go get`:

```
go get github.com/thomasheller/wikipush
```

## Usage

Run `wikipush` in a directory with a lot of `*.txt` files.
wikipush will tell you how many files there are.

If everything appears to be correct, start the bulk upload by
running `wikipush -run -url http://.../api.php`.

wikipush will ask you for your MediaWiki username and password and then begin
uploading the files.

Rules:

1. If a page doesn't exist, it will be created and the input file will be moved
   to the `done` sub-directory.
1. If a page already exists, with **different** content, wikipush will do
   nothing and move the input file to the `dupes` sub-directory.
1. If a page already exists, with the **same** content, wikipush will do nothing
   and move the input file to the `skipped` sub-directory.

If wikipush completed succesfully, it will print a few statistics about
how many files landed in which folder and how many errors occured during
processing.

## More options

### Filename extension

If your input files have a filename extension different from `.txt`, you can
tell wikipush to recognize a different extension using the `-ext` flag,
for example `-ext .foo` (include the dot, unless there really is none).

### Throttling

wikipush will wait half a second between each upload attempt. If you'd like to
increase/decrease the duration, use the `-pause` flag. You can specify values 
like `-pause 0`, `-pause 100ms`, `-pause 2s`, `-pause 5m` etc.

### Revision history

Uploads will show up as `Bulk upload by wikipush` in the page revision log.
You can change this message to something else using the `-summary` flag,
for instance `-summary "I love MediaWiki!"`.
