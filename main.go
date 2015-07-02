package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/rs/cors"
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
	fmt.Println("looking for comments in", path)
	fmt.Println("will listen for http traffic on", addr)

	http.Handle("/", cors.Default().Handler(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "POST" {
					handlePost(w, r)
				} else {
					handleGet(w, r)
				}
			},
		),
	),
	)
	http.ListenAndServe(addr, nil)
}

func handlePost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		fmt.Println(r.URL.Path, "POST", "unparseable form", err)
		http.Error(w, "Couldn't parse form: "+err.Error(), http.StatusBadRequest)
		return
	}
	// honey pot
	if r.Form.Get("company") != "" {
		fmt.Println(r.URL.Path, "POST", "company not empty:", r.Form.Get("company"))
		http.Error(w, "ok", http.StatusTeapot)
		return
	}

	c := Comment{
		Parent:    r.Form.Get("post"),
		Ts:        time.Now(),
		Message:   r.Form.Get("message"),
		Ipaddress: "", // TODO
		Author:    r.Form.Get("name"),
		Email:     r.Form.Get("email"),
		Link:      r.Form.Get("url"),
	}
	c.Hash = fmt.Sprintf("%x", md5.Sum([]byte(c.Email)))

	err = c.Save(path)
	if err != nil {
		fmt.Println(r.URL.Path, "POST", "Couldn't save comment: ", err)
		http.Error(w, "Couldn't save comment: "+err.Error(), http.StatusInternalServerError)
		return
	}
	bytes, e := json.Marshal(c)
	if e != nil {
		http.Error(w, fmt.Sprintf("Error marshalling JSON:'%s'", e), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	// strip out all /prependend/path/separators/if/any so that it works under arbitrary paths, for proxying etc
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
	sort.Sort(ByTsAsc(comments))
	fmt.Printf("> %s -> serving %d comments\n", slug, len(comments))
	bytes, e := json.Marshal(comments)
	if e != nil {
		http.Error(w, fmt.Sprintf("Error marshalling JSON:'%s'", e), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}
