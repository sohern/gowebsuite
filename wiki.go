package main

import (
	"gowiki/structures"
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"
)

//Cache templates
var templates = template.Must(template.ParseFiles("templates/edit.html", "templates/view.html"))

//Validate path
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

//Page imported from structures
type Page structures.Page

func (p *Page) save() error {
	filename := "content/" + p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := "content/" + title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func frontPageHandler(w http.ResponseWriter, r *http.Request) {
	title := "frontpage"
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")

	createPageLinks([]byte(body))

	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func createPageLinks(body []byte) string {
	//search variable provides regex for finding the bracketed pages
	search := regexp.MustCompile("\\[[a-zA-Z]+\\]")
	body = search.ReplaceAllFunc(body, func(s []byte) []byte {
		//for each page link found in body, put just the text in m and return formatted html
		m := string(s[1 : len(s)-1])
		return []byte("<a href=\"/view/" + m + "\">" + m + "</a>")
	})
	//fmt.Println(string(body))
	return string(body)
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Here we will extract the page title from the Request,
		// and call the provided handler 'fn'
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func main() {
	//Define routes
	http.HandleFunc("/", frontPageHandler)
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))

	//Launch Server
	http.ListenAndServe(":8080", nil)
}
