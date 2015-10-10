package main

import (
	"fmt"
	"github.com/bmizerany/pat"
	"log"
	"net/http"
	"io/ioutil"
	"regexp"
	"strings"
	"time"
)

type Configuration struct {
	Title     			string
	Url       			string
	Articles  			string
	Templates 			string
	Host      			string
	Frontpage 			string
	Editable  			bool
	MinimumTextLength 	int
}

type Article struct {
	Title      string
	ModifyDate time.Time
	Content    string
}

const (
	VIEW_TEMPLATE      = "view.template"
	EDIT_TEMPLATE      = "edit.template"
	ERROR_TEMPLATE     = "error.template"
	CONTAINER_TEMPLATE = "main.template"
)

var templates map[string]string
var articles map[string]Article
var cfg *Configuration

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
	context := make(map[string]string)
	activeTemplate := ""

	if article, exists := articles[articleName]; exists {
		context = article.CreateContext()
		activeTemplate = VIEW_TEMPLATE
	} else {
		context = CreateErrorContext(404, articleName + " was not found. You may want to <a href=\"" +articleName + "/edit\">create this page!</a>")
		activeTemplate = ERROR_TEMPLATE
	}

	fmt.Fprint(res, RenderContainer(activeTemplate, context))
}

func editArticle(res http.ResponseWriter, req *http.Request) {
	articleName := strings.ToLower(req.URL.Query().Get(":article"))

	context := make(map[string]string)
	if article, exists := articles[articleName]; exists {
		context = article.CreateContext()
	} else {
		context = CreateCustomContext(articleName, "")
	}
	fmt.Fprint(res, RenderContainer(EDIT_TEMPLATE, context))
}

func saveArticle(res http.ResponseWriter, req *http.Request) {
	articleName := strings.ToLower(req.URL.Query().Get(":article"))
	savedText := req.FormValue("textcontent")

	errorMessage := ""

	if len(savedText) >= cfg.MinimumTextLength {
		if article, ok := articles[articleName]; ok {
			err := ioutil.WriteFile(cfg.Articles+article.Title+".md", []byte(savedText), 0644)
			article.Content = savedText
			article.ModifyDate = time.Now()
			if err == nil {
				articles[articleName] = article
				http.Redirect(res, req, "/"+article.Title, 301)
				return
			} else {
				log.Fatal(err)
				errorMessage = err.Error()
			}
		} else {
			validName, _ := regexp.MatchString("([A-Za-z\\-]{1,64})", articleName)
			if validName {
				activeArticle := Article{articleName, time.Now(), savedText}
				err := ioutil.WriteFile(cfg.Articles+activeArticle.Title+".md", []byte(savedText), 0644)
				if err == nil {
					articles[articleName] = activeArticle
					http.Redirect(res, req, "/"+activeArticle.Title, 301)
					return
				} else {
					log.Fatal(err)
					errorMessage = err.Error()
				}
			} else {
				// name not valid
				errorMessage = "The article's name is not valid."
			}
		}
	} else {
		errorMessage = fmt.Sprintf("The text content of the entry had a length of less than %d characters.", cfg.MinimumTextLength)
	}
	context := CreateErrorContext(500, errorMessage)
	res.WriteHeader(500)
	fmt.Fprint(res, RenderContainer(ERROR_TEMPLATE, context))
}

func main() {
	start_time := time.Now()

	LoadConfiguration("config.json")
	LoadArticles(cfg.Articles)
	LoadTemplates(cfg.Templates)

	mux := pat.New()
	mux.Get("/", http.HandlerFunc(showFrontpage))
	mux.Get("/:article", http.HandlerFunc(viewArticle))

	// create edit paths
	if cfg.Editable {
		mux.Get("/:article/edit", http.HandlerFunc(editArticle))
		mux.Post("/:article/save", http.HandlerFunc(saveArticle))
	}

	http.Handle("/", mux)

	diff_time := float64(time.Now().Nanosecond()-start_time.Nanosecond()) / 1000000.0
	log.Printf("Server up and running after %.3f milliseconds\n", diff_time)

	// Run webserver
	log.Fatal(http.ListenAndServe(cfg.Host, nil))
}
