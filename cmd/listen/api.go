package main

import (
  "github.com/MuddCreates/hyperschedule-api-go/internal/api"
  "log"
  "net/http"
  "encoding/json"
)

func apiV3Handler(resp http.ResponseWriter, req *http.Request) {
  output, err := json.Marshal(api.MakeV3(state.GetData()))
  if err != nil {
    log.Printf("api: failed to jsonify, %s", err)
    resp.WriteHeader(http.StatusInternalServerError)
  }

  resp.Header().Add("Content-Type", "application/json")
  resp.Write(output)
}
