package main

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
)

var (
	tmpl = `
	<html>
	<head>xrun debugger</head>
	<body>
	<div>To be debuggered</div>
	</body>
	</html>
	`
)

type DebugInfo struct {
}

var (
	debugInfo DebugInfo
)

// debugger home page
func home(resp http.ResponseWriter, req *http.Request) {
	t, err := template.New("debug").Parse(tmpl)
	if err != nil {
		fmt.Println(err)
		return
	}
	bytes := bytes.NewBufferString("")
	err = t.Execute(bytes, &debugInfo)
	if err != nil {
		fmt.Println(err)
		return
	}
	resp.Write(bytes.Bytes())
}

// add break point or remove break point
func breakpoint(resp http.ResponseWriter, req *http.Request) {

}

// execute next statement
func step(resp http.ResponseWriter, req *http.Request) {

}

// run web interface
func web() {
	http.HandleFunc("/", home)
	http.ListenAndServe("127.0.0.1:"+webPort, nil)
}
