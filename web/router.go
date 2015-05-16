package web

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type Router map[string]map[string]func(http.ResponseWriter, *http.Request, httprouter.Params)

func (r Router) Handler() (h http.Handler, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf(fmt.Sprint(r))
		}
	}()
	router := httprouter.New()

	// to lookup method names to router Methods
	methods := map[string]func(string, httprouter.Handle){
		"GET":     router.GET,
		"POST":    router.POST,
		"DELETE":  router.DELETE,
		"PATCH":   router.PATCH,
		"HEAD":    router.HEAD,
		"OPTIONS": router.OPTIONS,
	}

	for methodName, sr := range r {
		if method, ok := methods[methodName]; !ok {
			return nil, fmt.Errorf("method %q not supported", methodName)
		} else {
			for path, handler := range sr {
				method(path, handler)
			}
		}
	}
	return router, nil
}
