# tskip

A little bit more convenience and locality for writing go tests.

Introducing: "function based tests"

## What

tskip allows using the go test methods `testing.Log` etc. from inside test helpers without loosing locality. It can also fix the locality of errors in sub packages.

## How

The tskip library methods `found in tskip/tskip` uses `caller.Runtime` with the right skip values to get the location from where the function is called. It then prepends the output with `\r` to "remove" the location which is printed by the stdlib.

The vim quickfix window does not care about carriage returns though, so vim user would still be directed to location deleted by the carriage return. That is why tskip also provides a simple decorator binary/script which wraps `go test` and completely strips this not wanted output from what vim sees.

	tskip -v ./... # just decorates the output of go test -v ./...

As go changes to the directories of all subpackages when running tests, most testing output in go only displays file names which can be misleading also mess up the vim quickfix window. To fix this, tskip sets the `TEST_ROOT` to the directory from which tskip is executed so this prefix can be removed from the file names. If `TEST_ROOT` is not set the tskip library prints the full file path (which also does not break vim quickfix).

## What can I do with it?

With tskip you can write your own test helpers but also improve locality for table based tests by migrating them to "function based tests".


TODO: add example and explain why this is better

## Examples

See `main_test.go` for some examples.

Run 

	TEST_SIMULATE=true go run main.go -v ./...

to test the output on the command line or

	:make simulate

inside vim to see how this fixes the quickfix issue.
