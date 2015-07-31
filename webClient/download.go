package webClient

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/m-o-s-e-s/mgm/mgm"
	"github.com/satori/go.uuid"
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
	ip := strings.Split(r.RemoteAddr, ":")[0]

	hdrRealIp := r.Header.Get("X-Real-Ip")
	//hdrForwardedFor := r.Header().Get("X-Forwarded-For")
	if hdrRealIp != "" {
		ip = hdrRealIp
	}

	hc.logger.Info("download file requested from %v", ip)

	//test if connecting party is a host
	isHost := false
	for _, h := range hc.mgm.GetHosts() {
		if h.Address == ip {
			isHost = true
			hc.logger.Info("%v is a host", ip)
		}
	}

	//if not a host, check for a client session
	var uid uuid.UUID
	if !isHost {
		s, _ := hc.store.Get(r, "MGM")
		if s.Values["guid"] == nil {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("Access Denied"))
			hc.logger.Info("Access denied to non host non authenticated user")
			return
		}
		hc.logger.Info("%v is client %v", ip, s.Values["guid"].(uuid.UUID).String())
		uid = s.Values["guid"].(uuid.UUID)
	}

	vars := mux.Vars(r)

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		hc.logger.Error(fmt.Sprintf("Error, job id missing: %v", err.Error()))
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	job, found := hc.jMgr.GetJobByID(int64(id))

	if !found {
		hc.logger.Error("Error: Job not found")
		http.Error(w, "Job not found", http.StatusInternalServerError)
		return
	}

	if !isHost && job.User != uid {
		hc.logger.Error("Error: Permission denied")
		http.Error(w, "PermissionDenied", http.StatusInternalServerError)
		return
	}

	var jd mgm.JobData

	switch job.Type {
	case "save_oar":
		hc.logger.Info("Parsing save oar")
		jd = job.ReadData()
		if jd.Status != "Done" || jd.File != "" || jd.Filename != "" {
			hc.logger.Error("Error: Save oar is not complete, or an error occurred")
			http.Error(w, "Job Error", http.StatusInternalServerError)
			return
		}
	case "save_iar":
		hc.logger.Info("PArsing save iar")
		jd = job.ReadData()
		if jd.Status != "Done" || jd.File != "" || jd.Filename != "" {
			hc.logger.Error("Error: Save oar is not complete, or an error occurred")
			http.Error(w, "Job Error", http.StatusInternalServerError)
			return
		}
	case "load_oar":
		hc.logger.Info("processing load oar")
		jd = job.ReadData()
	case "load_iar":
		hc.logger.Info("processing load iar")
		jd = job.ReadData()
	default:
		hc.logger.Error("Error: Invalid job for download")
		http.Error(w, "Invalid job type", http.StatusInternalServerError)
		return
	}

	//serve file
	f, err := os.Open(jd.File)
	if err != nil {
		hc.logger.Error(fmt.Sprintf("Error on file %v: %v", jd.Filename, err.Error()))
		http.Error(w, "Error opening file", http.StatusInternalServerError)
		return
	}
	statinfo, _ := f.Stat()
	defer f.Close()

	w.Header().Set("Content-Description", "File Transfer")
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%v", jd.Filename))
	w.Header().Set("Content-Transfer-Encoding", "binary")
	w.Header().Set("Expires", "0")
	w.Header().Set("Cache-Control", "must-revalidate")
	w.Header().Set("Pragma", "public")
	w.Header().Set("Content-Length", strconv.FormatInt(statinfo.Size(), 10))

	hc.logger.Info("Serving %v bytes", strconv.FormatInt(statinfo.Size(), 10))

	io.Copy(w, f)

	hc.logger.Info("Serve download complete")

	//outputWriter := w.(io.Writer)
	//
	//	buf := make([]byte, min(fsMaxbufsize, statinfo.Size()))
	//	n := 0
	//	for err == nil {
	//		n, err = f.Read(buf)
	//		outputWriter.Write(buf[0:n])
	//	}
}
