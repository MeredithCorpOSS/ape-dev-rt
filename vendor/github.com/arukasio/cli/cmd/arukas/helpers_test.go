package main

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
)

type Response struct {
	method, path, query, contentType, body string
	StatusCode                             int
}

var appStopped = `
{
	"id": "cf727ac6-f98d-4da6-aa05-6e7247daf5b8",
	"type": "apps",
	"attributes": {
		"name": "nginx",
		"image_id": "2b21fe34-328f-4d7e-8678-726d9eff2b7f",
		"created_at": "2015-10-19T15:05:34.722+09:00",
		"updated_at": "2015-10-20T10:53:46.169+09:00"
	},
	"relationships": {
		"user": {
			"data": {
				"id": "83c5a435-1e24-477f-9fb9-9c6c6fb29a64",
				"type": "users"
			}
		},
		"container": {
			"data": {
				"id": "2b21fe34-328f-4d7e-8678-726d9eff2b7f",
				"type": "containers"
			}
		}
	}
}
`

var appRunning = `
{
	"id": "97d464dd-e685-44c7-ae34-5e2b9025ec8f",
	"type": "apps",
	"attributes": {
		"name": "nostalgic_meitner9",
		"image_id": "d19b004c-0d59-4f4f-955c-5bace7c49a34",
		"created_at": "2015-12-21T19:48:17.181+09:00",
		"updated_at": "2015-12-21T19:48:17.181+09:00"
	},
	"relationships": {
		"user": {
			"data": {
				"id": "83c5a435-1e24-477f-9fb9-9c6c6fb29a64",
				"type": "users"
			}
		},
		"container": {
			"data": {
				"id": "d19b004c-0d59-4f4f-955c-5bace7c49a34",
				"type": "containers"
			}
		}
	}
}
`

var containerStopped = `
{
	"id":"2b21fe34-328f-4d7e-8678-726d9eff2b7f",
	"type":"containers",
	"attributes":{
		"app_id":"cf727ac6-f98d-4da6-aa05-6e7247daf5b8",
		"image_name":"nginx:latest",
		"cmd":null,
		"is_running":false,
		"instances":1,
		"mem":256,"envs":null,
		"name": "stopped-container",
		"end_point": "stopped-container.arukascloud.io",
		"ports":[
			{"protocol":"tcp","number":80},
			{"protocol":"tcp","number":8080}
		],
		"port_mappings":null,
		"created_at":"2015-10-19T15:05:34.843+09:00",
		"updated_at":"2015-12-08T14:13:35.152+09:00",
		"status_text":"interrupted"
	},
	"relationships":{
		"app":{
			"data":{
				"id":"cf727ac6-f98d-4da6-aa05-6e7247daf5b8",
				"type":"apps"
			}
		}
	}
}
`

var containerRunning = `
{
	"id":"d19b004c-0d59-4f4f-955c-5bace7c49a34",
	"type":"containers",
	"attributes":{
		"app_id":"97d464dd-e685-44c7-ae34-5e2b9025ec8f",
		"image_name":"nginx:latest",
		"cmd":"",
		"is_running":true,
		"instances":1,
		"mem":256,
		"name": "test-con1",
		"end_point": "test-con1.arukascloud.io",
		"envs":null,
		"ports":[
			{"protocol":"tcp","number":80}
		],
		"port_mappings":[
			[
				{
					"container_port":80,
					"service_port":31698,
					"host":"seaof-153-120-165-41.jp-ishikari-01.arukascloud.io"
				}
			]
		],
		"created_at":"2015-12-21T19:48:17.230+09:00",
		"updated_at":"2015-12-21T20:03:20.182+09:00",
		"status_text":"running"
	},
	"relationships":{
		"app":{
			"data":{
				"id":"97d464dd-e685-44c7-ae34-5e2b9025ec8f",
				"type":"apps"
			}
		}
	}
}
`

var responses = []Response{
	Response{
		method:      "POST",
		path:        "/app-sets",
		contentType: "application/vnd.api+json",
		body:        `{ "data": [` + appStopped + `, ` + containerStopped + ` ], "meta": { "name": "Deprecated Attribute" } }`,
		StatusCode:  http.StatusCreated,
	}, Response{
		method:      "GET",
		path:        "/apps/cf727ac6-f98d-4da6-aa05-6e7247daf5b8",
		contentType: "application/vnd.api+json",
		body:        `{ "data": ` + appStopped + `}`,
		StatusCode:  http.StatusOK,
	}, Response{
		method:      "DELETE",
		path:        "/apps/cf727ac6-f98d-4da6-aa05-6e7247daf5b8",
		contentType: "application/vnd.api+json",
		body:        "",
		StatusCode:  http.StatusNoContent,
	}, Response{
		method:      "GET",
		path:        "/apps/d19b004c-0d59-4f4f-955c-5bace7c49a34",
		contentType: "application/vnd.api+json",
		body:        `{ "data": ` + containerStopped + `}`,
		StatusCode:  http.StatusOK,
	}, Response{
		method:      "GET",
		path:        "/containers",
		contentType: "application/vnd.api+json",
		body: `{
			"data":[
				` + containerStopped + `,
				` + containerRunning + `
			]
		}`,
	}, Response{
		method:      "GET",
		path:        "/containers/2b21fe34-328f-4d7e-8678-726d9eff2b7f",
		contentType: "application/vnd.api+json",
		StatusCode:  http.StatusOK,
		body:        `{"data": ` + containerStopped + `}`,
	}, Response{
		method:      "GET",
		path:        "/containers/97d464dd-e685-44c7-ae34-5e2b9025ec8f",
		contentType: "application/vnd.api+json",
		StatusCode:  http.StatusOK,
		body:        `{"data": ` + containerRunning + `}`,
	}, Response{
		method:      "POST",
		path:        "/containers/2b21fe34-328f-4d7e-8678-726d9eff2b7f/power",
		contentType: "application/vnd.api+json",
		StatusCode:  http.StatusCreated,
	}, Response{
		method:      "POST",
		path:        "/containers/d19b004c-0d59-4f4f-955c-5bace7c49a34/power",
		contentType: "application/vnd.api+json",
		StatusCode:  422,
		body:        `{"message":"The app is not bootable."}`,
	}, Response{
		method:      "DELETE",
		path:        "/containers/d19b004c-0d59-4f4f-955c-5bace7c49a34/power",
		contentType: "application/vnd.api+json",
		StatusCode:  http.StatusAccepted,
	}, //Response{
	// 	method:      "POST",
	// 	path:        "/containers/d19b004c-0d59-4f4f-955c-5bace7c49a34/power",
	// 	contentType: "application/vnd.api+json",
	// 	StatusCode:  422,
	// 	body:        `{"message":"The app is not bootable."}`,
	// },
}

var handler = func(w http.ResponseWriter, r *http.Request) {
	// Check request.
	var (
		response    *Response
		body        string
		contentType string
	)
	for _, res := range responses {
		if r.URL.Path == res.path && r.Method == res.method {
			response = &res
			body = res.body
			contentType = res.contentType
			break
		}
	}
	if response == nil {
		log.Fatalf("No response date for %s, %s", r.Method, r.URL.Path)
		return
	}

	// Send response.
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(response.StatusCode)
	io.WriteString(w, body)
}

func runCommand(args []string) int {
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	os.Setenv("ARUKAS_JSON_API_URL", server.URL)
	return RunTest(args)
}

func RunTest2(args []string) int {
	return Run(args)
}
