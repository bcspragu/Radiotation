package main

func serveSearch(c Context) {
	tracks, err := c.Room.SongServer.Search(c.r.FormValue("search"))
	if err != nil {
		serveError(c.w, err)
		return
	}

	data := allData{
		"Host":   c.r.Host,
		"Tracks": tracks,
		"Queue":  c.Queue,
		"Room":   c.Room,
		"Raw":    true,
	}

	err = templates.ExecuteTemplate(c, "search.html", data)
	if err != nil {
		serveError(c.w, err)
	}
}
