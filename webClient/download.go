package webClient

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/m-o-s-e-s/mgm/mgm"
)

const fsMaxbufsize = 4096

func min(x int64, y int64) int64 {
	if x < y {
		return x
	}
	return y
}

// UploadHandler accepts POST uploaded files for user jobs
func (hc httpConn) DownloadHandler(w http.ResponseWriter, r *http.Request) {

	hc.logger.Info("download file requested")

	//confirm that this is attached to a user session
	s, _ := hc.store.Get(r, "MGM")
	if s.Values["guid"] == nil {
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

	job, found := hc.jMgr.GetJobByID(int64(id))

	if !found {
		hc.logger.Error("Error: Job not found")
		http.Error(w, "Job not found", http.StatusInternalServerError)
		return
	}

	var jd mgm.JobData

	switch job.Type {
	case "save_oar":
		jd = job.ReadData()
		if jd.Status != "Done" || jd.File != "" || jd.Name != "" {
			hc.logger.Error("Error: Save oar is not complete, or an error occurred")
			http.Error(w, "Job Error", http.StatusInternalServerError)
			return
		}
	case "save_iar":
		jd = job.ReadData()
		if jd.Status != "Done" || jd.File != "" || jd.Name != "" {
			hc.logger.Error("Error: Save oar is not complete, or an error occurred")
			http.Error(w, "Job Error", http.StatusInternalServerError)
			return
		}
	default:
		hc.logger.Error("Error: Invalid job for download")
		http.Error(w, "Invalid job type", http.StatusInternalServerError)
		return
	}

	//serve file
	f, _ := os.Open(jd.File)
	statinfo, _ := f.Stat()
	defer f.Close()

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%v", jd.Name))

	hc.logger.Info("Serving %v bytes", strconv.FormatInt(statinfo.Size(), 10))

	outputWriter := w.(io.Writer)

	buf := make([]byte, min(fsMaxbufsize, statinfo.Size()))
	n := 0
	for err == nil {
		n, err = f.Read(buf)
		outputWriter.Write(buf[0:n])
	}
}
