# namzd

### Quickly find files by name or extension.

#### Downloads

[Download releases](https://github.com/bengarrett/namzd/releases) for 
[Windows](https://github.com/bengarrett/namzd/releases/latest/download/namzd_windows.zip), 
[Apple](https://github.com/bengarrett/namzd/releases/latest/download/namzd_apple_silicon.tgz) and 
[Linux](https://github.com/bengarrett/namzd/releases/latest/download/namzd_linux.tgz).

The download is a compressed binary that is standalone terminal application. 
Windows users can use File Explorer to decompress it.

```
# replace 'foo' with the remainder of the filename
$ tar zxf namzd_foo.tgz

# after decompression, to confirm the download and version
$ namzd -V
```

Before use, macOS users will need to delete the 'quarantine' extended attribute that is applied to all 
program downloads that are not notarized by Apple for a fee.

```
$ xattr -d com.apple.quarantine namzd
```

#### Homebrew

macOS and Linux users can install via Homebrew:

```bash
brew tap bengarrett/namzd https://github.com/bengarrett/namzd
brew install bengarrett/namzd/namzd
```

Update to the latest version with:

```bash
brew upgrade bengarrett/namzd/namzd
```

#### Usage

```
Usage: namzd <match> <paths> ... [flags]

Quickly find files by name or extension.

A <match> query is a filename, extension or pattern to match. These are case-insensitive
by default and should be quoted:

    'readme' matches README, Readme, readme, etc.
    'file.txt' matches file.txt, File.txt, file.TXT, etc.
    '*.txt' matches readme.txt, File.txt, DOC.TXT, etc.
    '*.tar*' matches files.tar.gz, FILE.tarball, files.tar, files.tar.xz, etc.
    '*.tar.??' matches files.tar.gz, files.tar.xz, etc.

Arguments:
  <match>        Filename, extension or pattern to match.
  <paths> ...    Paths to lookup.

Flags:
  -h, --help              Show context-sensitive help.
  -V, --version           Show the version information and exit.
  -c, --case-sensitive    Case sensitive match.
  -n, --count             Count the number of matches.
  -m, --last-modified     Show the last modified time of the match (yyyy-mm-dd).
  -o, --oldest            Show the oldest file match.
  -N, --newest            Show the newest file match.
  -d, --directory         Include directory matches.
  -f, --follow            Follow symbolic links.

Archives:
  Also search within archives for matching files. This will not open or decompress
  archives to read archives within archives.

  -a, --archive    Archive mode will also search within supported archives for matched
                   filenames.

Copier:
  Copy all matched files to a target directory. This option cannot be used with the
  archive options or the directory flag.

  -x, --destination=STRING    Destination directory path to copy matches.

Errors:
  -e, --errors    Errors mode displays any file and directory read or access errors.
  -p, --panic     Exits on any errors including file and directory read or access errors.
```

---

#### Example 1

```sh
$ namzd 'go.*' ~/github/namzd --count

1	/Users/ben/github/namzd/go.mod
2	/Users/ben/github/namzd/go.sum
```

- `namzd` is the application name.
- `'go.*'` is the pattern to match all files named 'go' using any file extension.
- `~/github/namzd` is the directory to lookup and search.
- `--count` is a flag to count the number of matches.

These are the two matching results with the match count and the absolute path to the file locations.

```
1	/Users/ben/github/namzd/go.mod
2	/Users/ben/github/namzd/go.sum
```

#### Example 2

This example matches both the names of files found in the directories and within zip and uncompressed tar archives.
It also shows the last modified date of the matches and the oldest match.

```sh
$ namzd 'file_id.diz' /home/ben/downloads --count --archive --last-modified --oldest

1 file_id.diz (1996-12-30) > /home/ben/downloads/stuff.tar
2 FILE_ID.DIZ (1993-01-19) > /home/ben/downloads/WOLFUPD.ZIP
3 FILE_ID.DIZ (1993-10-16) > /home/ben/downloads/YOLKFOLK.ZIP
Oldest found match:
2 FILE_ID.DIZ (1993-01-19) > /home/ben/downloads/WOLFUPD.ZIP
```

---

© 2024-2026 Ben Garrett - GPL-3.0 license
