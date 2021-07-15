# ghinfo

## What this system does

This is a CLI app that fetches a list of GH public repositories. It returns a curated list of all repositories starting at a the specified repository ID and ending at the specified repository maximum ID..

## Steps to run:

- git clone the source
- set an enviroment variable named `GH_TOKEN` to a GH personal access token
- at the root of the project, run `go run main.go`