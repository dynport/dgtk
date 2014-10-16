#!/bin/bash

{ cat script/goassets.go | grep -v "^const tpl"; echo -n 'const tpl = `'; cat tpl/gen.go | sed 's/^package main/package {{ .Pkg }}/'; cat tpl/compiled.go.tpl; echo '`'; }
