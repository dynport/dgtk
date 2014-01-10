package cli2

import (
	"fmt"
	"sort"
	"strings"
)

// The basic datastructure used in cli2. A router is added actions for different paths using the "Register" and
// "RegisterFunc" methods. These actions can be executed using the "Run" and "RunWithArgs" methods.
type Router struct {
	root *routingTreeNode

	initFailed bool
}

// A tree used for easy access to the matching action. An action can only be set if there are no children, i.e. only
// leaf nodes can have actions.
type routingTreeNode struct {
	children map[string]*routingTreeNode
	action   *action
}

func (rt *routingTreeNode) showHelp() {
	if rt.action != nil {
		rt.action.showHelp()
	} else {
		t := &table{}
		rt.showTabularHelp(t)
		fmt.Println(t)
	}
}

func (rt *routingTreeNode) showTabularHelp(t *table) {
	if rt.action != nil {
		rt.action.showTabularHelp(t)
	} else {
		pathSegments := make([]string, 0, len(rt.children))
		for k, _ := range rt.children {
			pathSegments = append(pathSegments, k)
		}
		sort.Strings(pathSegments)

		for _, ps := range pathSegments {
			rt.children[ps].showTabularHelp(t)
		}
	}
}

// Find the node matching most segments of the given path. Will return the according tree node and the remaining (non
// matched) path segments.
func (r *Router) findNode(pathSegments []string, fuzzy bool) (*routingTreeNode, []string) {
	node := r.root
	for i, p := range pathSegments {
		if c, found := node.children[p]; found {
			node = c
		} else {
			if fuzzy { // try fuzzy search
				candidates := []string{}
				for key, _ := range node.children {
					if strings.HasPrefix(key, p) {
						candidates = append(candidates, key)
					}
				}
				if len(candidates) == 1 {
					node = node.children[candidates[0]]
					continue
				}
			}
			return node, pathSegments[i:]
		}
	}
	return node, nil
}
