package main_module

import "net/http"

func init() {
	// IMPORTANT: Keep the trailing slash
	http.HandleFunc("/api/blob/", apiHandler)
}
