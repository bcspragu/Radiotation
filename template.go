package main

import (
	"html/template"
	"room"
)

type RadioTemplate struct {
	*template.Template
}

type allData map[string]interface{}

func (p *RadioTemplate) ExecuteTemplate(c Context, name string, data allData) error {
	raw, _ := data["Raw"].(bool)

	data["Host"] = c.r.Host
	data["Dev"] = dev
	if c.Room != nil {
		data["Room"] = c.Room
	}

	if data["Room"] == nil {
		data["Room"] = &room.Room{}
	}

	if !raw {
		if err := p.Template.ExecuteTemplate(c.w, "header.html", data); err != nil {
			return err
		}
	}

	if err := p.Template.ExecuteTemplate(c.w, name, data); err != nil {
		return err
	}

	if !raw {
		if err := p.Template.ExecuteTemplate(c.w, "footer.html", data); err != nil {
			return err
		}
	}

	return nil
}
