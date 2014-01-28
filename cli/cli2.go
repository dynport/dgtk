// golang CLI parameter handling
//
// This is dynport's version of a golang CLI handler. Given a struct that implements an interface with fields that have
// annotations, an CLI is built that can be run against a list of strings (taken from os.Args for example).
//
// See the basic example on how to use this library with a router. The other example shows the short annotation notation
// and the direct usage of actions without a router.
//
// The following constraints or special behaviors are to be taken into account:
//	* Options (type "opt") are given in short or long form ("-h" vs. "--help"). Each option must have at least one
//	  modifier set.
//	* Required options must be present. A default value is preset in the struct.
//	* Options with a boolean value are internally handled as flags, i.e. presence of the flag indicates true.
//	* Ordering of arguments is defined by the position in the action's struct (first come first serve).
//	* Arguments (type "arg") may be variadic (type in the struct must be a slice), i.e. arbitrary can be given. If the
//	  argument is required, at least one value must be present. Only the last arguments can be variadic.
//	* Non variadic arguments must always be given.
package cli
