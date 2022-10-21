#!/bin/sh

set -e

NOW=$(date +%Y.%-m%d.%-H%M)
echo -n "$NOW" > .version

git add .version && git commit -m "ci($NOW): ✨🐛🚨"

TARGET=${1:-origin}
echo -e "\n---------------------------"
echo -e "Pushing... NOW --> $TARGET"
echo -e "---------------------------\n"
git push "$TARGET"
