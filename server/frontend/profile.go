package frontend

import (
	"net/http"
	"text/template"

	"github.com/gorilla/mux"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (fe *Frontend) userProfile(w http.ResponseWriter, r *http.Request) error {
	uid := mux.Vars(r)["id"]
	if uid == "" {
		return status.Error(codes.InvalidArgument, "url does not have user id")
	}
	u, ok, err := fe.DB.Get(r.Context(), uid)
	if err != nil {
		return err
	} else if !ok {
		return status.Error(codes.NotFound, "user not found")
	}

	orders, err := fe.DB.UserOrderHistory(r.Context(), uid)
	if err != nil {
		return err
	}
	for i := 0; i < len(orders)/2; i++ {
		orders[i], orders[len(orders)-1-i] = orders[len(orders)-1-i], orders[i]
	}
	tpl := `ID: {{.u.ID}}
Profile {{.u.ProfileURL}}
Sign up date {{.u.CreatedAt}}

{{ with .orders }}
ORDER HISTORY ({{ len .}})
=============
{{- range . }}
{{ .Date }} -- {{ .Action }} '{{ .Ticker }}' -- {{ rp .Size }} @ ${{ rp .Price }}
{{- end }}
{{ end }}
`

	t, err := template.New("").Funcs(funcs).Parse(tpl)
	if err != nil {
		return err
	}
	return t.Execute(w, map[string]interface{}{
		"u":      u,
		"orders": orders,
	})
}
