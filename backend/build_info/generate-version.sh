##!/bin/sh
#
## Get the current tag (if any)
#tag=$(git describe --exact-match --tags 2> /dev/null)
#
## If there's a tag and it starts with "backend/", write it to version.txt
#if [ -n "$tag" ] && echo "$tag" | grep -q '^backend/'; then
#  echo "exact tag"
#  echo "$tag" > version.txt
#else
#  echo "not on a tag using branch as version"
#  # Otherwise, write the current branch and commit sha to version.txt
#  branch=$(git rev-parse --abbrev-ref HEAD)
#  commit_sha=$(git rev-parse --short HEAD)
#  echo "$branch-$commit_sha" > version.txt
#fi

git describe --tags --always --match "backend/*" > version.txt
