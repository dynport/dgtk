#!/bin/bash

{ cat script/goassets.go | grep -v "^const TPL"; echo -n 'const TPL = `'; cat tpl/gen.go | sed 's/^package main/package {{ .Pkg }}/'; cat tpl/compiled.go.tpl; echo '`'; }
