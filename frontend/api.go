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
)

func (fe *frontend) apiPortfolioHistory(w http.ResponseWriter, req *http.Request) {
	id, ok := mux.Vars(req)["id"]
	if !ok || id == "" {
		w.WriteHeader(http.StatusBadRequest)
		// TODO return proper json error and handle in the client
		return
	}

	log := loggerFrom(req.Context())
	cacheKey := fmt.Sprintf("pv_%s_%d", id, time.Now().Truncate(time.Hour).Unix())
	expiration := time.Now().Truncate(time.Hour).Add(time.Hour).Sub(time.Now())

	if b, err := fe.Redis.Get(req.Context(), cacheKey).Bytes(); err == nil {
		w.Write(b)
		return
	} else {
		if !errors.Is(err, redis.Nil) {
			log.Error("redis failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError) // TODO return proper json error
			return
		}
	}

	vals, err := fe.DB.UserValuationHistory(req.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		// TODO return proper json error and handle in the client
		return
	}

	type resp [][2]interface{}
	var r resp
	for _, v := range vals {
		d := v.Date.Unix() * 1000
		c, _ := v.Value.F().Float64()
		r = append(r, [2]interface{}{d, c})
	}
	var jsonBuf bytes.Buffer
	mw := io.MultiWriter(w, &jsonBuf)
	_ = json.NewEncoder(mw).Encode(r) // TODO handle err

	if err := fe.Redis.Set(req.Context(), cacheKey, jsonBuf.Bytes(), expiration).Err(); err != nil {
		log.Warn("cache set fail", zap.Error(err))
	}
}
