package vulinbox

import (
	"net/http"
	"net/url"
	"strings"
)

func (s *VulinServer) registerOpenRedirect() {
	var router = s.router
	router.HandleFunc("/redirect/safe", func(writer http.ResponseWriter, request *http.Request) {
		var u = request.URL.Query().Get("url")
		if u == "a" {
			writer.Write([]byte("修改url参数跳转页面"))
			writer.WriteHeader(200)
			return
		}
		urlStruct, err := url.Parse(u)
		if err != nil {
			writer.Write([]byte(err.Error()))
			writer.WriteHeader(500)
			return
		}

		if urlStruct.Host != request.Host {
			writer.Write([]byte("不安全的重定向"))
			writer.WriteHeader(500)
			return
		}

		writer.Header().Set("Location", u)
		writer.WriteHeader(302)
		return
	})

	router.HandleFunc("/redirect/include-white-list", func(writer http.ResponseWriter, request *http.Request) {
		var u = request.URL.Query().Get("url")
		if u == "a" {
			writer.Write([]byte("修改url参数跳转页面"))
			writer.WriteHeader(200)
			return
		}

		if !strings.Contains(u, request.Host) {
			writer.Write([]byte("不安全的重定向"))
			writer.WriteHeader(500)
			return
		}

		writer.Header().Set("Location", u)
		writer.WriteHeader(302)
		return
	})

	router.HandleFunc("/redirect/no_protect", func(writer http.ResponseWriter, request *http.Request) {
		var u = request.URL.Query().Get("url")
		if u == "a" {
			writer.Write([]byte("修改url参数跳转页面"))
			writer.WriteHeader(200)
			return
		}

		writer.Header().Set("Location", u)
		writer.WriteHeader(302)
		return
	})
}
