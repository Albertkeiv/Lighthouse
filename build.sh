#!/usr/bin/env bash
set -e

fyne-cross linux -arch=amd64
fyne-cross windows -arch=amd64
fyne-cross darwin -arch=amd64
