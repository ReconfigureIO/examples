#!/bin/bash

find . -maxdepth 1 -mindepth 1 -type d -not -path '*/\.*' -not -path './scripts' | xargs -I{} bash -c "cd {} && cp ../scripts/files/glide.lock . && glide --yaml ../scripts/files/glide.yaml i && rm glide.lock && git add -A vendor"
