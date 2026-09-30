# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

Two Go command-line connectors that pull updated items from a source repository, build a bag directory (payload files, `native-payload.json`, title and description files, `manifest-md5.txt`) on a scratch filesystem, and submit it to the APTrust submission service:

- `dataverse-connector`: pulls from Dataverse and submits to the `LibraData` collection with bag prefix `DataVerse`.
- `dspace-connector`: pulls from DSpace and submits to the `LibraOpen` collection with bag prefix `LibraOpen`.

Each connector is its own Go module (`package main`) with its own `go.mod`. There are no tests in this repo.

## Shared code via symlinks (important)

`service-common/*.go` is **not** a Go package. Each connector's Makefile `common` target symlinks those files into the connector directory, and they compile as part of that connector's `package main`. The root `.gitignore` ignores the symlinked copies.

- Always build with `make` (or run `make common` first). A plain `go build` or `go vet` in a connector directory fails until the symlinks exist.
- Edit shared code only in `service-common/`, never through the symlinks.
- Adding a new file to `service-common/` means adding it to the `common` target in **both** Makefiles **and** to the ignore list in `.gitignore`. `s3.go` is currently not symlinked, so neither connector compiles it.
- Shared code calls functions that each connector must define: `loadConfiguration()` / `ServiceConfig`, `authenticate()`, `getUpdatedIds()`, `processUpdatedIds()`, `createBagContents()`, and types like `ItemResponse`. The two connectors keep the same file layout (`auth.go`, `bag-contents.go`, `config.go`, `*-api.go`, `get-updated-ids.go`, `process-updated-ids.go`) so these match up.

## Commands

Run from inside `dataverse-connector/` or `dspace-connector/`:

```sh
make            # symlink common files + build bin/<name>.darwin (amd64, -race)
make linux      # symlink common files + build static bin/<name>.linux
make vet        # go vet
make check      # staticcheck (all,-S1002,-ST1003) + shadow vet
make dep        # go get -u, go mod tidy, go mod verify
make clean
```

Build the container from the repo root: `docker build -f package/Dockerfile .` (both binaries go into `/aptrust-submit-connectors/bin/`). CI (`pipeline/buildspec.yml`, AWS CodeBuild) builds this image and pushes it to ECR. The Go version appears in each `go.mod` and in the Dockerfile builder image. Bump them together.

## Runtime

Flow (`service-common/main-cmdline.go`): parse flags → `loadConfiguration()` → `authenticate()` → `getUpdatedIds()` (or use `-singleid`) → `processUpdatedIds()`. For each item, `processUpdatedIds()` fetches the item, calls `createBagContents()`, then `submitBagContents()` (APTrust register → S3 sync to the returned bucket/path via s3sync → initiate), and finally cleans the scratch directory.

Flags: `-startdate`/`-enddate` (YYYY-MM-DD, required unless `-singleid`), `-singleid`, `-resumeid` (skip items until this id), `-noprocess` (list ids only), `-nofiles` (skip downloads, implies `-nosubmit`), `-nosubmit`.

Configuration comes only from environment variables and any missing one is fatal (see each `config.go`). Common ones: `APT_REGISTER_URL`, `APT_SUBMIT_URL`, `APT_CLIENT_ID`, `API_ENDPOINT`, `API_UPDATED_PATH`, `API_ITEM_PATH`, `SCRATCH_FS`, `HTTP_TIMEOUT`. Dataverse also needs `API_TOKEN` and `API_FILE_PATH`. DSpace also needs `API_USERNAME` and `API_PASSWORD`. AWS credentials come from the default SDK chain. API path templates use placeholders such as `{{:id}}`.

Auth differs by connector. Dataverse sends an `X-Dataverse-key` header. DSpace (`dspace-connector/auth.go`) does an XSRF-cookie login flow, keeps a bearer token in package globals, and renews it periodically. Requests must go through `addAuthHeader()`.

`Version()` reads the `buildtag.*` file the Dockerfile creates in the working directory.

## Conventions

- Log with `log.Printf` using `INFO:` / `ERROR:` / `FATAL ERROR:` / `[CONFIG]` prefixes. Log secrets as `[REDACTED]`.
- Files open with a `//` comment header and end with `// end of file`.
