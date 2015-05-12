package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

var path string
var addr string

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "%s <comments-path> <addr>\n", os.Args[0])
		os.Exit(2)
	}
	path = os.Args[1]
	if strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}
	addr = os.Args[2]
	fmt.Println("looking for comments in" + path)
	fmt.Println("will listen for http traffic on" + addr)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			handlePost(w, r)
		} else {
			handleGet(w, r)
		}
	})
	http.ListenAndServe(addr, nil)
}

func handlePost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Couldn't parse form: "+err.Error(), http.StatusBadRequest)
		return
	}
	if r.Form.Get("company") != "" {
		http.Error(w, "ok", http.StatusTeapot)
		return
	}

	c := Comment{
		Parent:    r.Form.Get("post"),
		Ts:        time.Now(),
		Message:   r.Form.Get("message"),
		Ipaddress: "foo",
		Author:    r.Form.Get("name"),
		Email:     r.Form.Get("email"),
		Link:      r.Form.Get("url"),
	}
	err = c.Save(path)
	if err != nil {
		http.Error(w, "Couldn't save comment: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	// strip out all /prependend/path/separators/if/any so that it work under arbitrary paths, for proxying etc
	cutoff := strings.LastIndex(r.URL.Path, "/")
	slug := r.URL.Path[cutoff+1:]
	if slug == "" {
		http.Error(w, "specify a slug, you dufus", http.StatusMethodNotAllowed)
		return
	}
	comments, err := FindComments(path, slug)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	bytes, e := json.Marshal(comments)
	if e != nil {
		http.Error(w, fmt.Sprintf("Error marshalling JSON:'%s'", e), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}
