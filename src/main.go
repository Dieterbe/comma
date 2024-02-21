package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/rs/cors"
)

var path string
var addr string
var special string

func main() {
	if len(os.Args) != 3 && len(os.Args) != 4 {
		fmt.Fprintf(os.Stderr, "%s <comments-path> <addr> [special-value]\n", os.Args[0])
		os.Exit(2)
	}
	path = strings.TrimSuffix(os.Args[1], "/")

	addr = os.Args[2]
	fmt.Println("looking for comments in", path)
	fmt.Println("will listen for http traffic on", addr)
	if len(os.Args) == 4 {
		special = os.Args[3]
	}

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
	err := r.ParseMultipartForm(1024 * 10)
	if err != nil {
		fmt.Println(r.URL.Path, "POST", "unparseable form", err)
		http.Error(w, "Couldn't parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	if special != "" && r.Form.Get("special") != special {
		fmt.Println(r.URL.Path, "POST", "bad 'special' value", r.Form.Get("special"))
		http.Error(w, "incorrect value provided", http.StatusForbidden)
		return
	}

	c, err := CommentFromForm(r.Form)
	if err != nil {
		fmt.Println(r.URL.Path, "POST", "invalid submission", err)
		http.Error(w, "Invalid submission: "+err.Error(), http.StatusBadRequest)
		return
	}

	err = c.Save(path)
	if err != nil {
		fmt.Println(r.URL.Path, "POST", "Error: Couldn't save comment: ", err)
		http.Error(w, "Couldn't save comment: "+err.Error(), http.StatusInternalServerError)
		return
	}
	bytes, e := json.Marshal(c)
	if e != nil {
		fmt.Println(r.URL.Path, "POST", "Error marshalling JSON: ", err)
		http.Error(w, fmt.Sprintf("Error marshalling JSON:'%s'", e), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

func CommentFromForm(form url.Values) (Comment, error) {
	if form.Get("message") == "" {
		return Comment{}, fmt.Errorf("message is required")
	}
	if form.Get("name") == "" {
		return Comment{}, fmt.Errorf("name is required")
	}
	if form.Get("email") == "" {
		return Comment{}, fmt.Errorf("email is required")
	}
	if form.Get("post") == "" {
		return Comment{}, fmt.Errorf("post is required")
	}
	return Comment{
		Parent:    form.Get("post"),
		Ts:        time.Now(),
		Message:   form.Get("message"),
		Ipaddress: "", // TODO
		Author:    form.Get("name"),
		Email:     form.Get("email"),
		Link:      form.Get("url"),
		Hash:      fmt.Sprintf("%x", md5.Sum([]byte(form.Get("email")))),
	}, nil

}

func handleGet(w http.ResponseWriter, r *http.Request) {
	// strip out all /prependend/path/separators/if/any so that it works under arbitrary paths, for proxying etc
	cutoff := strings.LastIndex(r.URL.Path, "/")
	slug := r.URL.Path[cutoff+1:]
	if slug == "" {
		fmt.Println(r.URL.Path, "GET", "No slug specified")
		http.Error(w, "specify a slug, you dufus", http.StatusBadRequest)
		return
	}
	comments, err := FindComments(path, slug)
	if err != nil {
		fmt.Println(r.URL.Path, "GET", "Error: FindComments() failed: "+err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sort.Sort(ByTsAsc(comments))
	fmt.Printf("> %s -> serving %d comments\n", slug, len(comments))
	bytes, e := json.Marshal(comments)
	if e != nil {
		fmt.Println(r.URL.Path, "GET", "Error marshalling JSON: "+err.Error())
		http.Error(w, fmt.Sprintf("Error marshalling JSON:'%s'", e), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}
