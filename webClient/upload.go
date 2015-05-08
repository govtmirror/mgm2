package webClient

import (
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/M-O-S-E-S/mgm/core"
	"github.com/gorilla/mux"
	"github.com/satori/go.uuid"
)

// UploadHandler accepts POST uploaded files for user jobs
func (hc HttpConnector) UploadHandler(w http.ResponseWriter, r *http.Request) {

	//only POST is recognized here
	if r.Method != "POST" {
		http.Error(w, "Invalid Request", http.StatusInternalServerError)
		return
	}

	//confirm that this is attached to a user session
	session, _ := hc.store.Get(r, "MGM")
	if session.Values["guid"] == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("Access Denied"))
		return
	}

	vars := mux.Vars(r)

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		hc.logger.Error("Error: ", err.Error())
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		hc.logger.Error("Error: ", err.Error())
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		hc.logger.Error("Error: ", err.Error())
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	hc.logger.Info("Read %v bytes from %v", len(data), id)

	hc.fileUploadChan <- core.FileUpload{
		JobID: id,
		User:  session.Values["guid"].(uuid.UUID),
		File:  data,
	}

	w.Write([]byte("OK"))
}
