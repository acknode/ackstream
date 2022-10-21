#!/bin/sh

set -e

NOW=$(date +%Y.%-m%d.%-H%M)
echo "$NOW" > .version

git add .version && git commit -m "ci($NOW): ✨🐛🚨"

TARGET=${1:-origin}
echo "---------------------------"
printf "Pushing... NOW --> %s" "$TARGET"
echo "---------------------------"
git push "$TARGET"
