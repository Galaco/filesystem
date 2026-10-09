[![GoDoc](https://godoc.org/github.com/galaco/filesystem?status.svg)](https://godoc.org/github.com/galaco/filesystem)
[![Go report card](https://goreportcard.com/badge/github.com/galaco/filesystem)](hhttps://goreportcard.com/report/github.com/galaco/filesystem)
[![GolangCI](https://golangci.com/badges/github.com/galaco/filesystem.svg)](https://golangci.com/r/github.com/galaco/filesystem)
[![codecov](https://codecov.io/gh/galaco/filesystem/branch/master/graph/badge.svg)](https://codecov.io/gh/galaco/filesystem)
[![CircleCI](https://circleci.com/gh/galaco/filesystem.svg?style=svg)](https://circleci.com/gh/galaco/filesystem)

# Filesystem

> A filesystem utility for reading Source engine game structures.

Source Engine is a little annoying in that there are potentially unlimited possible
locations that engine resources can be located. Filesystem provides a way to register 
and organise any potential resource path or filesystem, while preserving the engine's
search priority.

A filesystem can either be manually defined, or created from a GameInfo.txt-derived KeyValues.

### Features
* Supports local directories
* Supports VPK's
* Supports BSP Pakfile
* Respects the Source Engine's search priority since SteamPipe: the pakfile first, then each search path in the
order it was registered, which is the order gameinfo.txt lists them, vpks and directories alike
* A ready to use Filesystem can be constructed from GameInfo.txt definitions. Only "game" search paths are added,
and a path ending in `/*` (such as `custom/*`) adds each of its subdirectories and vpks in alphabetical order