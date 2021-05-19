package frontend

import (
	_ "embed"
	"net/http"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (fe *Frontend) userProfile(w http.ResponseWriter, r *http.Request) error {
	uid := mux.Vars(r)["id"]
	if uid == "" {
		return status.Error(codes.InvalidArgument, "url does not have user id")
	}
	if s := trace.SpanFromContext(r.Context()); s != nil {
		s.SetAttributes(attribute.String("user.id", uid))
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
	quotes, err := fe.getQuotes(r.Context())
	if err != nil {
		return err
	}
	return tpl.Funcs(funcs).ExecuteTemplate(w, "profile.tpl", map[string]interface{}{
		"u":      u,
		"orders": orders,
		"quotes": quotes,
	})
}
