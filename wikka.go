package main

import (
	"encoding/json"
	"fmt"
	"github.com/bmizerany/pat"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

type Configuration struct {
	Title     string
	Url       string
	Articles  string
	Templates string
	Host      string
	Frontpage string
	Editable  bool
}

type Article struct {
	Title      string
	ModifyDate time.Time
	Content    string
}

const (
	viewTemplate      = "view.template"
	editTemplate      = "edit.template"
	errorTemplate     = "error.template"
	containerTemplate = "main.template"
)

var templates map[string]string
var articles map[string]Article
var cfg *Configuration

// load all articles from a specific path
func loadArticles(path string) {
	articles = make(map[string]Article)
	entries, err := ioutil.ReadDir(path)

	if err != nil {
		log.Fatal("Failed to load articles: " + path)
	}

	for _, file := range entries {
		isTemplate := strings.HasSuffix(file.Name(), ".md")

		if isTemplate {
			content, err := ioutil.ReadFile(path + file.Name())

			if err != nil {
				log.Fatal("Failed to read article: " + path + file.Name())
			}

			text := string(content)
			title := strings.Split(file.Name(), ".")[0]
			article := Article{title, file.ModTime(), text}

			articles[strings.ToLower(title)] = article
			fmt.Println("Loaded article " + file.Name())
		}
	}
}

// load all templates from a specific path
func loadTemplates(path string) {
	templates = make(map[string]string)
	entries, err := ioutil.ReadDir(path)

	if err != nil {
		log.Fatal("Failed to load templates: " + path)
	}

	for _, file := range entries {
		isTemplate := strings.HasSuffix(file.Name(), ".template")

		if isTemplate {
			content, err := ioutil.ReadFile(path + file.Name())

			if err != nil {
				log.Fatal("Failed to read template file: " + path + file.Name())
			}

			result := string(content)
			templates[file.Name()] = result
			fmt.Println("Loaded template " + file.Name())
		}
	}

	// pre-render templates
	for key, _ := range templates {
		templates[key] = RenderTemplate(key)
	}

	CreateContainerTemplate(containerTemplate)
}

// Creates a new article render context
func (art *Article) CreateContext() map[string]string {
	return map[string]string{
		"Wiki.Title":         cfg.Title,
		"Wiki.Url":           cfg.Url,
		"Article.Title":      art.Title,
		"Article.Content":    RenderMarkdown(art.Content),
		"Article.RawContent": art.Content,
		"Article.ModifyDate": formatDate(art.ModifyDate),
	}
}

func CreateErrorContext(code int, message string) map[string]string {
	return map[string]string{
		"Wiki.Title":    cfg.Title,
		"Wiki.Url":      cfg.Url,
		"Article.Title": fmt.Sprintf("Error %d", code),
		"Error.Code":    fmt.Sprintf("%d", code),
		"Error.Message": message,
	}
}

func CreateCustomContext(title string, content string) map[string]string {
	return map[string]string{
		"Wiki.Title":    cfg.Title,
		"Wiki.Url":      cfg.Url,
		"Article.Title": title,
		"Article.RawContent": content,
	}
}

func formatDate(date time.Time) string {
	return date.Format("Monday, 2. January 15:04")
}

func showFrontpage(res http.ResponseWriter, req *http.Request) {
	http.Redirect(res, req, "/"+cfg.Frontpage, 301)
}

func viewArticle(res http.ResponseWriter, req *http.Request) {
	articleName := strings.ToLower(req.URL.Query().Get(":article"))
	fmt.Println("Article request")
	context := make(map[string]string)
	activeTemplate := ""

	if article, exists := articles[articleName]; exists {
		context = article.CreateContext()
		activeTemplate = viewTemplate
	} else {
		context = CreateErrorContext(404, articleName + " was not found. You may want to <a href=\"" +articleName + "/edit\">create this page!</a>")
		activeTemplate = errorTemplate
	}

	fmt.Fprint(res, RenderContainer(activeTemplate, context))
}

func editArticle(res http.ResponseWriter, req *http.Request) {
	article_name := strings.ToLower(req.URL.Query().Get(":article"))

	context := make(map[string]string)
	if article, exists := articles[article_name]; exists {
		context = article.CreateContext()
	} else {
		context = CreateCustomContext(article_name, "")
	}
	fmt.Fprint(res, RenderContainer(editTemplate, context))
}

func saveArticle(res http.ResponseWriter, req *http.Request) {
	article_name := strings.ToLower(req.URL.Query().Get(":article"))
	input_text := req.FormValue("textcontent")

	if len(input_text) > 0 {
		if article, ok := articles[article_name]; ok {
			err := ioutil.WriteFile(cfg.Articles+article.Title+".md", []byte(input_text), 0644)
			article.Content = input_text
			article.ModifyDate = time.Now()
			if err == nil {
				articles[article_name] = article
				http.Redirect(res, req, "/"+article.Title, 301)
				return
			}
		} else {
			valid_name, _ := regexp.MatchString("([A-Za-z\\-]{1,64})", article_name)
			if valid_name {
				active_article := Article{article_name, time.Now(), input_text}
				err := ioutil.WriteFile(cfg.Articles+active_article.Title+".md", []byte(input_text), 0644)
				if err == nil {
					articles[article_name] = active_article
					http.Redirect(res, req, "/"+active_article.Title, 301)
					return
				}
			}
		}
	}
	context := CreateErrorContext(500, "There happened something bad on the wiki server")
	res.WriteHeader(500)
	fmt.Fprint(res, RenderContainer(errorTemplate, context))
}

func loadConfiguration(path string) {
	file, err := os.Open(path)
	if err != nil {
		panic("Couldn't find configuration file: " + path)
	}
	decoder := json.NewDecoder(file)
	cfg = new(Configuration)
	err = decoder.Decode(cfg)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	start_time := time.Now()

	loadConfiguration("config.json")
	loadArticles(cfg.Articles)
	loadTemplates(cfg.Templates)

	mux := pat.New()
	mux.Get("/", http.HandlerFunc(showFrontpage))
	mux.Get("/:article", http.HandlerFunc(viewArticle))

	// create edit paths
	if cfg.Editable {
		mux.Get("/:article/edit", http.HandlerFunc(editArticle))
		mux.Post("/:article/save", http.HandlerFunc(saveArticle))
	}

	http.Handle("/", mux)

	diff_time := float32(time.Now().Nanosecond()-start_time.Nanosecond()) / 1000000.0
	fmt.Printf("Server up and running after %f milliseconds\n", diff_time)

	// Run webserver
	log.Fatal(http.ListenAndServe(cfg.Host, nil))
}
