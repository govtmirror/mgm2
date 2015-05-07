package webClient

import (
  "io/ioutil"
  "net/http"
  "strconv"
  "github.com/gorilla/mux"
  "github.com/M-O-S-E-S/mgm/core"
)

func (hc HttpConnector) UploadHandler(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  uploadId := vars["id"]

  id, err := strconv.Atoi(uploadId)
  if err != nil {
    hc.logger.Error("Error: %v", err.Error())
    http.Error(w, "Internal error", http.StatusInternalServerError)
    return
  }

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

  hc.logger.Info("Read %v bytes from %v", len(data), id)

  hc.fileUploadChan <- core.FileUpload{id, data}

  w.Write([]byte("OK"));
}
