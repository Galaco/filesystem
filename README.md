[![GoDoc](https://godoc.org/github.com/golang-source-engine/filesystem?status.svg)](https://godoc.org/github.com/golang-source-engine/filesystem)
[![Go report card](https://goreportcard.com/badge/github.com/golang-source-engine/filesystem)](hhttps://goreportcard.com/report/github.com/golang-source-engine/filesystem)
[![GolangCI](https://golangci.com/badges/github.com/golang-source-engine/filesystem.svg)](https://golangci.com/r/github.com/golang-source-engine/filesystem)
[![codecov](https://codecov.io/gh/golang-source-engine/filesystem/branch/master/graph/badge.svg)](https://codecov.io/gh/golang-source-engine/filesystem)
[![CircleCI](https://circleci.com/gh/golang-source-engine/filesystem.svg?style=svg)](https://circleci.com/gh/golang-source-engine/filesystem)

# Filesystem

> A filesystem utility for reading Source engine game structures.

Source Engine is a little annoying in that there are potentially unlimited possible
locations that engine resources can be located. Filesystem provides a way to register 
and organise any potential resource path or filesystem, while preserving filesystem type
search priority.

A filesystem can either be manually defined, or created from a GameInfo.txt-derived KeyValues.

### Features
* Supports local directories
* Supports VPK's
* Supports BSP Pakfile
* Respects Source Engines search priority (pakfile->local directory->vpk)
* A ready to use Filesystem can be constructed from GameInfo.txt definitions