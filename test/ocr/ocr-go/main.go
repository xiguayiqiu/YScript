package main

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/golang-module/base64Captcha"
	"github.com/golang-module/base64Captcha/driver"
)

type memoryStore struct {
	mu   sync.Mutex
	data map[string]string
}

func (m *memoryStore) Put(key, value string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
}

func (m *memoryStore) Get(key string, use bool) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.data[key]
	if !ok {
		return ""
	}
	return v
}

func (m *memoryStore) Verify(key, answer string, remove bool) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.data[key]
	if !ok {
		return false
	}
	if remove {
		delete(m.data, key)
	}
	return v == answer
}

func (m *memoryStore) Delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
}

func (m *memoryStore) Set(key, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
	return nil
}

var base64Store = &memoryStore{data: make(map[string]string)}

type verifyReq struct {
	CaptchaID string `json:"captcha_id"`
	Code      string `json:"code"`
}

func generate(w http.ResponseWriter, r *http.Request) {
	captchaDriver := *driver.DefaultDriverString
	captchaDriver.Width = 200
	captchaDriver.Height = 70
	captchaDriver.Length = 5
	cp := base64Captcha.NewCaptcha(&captchaDriver, base64Store)

	id, b64, _, err := cp.Generate()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 去掉 "data:image/png;base64," 前缀
	parts := strings.SplitN(b64, ",", 2)
	imgBytes, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("X-Captcha-ID", id)
	w.Write(imgBytes)
}

func verify(w http.ResponseWriter, r *http.Request) {
	var req verifyReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ok := base64Store.Verify(req.CaptchaID, req.Code, true)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": ok})
}

func main() {
	http.HandleFunc("/captcha/generate", generate)
	http.HandleFunc("/captcha/verify", verify)
	addr := os.Getenv("OCR_TEST_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	log.Fatal(http.ListenAndServe(addr, nil))
}
