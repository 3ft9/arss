package main

import (
	"github.com/parnurzeal/gorequest"
)

func Stats(apiKey string, projectId string, collection string, data interface{}) {
	request := gorequest.New()
	resp, body, errs := request.Post("https://api.getconnect.io/events/"+collection).
		Set("X-API-Key", apiKey).
		Set("X-Project-Id", projectId).
		Send(data).
		End()
	if nil != errs {
		println(errs)
	} else {
		if resp.StatusCode != 200 {
			println("Connect error:")
			println(resp.StatusCode)
			println(body)
			println()
			println()
		}
	}
}
