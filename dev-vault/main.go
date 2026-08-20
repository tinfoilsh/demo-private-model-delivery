// Command dev-vault exercises the vault wire protocol without verifying attestation.
// It is intentionally unsuitable for production secret release.
package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	nonceSize     = 32
	challengeTTL  = 5 * time.Minute
	maxFetchBody  = 32 << 20
	maxSecretRefs = 256
)

type fetchRequest struct {
	Repo       string          `json:"repo"`
	SecretRefs []string        `json:"secret_refs"`
	Nonce      string          `json:"nonce"`
	Document   json.RawMessage `json:"document"`
}

type challenge struct {
	fingerprint string
	expiresAt   time.Time
}

type challengeStore struct {
	sync.Mutex
	entries map[string]challenge
}

func main() {
	addr := flag.String("addr", "127.0.0.1:8443", "HTTPS listen address")
	repo := flag.String("repo", "", "only repository allowed to request secrets")
	secretsPath := flag.String("secrets", "", "flat JSON secret map")
	tlsCert := flag.String("tls-cert", "", "TLS certificate chain PEM")
	tlsKey := flag.String("tls-key", "", "TLS private key PEM")
	unsafeSkip := flag.Bool("unsafe-skip-attestation", false, "required acknowledgement that attestation is not verified")
	unsafePublic := flag.Bool("unsafe-allow-non-loopback", false, "allow exposure beyond the loopback interface")
	allowedSource := flag.String("allow-source-cidr", "", "only non-loopback client CIDR allowed when exposed")
	flag.Parse()

	if !*unsafeSkip {
		log.Fatal("refusing to start without -unsafe-skip-attestation")
	}
	if *repo == "" || *secretsPath == "" || *tlsCert == "" || *tlsKey == "" {
		log.Fatal("-repo, -secrets, -tls-cert, and -tls-key are required")
	}
	if !*unsafePublic && !loopbackAddress(*addr) {
		log.Fatal("refusing non-loopback bind without -unsafe-allow-non-loopback")
	}
	var allowedNetwork *net.IPNet
	if !loopbackAddress(*addr) {
		_, parsedNetwork, parseErr := net.ParseCIDR(*allowedSource)
		if parseErr != nil {
			log.Fatal("non-loopback bind requires a valid -allow-source-cidr")
		}
		allowedNetwork = parsedNetwork
	}
	secrets, err := loadSecrets(*secretsPath)
	if err != nil {
		log.Fatalf("load secrets: %v", err)
	}

	server := &http.Server{
		Addr:              *addr,
		Handler:           restrictSource(newHandler(*repo, secrets), allowedNetwork),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      2 * time.Minute,
		IdleTimeout:       2 * time.Minute,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS13,
			ClientAuth: tls.RequestClientCert,
		},
	}
	log.Printf("UNSAFE: serving %d secret(s) without attestation verification on %s", len(secrets), *addr)
	log.Fatal(server.ListenAndServeTLS(*tlsCert, *tlsKey))
}

func newHandler(repo string, secrets map[string]string) http.Handler {
	challenges := &challengeStore{entries: make(map[string]challenge)}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.HandleFunc("POST /challenge", func(w http.ResponseWriter, r *http.Request) {
		fingerprint, err := clientFingerprint(r)
		if err != nil {
			http.Error(w, "client certificate required", http.StatusForbidden)
			return
		}
		nonce, err := challenges.issue(fingerprint)
		if err != nil {
			http.Error(w, "challenge unavailable", http.StatusServiceUnavailable)
			return
		}
		writeJSON(w, map[string]string{"nonce": nonce})
	})
	mux.HandleFunc("POST /fetch", func(w http.ResponseWriter, r *http.Request) {
		var request fetchRequest
		if err := decodeJSON(r.Body, maxFetchBody, &request); err != nil || request.Repo != repo || len(request.Document) == 0 {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		fingerprint, err := clientFingerprint(r)
		if err != nil || !challenges.consume(request.Nonce, fingerprint) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		released, ok := selectSecrets(secrets, request.SecretRefs)
		if !ok {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		log.Printf("UNSAFE: released %d secret(s) for %s", len(released), repo)
		writeJSON(w, released)
	})
	return mux
}

func (s *challengeStore) issue(fingerprint string) (string, error) {
	nonce := make([]byte, nonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	encoded := hex.EncodeToString(nonce)
	s.Lock()
	defer s.Unlock()
	now := time.Now()
	for value, entry := range s.entries {
		if !now.Before(entry.expiresAt) {
			delete(s.entries, value)
		}
	}
	if len(s.entries) >= 1024 {
		return "", errors.New("challenge capacity reached")
	}
	s.entries[encoded] = challenge{fingerprint: fingerprint, expiresAt: now.Add(challengeTTL)}
	return encoded, nil
}

func (s *challengeStore) consume(encoded, fingerprint string) bool {
	if encoded != strings.ToLower(encoded) {
		return false
	}
	nonce, err := hex.DecodeString(encoded)
	if err != nil || len(nonce) != nonceSize {
		return false
	}
	s.Lock()
	entry, found := s.entries[encoded]
	delete(s.entries, encoded)
	s.Unlock()
	return found && time.Now().Before(entry.expiresAt) && entry.fingerprint == fingerprint
}

func clientFingerprint(r *http.Request) (string, error) {
	if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
		return "", errors.New("missing client certificate")
	}
	return certificateFingerprint(r.TLS.PeerCertificates[0]), nil
}

func certificateFingerprint(cert *x509.Certificate) string {
	digest := sha256.Sum256(cert.RawSubjectPublicKeyInfo)
	return hex.EncodeToString(digest[:])
}

func selectSecrets(all map[string]string, names []string) (map[string]string, bool) {
	if len(names) == 0 || len(names) > maxSecretRefs {
		return nil, false
	}
	selected := make(map[string]string, len(names))
	for _, name := range names {
		if name == "" || len(name) > 256 {
			return nil, false
		}
		value, ok := all[name]
		if !ok {
			return nil, false
		}
		if _, duplicate := selected[name]; duplicate {
			return nil, false
		}
		selected[name] = value
	}
	return selected, true
}

func loadSecrets(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var secrets map[string]string
	if err := decodeJSON(file, 8<<20, &secrets); err != nil {
		return nil, err
	}
	if len(secrets) == 0 {
		return nil, errors.New("secret map is empty")
	}
	for name, value := range secrets {
		if name == "" || value == "" || value == "null" {
			return nil, fmt.Errorf("invalid secret %q", name)
		}
	}
	return secrets, nil
}

func decodeJSON(reader io.Reader, limit int64, target any) error {
	data, err := io.ReadAll(io.LimitReader(reader, limit+1))
	if err != nil {
		return err
	}
	if int64(len(data)) > limit {
		return errors.New("request too large")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return errors.New("trailing JSON data")
	}
	return nil
}

func loopbackAddress(addr string) bool {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return false
	}
	return host == "localhost" || net.ParseIP(host).IsLoopback()
}

func restrictSource(next http.Handler, allowed *net.IPNet) http.Handler {
	if allowed == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		remote := net.ParseIP(host)
		if err != nil || remote == nil || (!remote.IsLoopback() && !allowed.Contains(remote)) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("write response: %v", err)
	}
}
