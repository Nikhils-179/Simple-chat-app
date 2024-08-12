package main

import (
	"flag"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"text/template"
)

// templ represents a single template
type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

// serves HTTP request
func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})
	t.templ.Execute(w, r)
}

func main() {
	var addr = flag.String("addr", ":8080", "The address of the application")
	flag.Parse() // parse the flags

	r := newRoom()

	http.Handle("/", &templateHandler{filename: "chat.html"})

	http.Handle("/room",r)

	go r.run()

	log.Println("Starting web server on" , *addr)
	if err := http.ListenAndServe(*addr, nil);err != nil  {
		log.Fatal("Listen and serve:" , err)
	}

}
