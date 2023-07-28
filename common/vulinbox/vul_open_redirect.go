package vulinbox

import (
	"net/http"
	"net/url"
	"strings"
)

func (s *VulinServer) registerOpenRedirect() {
	urlRedirectGroup := s.router.PathPrefix("/redirect").Name("URL重定向漏洞").Subrouter()
	urlRedirectRoutes := []*VulInfo{
		{
			DefaultQuery: "url=a",
			Path:         "/safe",
			Title:        "安全的url重定向",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				var u = request.URL.Query().Get("url")
				if u == "a" {
					writer.Write([]byte("修改url参数跳转页面"))
					writer.WriteHeader(http.StatusInternalServerError)
					return
				}
				urlStruct, err := url.Parse(u)
				if err != nil {
					writer.Write([]byte(err.Error()))
					writer.WriteHeader(http.StatusInternalServerError)
					return
				}

				if urlStruct.Host != request.Host {
					writer.Write([]byte("不安全的重定向"))
					writer.WriteHeader(http.StatusInternalServerError)
					return
				}

				writer.Header().Set("Location", u)
				writer.WriteHeader(http.StatusFound)
				return
			},
		},
		{
			DefaultQuery: "url=a",
			Path:         "/include-white-list",
			Title:        "只匹配关键字的url重定向",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				var u = request.URL.Query().Get("url")
				if u == "a" {
					writer.Write([]byte("修改url参数跳转页面"))
					writer.WriteHeader(http.StatusOK)
					return
				}

				if !strings.Contains(u, request.Host) {
					writer.Write([]byte("不安全的重定向"))
					writer.WriteHeader(http.StatusInternalServerError)
					return
				}

				writer.Header().Set("Location", u)
				writer.WriteHeader(http.StatusFound)
				return
			},
		},
		{
			DefaultQuery: "url=a",
			Path:         "/no_protect",
			Title:        "无保护的url重定向",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				var u = request.URL.Query().Get("url")
				if u == "a" {
					writer.Write([]byte("修改url参数跳转页面"))
					writer.WriteHeader(http.StatusOK)
					return
				}

				writer.Header().Set("Location", u)
				writer.WriteHeader(http.StatusFound)
				return
			},
		},
	}
	for _, v := range urlRedirectRoutes {
		addRouteWithVulInfo(urlRedirectGroup, v)
	}
}
