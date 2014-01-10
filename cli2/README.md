# cli2

A library to easily create command line interfaces.


## Example

The following example creates a simple CLI action for running commands at a remote host (like `example run on host -c "uname -a" 192.168.1.1` given the binary has the `example` name and the `uname` command should be executed on `192.168.1.1`).

	// Struct used to configure an action.
	type ExampleRunner struct {
		Verbose bool   `cli:"type=opt short=v long=verbose"`               // Flag (boolean option) example. This is either set or not.
		Command string `cli:"type=opt short=c long=command required=true"` // Option that has a default value.
		Hosts   string `cli:"type=arg required=true"`                      // Argument with at least one required.
	}
	
	// Run the action. Called when the cli2.Run function is called with a route matching the one of this action.
	func (er *ExampleRunner) Run() error {
		// Called when action matches route.
		if er.Verbose {
			log.Printf("Going to execute %q at the following hosts: %v", er.Command, er.Hosts)
		}
		// [..] Executing the SSH command is left to the reader.
		return nil
	}
	
	// Basic example that shows how to register an action to a route and execute it.
	func Example_basic() {
		router := NewRouter()
		router.Register("run/on/hosts", &ExampleRunner{}, "This is an example that pretends to run a given command on a set of hosts.")
		router.Run("run", "on", "host", "-v", "-c", "uname -a", "192.168.1.1")
		router.RunWithArgs() // Run with args given on the command line.
	}


## Usage

There are three steps to take:
	1. Create a struct with the options and arguments to be supported. Description of those entities is done using
	   annotations.
	2. Implement the Runner interface for this struct.
	3. Register the struct as action with a path on the router.

