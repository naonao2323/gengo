package template

import (
	"log"
	"os"
	"text/template"
)

func Execut(file []byte) error {
	text := "{{ . }}\n"

	tpl, err := template.New("").Parse(text)
	if err != nil {
		log.Fatal(err)
	}

	value := "hello world"

	if err := tpl.Execute(os.Stdout, value); err != nil {
		log.Fatal(err)
	}
	return nil
}
