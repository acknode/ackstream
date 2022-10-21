#!/bin/sh

set -e

NOW=$(date +%Y.%-m%d.%-H%M)
echo "$NOW" > .version

git add .version && git commit -m "ci($NOW): ✨🐛🚨"

TARGET=${1:-origin}
printf "\n---------------------------"
printf "Pushing... NOW --> %s" "$TARGET"
printf "---------------------------\n"
git push "$TARGET"
