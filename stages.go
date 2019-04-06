package makeweb

import (
	"bytes"
	"github.com/PaperMountainStudio/makeweb/plugins"
	"html/template"
	"log"
	"os"
	"path"
)

func stageLink() error {
	ok, err := exists("static")
	if err != nil {
		return err
	}
	if !ok {
		log.Println("WARNING: static directory not found")
		return nil
	}

	err = os.Chdir("static")
	if err != nil {
		return err
	}
	files, err := recursiveLs(".")
	if err != nil {
		return err
	}
	log.Println("Link:")
	for _, d := range files {
		outPath := path.Join("../output/", d)
		err := os.MkdirAll(path.Dir(outPath), os.ModePerm)
		if err != nil {
			return err
		}
		log.Println("- " + d)
		if _, err = os.Stat(outPath); !os.IsNotExist(err) {
			log.Println("already there")
			continue
		}
		err = os.Link(d, outPath)
		if err != nil {
			return err
		}
	}
	err = os.Chdir("..")
	if err != nil {
		return err
	}
	return nil
}

func stageRender(pages []Page, varsGlobal map[string]interface{}, templates *template.Template) error {
	for _, page := range pages {
		outPath := path.Join("output/", page.Path)
		outPath = plugins.EventModifyOutPath(outPath)
		err := os.MkdirAll(path.Dir(outPath), os.ModePerm)
		if err != nil {
			return err
		}

		log.Println("- " + page.Path)

		// convert .text to html if needed
		page, err = toHTML(page)
		if err != nil {
			return err
		}

		vars, err := joinmaps(varsGlobal, page.Vars)
		if err != nil {
			return err
		}

		temporaryVars := vars
		contentwriter := bytes.NewBufferString("")
		contentTemplate := template.New("default")
		contentTemplate.Parse(page.Content)
		temporaryVars["template"] = "default"
		render(contentwriter, contentTemplate, temporaryVars)
		page.Content = contentwriter.String()
		vars["text"] = template.HTML(page.Content)

		// write
		f, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY, 0644)
		render(f, templates, vars)
		if err != nil {
			return err
		}
		f.Close()
	}
	return nil
}