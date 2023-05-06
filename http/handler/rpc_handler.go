package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/qml-123/GateWay/model"
)

func RPCHandler(methods map[string]model.RPCMethod) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		method, ok := methods[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}

		if r.Method != method.HTTPMethod {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx := context.Background()
		request, err := method.NewRequest()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = json.NewDecoder(r.Body).Decode(request)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		response, err := method.Call(ctx, request)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
