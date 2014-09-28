package main_module

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"lib/blob"

	"appengine"
)

const (
	NOT_FOUND_MSG = "404 page not found"
)

func apiHandler(rw http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	defer func() {
		r := recover()
		c.Errorf("%v\n", r)
	}()

	now := time.Now()

	var ID string
	parts := strings.FieldsFunc(r.URL.Path, func(r rune) bool { return r == '/' })
	if len(parts) != 2 && len(parts) != 3 {
		http.Error(rw, NOT_FOUND_MSG, http.StatusNotFound)
		return
	}

	if len(parts) == 3 {
		ID = parts[2]
	}

	switch r.Method {
	case "GET":
		if len(ID) == 0 {
			// Valid to do method not allowed since this is
			// technically a different URL (while get is
			// supported on /api/blob/blah, it's not on
			// /api/blob).
			http.Error(rw, "method GET not supported", http.StatusMethodNotAllowed)
			return
		}

		tp, err := blob.GetCurrentBlob(c, ID)
		if err != nil {
			// TODO(synful): differentiate between not found errors
			// (which are the user's fault) and DB errors (internal
			// server error)
			http.Error(rw, fmt.Sprintf("%v", err), http.StatusInternalServerError)
			return
		}
		fmt.Fprint(rw, string(tp.Primitive))
	case "POST":
		bodytmp, err := ioutil.ReadAll(r.Body)
		body := string(bodytmp)
		if err != nil {
			// TODO(synful): log error
			http.Error(rw, "", http.StatusInternalServerError)
			return
		}
		if len(ID) == 0 {
			ID, err := blob.PutNewBlob(c, now, body)
			if err != nil {
				// TODO(synful): differentiate between json parse errors
				// (bad request) and DB errors (internal server error)
				http.Error(rw, fmt.Sprintf("%v", err), http.StatusInternalServerError)
				return
			}
			c.Infof("%v\n", ID)
			fmt.Fprint(rw, ID)
		} else {
			err = blob.UpdateBlob(c, now, ID, body)
			if err != nil {
				// TODO(synful): differentiate between json parse errors
				// (bad request) and DB errors (internal server error)
				http.Error(rw, fmt.Sprintf("%v", err), http.StatusInternalServerError)
				return
			}
		}
	default:
		http.Error(rw, fmt.Sprintf("method %v not supported", r.Method), http.StatusMethodNotAllowed)
	}
}
