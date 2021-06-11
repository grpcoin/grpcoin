package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (fe *frontend) apiPortfolioHistory(w http.ResponseWriter, req *http.Request) error {
	id, ok := mux.Vars(req)["id"]
	if !ok || id == "" {
		return status.Error(codes.InvalidArgument, "id is not specified")
	}

	log := loggerFrom(req.Context())
	cacheKey := fmt.Sprintf("pv_%s_%d", id, time.Now().Truncate(time.Hour).Unix())
	expiration := time.Now().Truncate(time.Hour).Add(time.Hour).Sub(time.Now())

	if b, err := fe.Redis.Get(req.Context(), cacheKey).Bytes(); err == nil {
		w.Write(b)
		return nil
	} else {
		if !errors.Is(err, redis.Nil) {
			w.WriteHeader(http.StatusInternalServerError)
			return fmt.Errorf("redis failure: %w", err)
		}
	}

	vals, err := fe.DB.UserValuationHistory(req.Context(), id)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return status.Error(codes.NotFound, "user not found")
		}
		return err
	}

	type resp [][2]interface{}
	var r resp
	for _, v := range vals {
		d := v.Date.Unix() * 1000 // TODO migrate to UnixMillis in go1.17
		c, _ := v.Value.F().Float64()
		r = append(r, [2]interface{}{d, c})
	}
	var jsonBuf bytes.Buffer
	mw := io.MultiWriter(w, &jsonBuf)
	if err := json.NewEncoder(mw).Encode(r); err != nil {
		return fmt.Errorf("failed to encode the response: %w", err)
	}
	if err := fe.Redis.Set(req.Context(), cacheKey, jsonBuf.Bytes(), expiration).Err(); err != nil {
		log.Warn("cache set failure", zap.Error(err))
	}
	return nil
}
