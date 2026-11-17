#!/bin/sh
# Render the piece SVGs in pkg/assets/pieces/*/ to the PNGs chess-cli embeds.
# Needs Inkscape (and optionally optipng to shrink the results).
set -e
cd "$(dirname "$0")/../pkg/assets/pieces"
for svg in */*.svg; do
	png="${svg%.svg}.png"
	inkscape "$svg" --export-type=png --export-width=384 --export-height=384 --export-filename="$png" >/dev/null 2>&1
done
if command -v optipng >/dev/null; then optipng -quiet -o2 */*.png; fi
