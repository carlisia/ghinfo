# ghinfo

## What this system does

This is a CLI app that fetches a list of GH repositories. It returns a list of the first 100 repositories, ordered by ID number.

## Steps to run:

- git clone the source
- set an enviroment variable named `GH_TOKEN` to a GH personal access token
- at the root of the project, run `go run main.go`