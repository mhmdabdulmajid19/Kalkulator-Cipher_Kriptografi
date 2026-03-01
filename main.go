package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"kripto/cipher"
)

type CipherRequest struct {
	Text      string     `json:"text"`
	Key       string     `json:"key"`
	Mode      string     `json:"mode"`
	A         int        `json:"a"`
	B         int        `json:"b"`
	KeyMatrix [][]int    `json:"key_matrix"`
	// Rotor fields
	StartRotor int      `json:"start_rotor"`
	RotorKeys  []string `json:"rotor_keys"`
}

type APIResponse struct {
	OK     bool     `json:"ok"`
	Output string   `json:"output,omitempty"`
	Steps  []string `json:"steps,omitempty"`
	Error  string   `json:"error,omitempty"`
}

func withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions { w.WriteHeader(http.StatusNoContent); return }
		next(w, r)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func parseBody(r *http.Request, dst *CipherRequest) error {
	return json.NewDecoder(r.Body).Decode(dst)
}

func handleVigenere(w http.ResponseWriter, r *http.Request) {
	var req CipherRequest
	if err := parseBody(r, &req); err != nil { writeJSON(w, 400, APIResponse{Error: "body tidak valid"}); return }
	res, err := cipher.Vigenere(req.Text, req.Key, req.Mode)
	if err != nil { writeJSON(w, 400, APIResponse{Error: err.Error()}); return }
	writeJSON(w, 200, APIResponse{OK: true, Output: res.Output, Steps: res.Steps})
}

func handleAffine(w http.ResponseWriter, r *http.Request) {
	var req CipherRequest
	if err := parseBody(r, &req); err != nil { writeJSON(w, 400, APIResponse{Error: "body tidak valid"}); return }
	res, err := cipher.Affine(req.Text, req.A, req.B, req.Mode)
	if err != nil { writeJSON(w, 400, APIResponse{Error: err.Error()}); return }
	writeJSON(w, 200, APIResponse{OK: true, Output: res.Output, Steps: res.Steps})
}

func handlePlayfair(w http.ResponseWriter, r *http.Request) {
	var req CipherRequest
	if err := parseBody(r, &req); err != nil { writeJSON(w, 400, APIResponse{Error: "body tidak valid"}); return }
	res, err := cipher.Playfair(req.Text, req.Key, req.Mode)
	if err != nil { writeJSON(w, 400, APIResponse{Error: err.Error()}); return }
	writeJSON(w, 200, APIResponse{OK: true, Output: res.Output, Steps: res.Steps})
}

func handleHill(w http.ResponseWriter, r *http.Request) {
	var req CipherRequest
	if err := parseBody(r, &req); err != nil { writeJSON(w, 400, APIResponse{Error: "body tidak valid"}); return }
	if len(req.KeyMatrix) == 0 { writeJSON(w, 400, APIResponse{Error: "key_matrix wajib diisi"}); return }
	res, err := cipher.Hill(req.Text, req.KeyMatrix, req.Mode)
	if err != nil { writeJSON(w, 400, APIResponse{Error: err.Error()}); return }
	writeJSON(w, 200, APIResponse{OK: true, Output: res.Output, Steps: res.Steps})
}

func handleRotor(w http.ResponseWriter, r *http.Request) {
	var req CipherRequest
	if err := parseBody(r, &req); err != nil { writeJSON(w, 400, APIResponse{Error: "body tidak valid"}); return }
	rotReq := cipher.RotorRequest{
		Text:       req.Text,
		Mode:       req.Mode,
		StartRotor: req.StartRotor,
		RotorKeys:  req.RotorKeys,
	}
	res, err := cipher.RotorCipher(rotReq)
	if err != nil { writeJSON(w, 400, APIResponse{Error: err.Error()}); return }
	writeJSON(w, 200, APIResponse{OK: true, Output: res.Output, Steps: res.Steps})
}

func main() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	http.HandleFunc("/api/vigenere", withCORS(handleVigenere))
	http.HandleFunc("/api/affine",   withCORS(handleAffine))
	http.HandleFunc("/api/playfair", withCORS(handlePlayfair))
	http.HandleFunc("/api/hill",     withCORS(handleHill))
	http.HandleFunc("/api/rotor",    withCORS(handleRotor))

	port := "8080"
	fmt.Printf("🔐 Server berjalan di http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}