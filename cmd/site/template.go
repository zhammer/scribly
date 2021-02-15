package main

import (
	"net/http"

	"github.com/mssola/user_agent"
)

type ViewData struct {
	Data    interface{}
	Request *http.Request
}

// v *ViewData doesn't seem to work. i guess go templates don't have the same
// pointer-whichever-way-you-want-it magic as actual go code.
func (v ViewData) Query(key string) string {
	return v.Request.URL.Query().Get(key)
}

func (v ViewData) Mobile() bool {
	ua := user_agent.New(v.Request.Header.Get("user-agent"))
	return ua.Mobile()
}

// at this point this is just a helper function
func (v ViewData) Ternary(cond bool, a interface{}, b interface{}) interface{} {
	if cond {
		return a
	} else {
		return b
	}
}

// lets us pass some other data into other templates while preserving the top-level
// request data (and other context we may add later.)
func (v ViewData) Propogate(data interface{}) ViewData {
	return ViewData{
		Data:    data,
		Request: v.Request,
	}
}
