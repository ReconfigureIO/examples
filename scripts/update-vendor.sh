#!/bin/bash

find . -maxdepth 1 -mindepth 1 -type d -not -path '*/\.*' -not -path 'scripts' | xargs -I{} bash -c "cd {} && cp ../glide.lock . && glide --yaml ../glide.yaml i && rm glide.lock && git add -A vendor"
