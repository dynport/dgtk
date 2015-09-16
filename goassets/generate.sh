#!/bin/bash

# use tpl/goassets-script-tpl/goassets.go and replace the `const tpl` line with the content of tpl/goassets-tpl/gen.go but replace the `package main` with `package {{ .Pkg }}` and append the init template from tpl/goassets-tpl/compiled.go.tpl
{ cat tpl/goassets-script-tpl/goassets.go | grep -v "^const tpl"; echo -n 'const tpl = `'; cat tpl/goassets-tpl/gen.go | sed 's/^package main/package {{ .Pkg }}/'; cat tpl/goassets-tpl/compiled.go.tpl; echo '`'; }
