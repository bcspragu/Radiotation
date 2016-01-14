package main

import "github.com/gopherjs/gopherjs/js"

func main() {
	js.Global.Set("Game", map[string]interface{}{
		"StateFromBlob": StateFromBlob,
	})
}

func StateFromBlob(jsonBlob string) *js.Object {
	return js.MakeWrapper(struct{}{})
}
