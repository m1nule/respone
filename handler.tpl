package {{.PkgName}}

import (
	"net/http"

	"github.com/m1nule/respone"
	"github.com/zeromicro/go-zero/rest/httpx"
	{{.ImportPackages}}
)

func {{.HandlerName}}(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		{{if .HasRequest}}var req types.{{.RequestType}}
		if err := httpx.Parse(r, &req); err != nil {
			respone.ResponseCtx(r.Context(), w, nil, respone.NewCodeError(respone.CodeParam, err.Error()))
			return
		}

		{{end}}l := {{.LogicName}}.New{{.LogicType}}(r.Context(), svcCtx)
		{{if .HasResp}}resp, {{end}}err := l.{{.Call}}({{if .HasRequest}}&req{{end}})
		{{if .HasResp}}respone.ResponseCtx(r.Context(), w, resp, err){{else}}respone.ResponseCtx(r.Context(), w, nil, err){{end}}
	}
}
