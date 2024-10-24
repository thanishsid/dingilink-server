package asset

import (
	"embed"
	"html/template"
	"log"
)

//go:embed email-templates/*.html
var files embed.FS

var MailTemplates *template.Template

func init() {
	tmpl, err := template.ParseFS(files, "email-templates/*.html")
	if err != nil {
		log.Fatal(err)
	}
	MailTemplates = tmpl
}

func GetFs() embed.FS {
	return files
}
