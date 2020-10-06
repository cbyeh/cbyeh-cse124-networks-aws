// Server to be hosted on AWS
package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/gendata", gendataHandler)
	log.Fatal(http.ListenAndServe("0.0.0.0:8000", nil))
}

// Handler prints numBytes number of "!"
func gendataHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Print(err)
	}
	numBytes, err := strconv.Atoi(r.Form.Get("numBytes"))
	if err != nil {
		log.Print(err)
	}
	fmt.Fprintf(w, "%s", strings.Repeat("!", numBytes))
}

// Handler echoes the HTTP request.
func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s %s %s\n", r.Method, r.URL, r.Proto)
	for k, v := range r.Header {
		fmt.Fprintf(w, "Header[%q] = %q\n", k, v)
	}
	fmt.Fprintf(w, "Host = %q\n", r.Host)
	fmt.Fprintf(w, "RemoteAddr = %q\n", r.RemoteAddr)
	if err := r.ParseForm(); err != nil {
		log.Print(err)
	}
	for k, v := range r.Form {
		fmt.Fprintf(w, "Form[%q] = %q\n", k, v)
	}
}
