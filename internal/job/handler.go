package job

import "net/http"

func listJobs(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("List jobs"))
}

func createJob(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Create job"))
}

func getJob(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Get job"))
}
