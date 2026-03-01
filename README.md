# wget

A simple implementation of the `wget` tool in Go.

This project supports:

-   Downloading a single file
-   Custom output filename (`-O`)
-   Custom output directory (`-P`)
-   Download speed limiting (`--rate-limit`)
-   Background mode logging (`-B`)
-   Batch download from file (`-i`)
-   Recursive website mirroring (`--mirror`)
-   Progress bar with ETA

------------------------------------------------------------------------

## Build

    go build -o wget

------------------------------------------------------------------------

## Usage

### Download a file

    ./wget https://example.com/file.zip

------------------------------------------------------------------------

### Save with custom filename

    ./wget -O=myfile.zip https://example.com/file.zip

------------------------------------------------------------------------

### Save to specific directory

    ./wget -P=downloads https://example.com/file.zip

------------------------------------------------------------------------

### Limit download speed

    ./wget --rate-limit=200k https://example.com/file.zip

Supported units: - `k` = kilobytes per second - `m` = megabytes per
second

------------------------------------------------------------------------

### Background mode (log to file)

    ./wget -B https://example.com/file.zip

Output will be written to `wget-log`.

------------------------------------------------------------------------

### Batch download (multiple URLs)

Create a file `links.txt`:

    https://example.com/file1.zip
    https://example.com/file2.zip

Then run:

    ./wget -i=links.txt

------------------------------------------------------------------------

### Mirror a website

    ./wget --mirror https://example.com

This will create a folder with the domain name and download the website
structure inside it.

------------------------------------------------------------------------

## Notes

-   Shows download progress and ETA when content length is available.
-   Automatically handles unknown file sizes.
-   Uses a custom User-Agent header.

------------------------------------------------------------------------

## Author

Project developed as part of a Go programming exercise, by Andrikopoulos Andreas-Rafail, Vasileios Parikoglou, Athanasios Ziagakis and Athanasios Diridis.
