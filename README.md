# wikka [![Build Status](https://travis-ci.org/Mooxmirror/wikka.svg)](https://travis-ci.org/Mooxmirror/wikka)

Wikka is a wiki environment written in Go focused on speed and simplicity.
## Building
To build the binaries, you need to download the Go runtime from [golang.org](https://golang.org).
Then switch to your workspace and enter
```bash
go get github.com/mooxmirror/wikka
```
If you don't want to run it in the repository folder, you have to copy the files `config.json` and the `templates` folder to the same directory as the executable.
## Contribute
Just fork the project and start working on some nice stuff. If you need some inspiration, you should fine some in our Issues section.
## Configuration
Configuration of the server is done via the `config.json`. Here is an example:
```json
{
  "Title": "Your nice wiki",
  "Url": "http://my-wiki-url-here.com",
  "Articles": "articles/",
  "Templates": "templates/",
  "Host": "my-wiki-url-here.com:80",
  "Frontpage": "index"
}

```
The name of the `main.template`, `edit.template`, `view.template` and `error.template` are constant, they can't be changed. Every other template file can be changed though.
