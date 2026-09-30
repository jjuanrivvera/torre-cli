// Command demoserver supplies invented API responses for an account-free VHS demo.
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	addr := "127.0.0.1:8644"
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}
	server := &http.Server{
		Addr:              addr,
		Handler:           http.HandlerFunc(serveDemo),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}
	log.Println("invented demo API listening")
	log.Fatal(server.ListenAndServe())
}

func serveDemo(w http.ResponseWriter, r *http.Request) {
	var result any
	switch r.Method + " " + r.URL.Path {
	case "POST /opportunities/_search/":
		result = map[string]any{"total": 2, "size": 2, "offset": 0, "results": []any{
			map[string]any{"id": "demo-go-01", "objective": "Go Platform Engineer", "opportunity": "employment", "remote": true, "status": "open", "locations": []string{"Worldwide"}, "place": map[string]any{"locationType": "remote_anywhere"}},
			map[string]any{"id": "demo-go-02", "objective": "Go Backend Engineer", "opportunity": "employment", "remote": true, "status": "open", "locations": []string{"Worldwide"}, "place": map[string]any{"locationType": "remote_anywhere"}},
		}}
	case "GET /suite/opportunities/demo-go-01":
		result = map[string]any{"id": "demo-go-01", "objective": "Go Platform Engineer", "status": "open", "organizations": []any{map[string]any{"name": "Demo Orbit Workshop"}}, "compensation": map[string]any{"currency": "USD$", "minAmount": 9000, "maxAmount": 12000, "periodicity": "monthly"}, "place": map[string]any{"locationType": "remote_anywhere"}}
	case "POST /people/_search":
		result = map[string]any{"total": 2, "size": 2, "offset": 0, "results": []any{
			map[string]any{"ggId": "demo-person-01", "name": "Demo Candidate A", "username": "demo_candidate_a", "professionalHeadline": "Go engineer building reliable APIs", "verified": false},
			map[string]any{"ggId": "demo-person-02", "name": "Demo Candidate B", "username": "demo_candidate_b", "professionalHeadline": "Platform engineer and tooling builder", "verified": false},
		}}
	case "GET /genome/bios/demo_candidate_a":
		result = map[string]any{"person": map[string]any{"name": "Demo Candidate A", "publicId": "demo_candidate_a", "professionalHeadline": "Go engineer building reliable APIs"}, "strengths": []any{map[string]any{"name": "Go", "proficiency": "expert"}, map[string]any{"name": "API design", "proficiency": "proficient"}}}
	default:
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Print(err)
	}
}
