#!/usr/bin/env bash

if [[ ! $(which deadcode) ]]; then
  echo "deadcode must be installed via go get github.com/remyoudompheng/go-misc/deadcode"
  exit 1
fi

dirs=$(find . -name "*.go" | grep -v Godep | while read dir; do dirname $dir; done | sort | uniq)

for d in $dirs; do
  res=$(cd $d && deadcode 2>&1 | cut -d ' ' -f 2- | while read line; do echo "$d/$line"; done)
  if [[ -n $res ]]; then
    echo "$res"
  fi
done
