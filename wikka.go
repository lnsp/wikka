package main

import (
	"encoding/json"
	"fmt"
	"github.com/bmizerany/pat"
	"github.com/shurcooL/go/github_flavored_markdown"
  "github.com/microcosm-cc/bluemonday"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Configuration struct {
	Title     string
	Url       string
	Articles  string
	Templates string
}

type Article struct {
	Title      string
	ModifyDate time.Time
	Content    string
}

var templates map[string]string
var articles map[string]Article
var cfg *Configuration

func load_articles(path string) {
	articles = make(map[string]Article)
	info, err := ioutil.ReadDir(path)

	if err != nil {
		log.Fatal("Failed to load articles: " + path)
	}

	for _, file := range info {
		isTemplate := strings.HasSuffix(file.Name(), ".md")

		if isTemplate {
			content_bytes, err := ioutil.ReadFile(path + file.Name())

			if err != nil {
				log.Fatal("Failed to read article: " + path + file.Name())
			}

			content := string(content_bytes)
			title := strings.Split(file.Name(), ".")[0]
			article := Article{title, file.ModTime(), content}

			articles[strings.ToLower(title)] = article
			fmt.Println("Loaded article " + file.Name())
		}
	}
}

func load_templates(path string) {
	templates = make(map[string]string)
	info, err := ioutil.ReadDir(path)

	if err != nil {
		log.Fatal("Failed to load templates: " + path)
	}

	for _, file := range info {
		isTemplate := strings.HasSuffix(file.Name(), ".template")

		if isTemplate {
			content_bytes, err := ioutil.ReadFile(path + file.Name())

			if err != nil {
				log.Fatal("Failed to read template file: " + path + file.Name())
			}

			content := string(content_bytes)
			templates[file.Name()] = content
			fmt.Println("Loaded template " + file.Name())
		}
	}
}

func render_markdown(md string) string {
	md_bytes := []byte(md)
	text_bytes := github_flavored_markdown.Markdown(md_bytes)
	sanitized_bytes := bluemonday.UGCPolicy().SanitizeBytes(text_bytes)
	return string(sanitized_bytes)
}

func render_template(tmp string, context map[string]string) string {
	result := templates[tmp]

	for key, value := range context {
		result = strings.Replace(result, "{"+key+"}", value, -1)
	}

	return result
}

func render_combined(context map[string]string) string {
	result := templates["combined.template"]
	for key, value := range templates {
		result = strings.Replace(result, "{"+key+"}", value, -1)
	}
	for key, value := range context {
		result = strings.Replace(result, "{"+key+"}", value, -1)
	}
	return result
}

func handle_index(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Location", "/index")
	res.WriteHeader(301)

	if req.Method == "GET" {
		fmt.Fprintln(res, "<a href\"="+"/index\">Redirect to index page ...</a>")
	}
}

func handle_view(res http.ResponseWriter, req *http.Request) {
	article_name := strings.ToLower(req.URL.Query().Get(":article"))
	content_tmp := "notfound.template"

	context := map[string]string{
		"Wiki.Title":    cfg.Title,
		"Wiki.Url":      cfg.Url,
		"Article.Title": article_name,
	}

	if val, ok := articles[article_name]; ok {
		context["Article.Title"] = val.Title
		context["Article.ModifyDate"] = val.ModifyDate.Format("Mon Jan 2 15:04:05 -0700 MST 2006")
		context["Article.Content"] = render_markdown(val.Content)
		content_tmp = "view.template"
	}

	context["content"] = render_template(content_tmp, context)
	fmt.Fprint(res, render_combined(context))
}

func handle_edit(res http.ResponseWriter, req *http.Request) {
	fmt.Fprint(res, "Hello from edit page!") // TODO: Implement editing
}

func handle_save(res http.ResponseWriter, req *http.Request) {
	fmt.Fprint(res, "Hello from save page!") // TODO: Implement saving
}

func load_config(path string) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal("Couldn't find configuration file: " + path)
	}
	decoder := json.NewDecoder(file)
	cfg := new(Configuration)
	err = decoder.Decode(cfg)
	if err != nil {
		log.Fatal("Error while parsing configuration file")
	}
}

func main() {
	load_config("config.json")
	load_templates(cfg.Templates)
	load_articles(cfg.Articles)

	mux := pat.New()
	mux.Get("/", http.HandlerFunc(handle_index))
	mux.Get("/:article", http.HandlerFunc(handle_view))
  mux.Get("/:article/", http.HandlerFunc(handle_view))
	mux.Get("/:article/edit", http.HandlerFunc(handle_edit))
	mux.Post("/:article/save", http.HandlerFunc(handle_save))

	http.Handle("/", mux)

	// Run webserver
	log.Fatal(http.ListenAndServe(":3000", nil))
}
