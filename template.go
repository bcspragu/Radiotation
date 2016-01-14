package main

import "html/template"

type RadioTemplate struct {
	*template.Template
}

type allData map[string]interface{}

func (p *RadioTemplate) ExecuteTemplate(c Context, name string, data allData) error {
	data["Host"] = c.r.Host
	data["Dev"] = dev

	if err := p.Template.ExecuteTemplate(c.w, "header.html", data); err != nil {
		return err
	}

	if err := p.Template.ExecuteTemplate(c.w, name, data); err != nil {
		return err
	}

	if err := p.Template.ExecuteTemplate(c.w, "footer.html", data); err != nil {
		return err
	}

	return nil
}
