package webClient

import (
  "io/ioutil"
  "net/http"
  "github.com/gorilla/mux"
)

func (hc HttpConnector) UploadHandler(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  uploadId := vars["id"]

  file, _, err := r.FormFile("file")
  if err != nil {
    hc.logger.Error("Error: %v", err.Error())
    http.Error(w, "Internal error", http.StatusInternalServerError)
  return
  }
  data, err := ioutil.ReadAll(file)
  if err != nil {
    hc.logger.Error("Error: %v", err.Error())
    http.Error(w, "Internal error", http.StatusInternalServerError)
    return
  }

  hc.logger.Info("Read %v bytes from %v", len(data), uploadId)


  http.Error(w, "something else", http.StatusInternalServerError)
  return
}
