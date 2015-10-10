package main

import (
	"log"
	"io/ioutil"
	"strings"
	"encoding/json"
	"os"
)

func LoadConfiguration(path string) {
	file, err := os.Open(path)
	if err != nil {
		log.Panic(err)
	}
	decoder := json.NewDecoder(file)
	cfg = new(Configuration)
	err = decoder.Decode(cfg)
	if err != nil {
		log.Panic(err)
	}
}

// load all articles from a specific path
func LoadArticles(path string) {
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
		}
	}

	log.Println(len(articles), "articles loaded")
}

// load all templates from a specific path
func LoadTemplates(path string) {
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
		}
	}

	log.Println(len(templates), "templates loaded")

	// pre-render templates
	for key, _ := range templates {
		templates[key] = RenderTemplate(key)
	}

	CreateContainerTemplate(CONTAINER_TEMPLATE)
}
