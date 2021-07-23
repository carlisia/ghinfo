# ghinfo

## What this system does

This is a CLI app that fetches a list of GH public repositories. It returns a curated list of all repositories starting at a the specified repository ID and ending at the specified repository maximum ID..

It also fetches the associated star count and license type per repository.

Running the app via the terminal will provide the prompts to print the available reports.

## Steps to run:

- git clone or download and unzip the source code
- cd into the root of the project
- set an enviroment variable named `GH_TOKEN` to a GH personal access token
- run `go mod tidy`
- run `go run main.go`
- follow the prompts

## Previews

### Stargazers report
[![asciicast](https://asciinema.org/a/426875.svg)](https://asciinema.org/a/426875?autoplay=1&preload=1)

### License type report
[![asciicast](https://asciinema.org/a/426879.svg)](https://asciinema.org/a/426879?autoplay=1&preload=1)