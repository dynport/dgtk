package cli

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// Interface that must be implemented by actions. This is interface is used by the RegisterAction function. The Run
// method of the implementing type will be called if the given route matches it.
type Runner interface {
	Run() error
}

// Create a new router.
func NewRouter() *Router {
	r := &Router{}
	r.root = &routingTreeNode{children: map[string]*routingTreeNode{}}
	return r
}

var NoRouteError = fmt.Errorf("no route matched")

// Run the given arguments against the registered actions, i.e. try to find a matching route and run the according
// action.
func (r *Router) Run(args ...string) (e error) {
	if r.initFailed {
		log.Fatal("errors found during initialization")
	}
	// Find action and parse args.
	node, args := r.findNode(args, true)
	if node != nil && node.action != nil {
		if e := node.action.parseArgs(args); e != nil {
			node.showHelp()
			return e
		}
	} else { // Failed to find node.
		node.showHelp()
		return NoRouteError
	}

	return node.action.runner.Run()
}

// Run the arguments from the commandline (aka os.Args) against the registered actions, i.e. try to find a matching
// route and run the according action.
func (r *Router) RunWithArgs() (e error) {
	return r.Run(os.Args[1:]...)
}

type annonymousAction struct {
	runner      func() error
	description string
}

func (aA *annonymousAction) Run() error {
	return aA.runner()
}

// Register the given function as handler for the given route. This is a shortcut for actions that don't need options or
// arguments. A description can be provided as an optional argument.
func (r *Router) RegisterFunc(path string, f func() error, desc string) {
	aA := &annonymousAction{runner: f}
	r.Register(path, aA, desc)
}

// Register the given action (some struct implementing the Runner interface) for the given route.
func (r *Router) Register(path string, runner Runner, desc string) {
	a, e := newAction(path, runner, desc)
	if e != nil {
		log.Printf("%s", e)
		r.initFailed = true
		return
	}

	pathSegments := strings.Split(a.path, "/")
	node, pathSegments := r.findNode(pathSegments, false)
	if node != nil {
		if node.action != nil {
			log.Printf("failed to register action for path %q: action for path %q already registered", a.path, node.action.path)
			r.initFailed = true
			return
		} else if len(pathSegments) == 0 && len(node.children) > 0 {
			log.Printf("failed to register action for path %q: longer paths with this prefix exist", a.path)
			r.initFailed = true
			return
		}
	} else {
		node = r.root
	}

	for _, p := range pathSegments {
		newNode := &routingTreeNode{children: map[string]*routingTreeNode{}}
		node.children[p] = newNode
		node = newNode
	}

	node.action = a
}

func (r *Router) showHelp() {
}
