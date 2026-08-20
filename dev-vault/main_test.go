package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChallengeFetchAndReplay(t *testing.T) {
	handler := newHandler("owner/workload", map[string]string{"PRIVATE_MODEL_KEY": "secret", "OTHER": "hidden"})
	clientCert := &x509.Certificate{RawSubjectPublicKeyInfo: []byte("client-a")}

	challengeRequest := httptest.NewRequest(http.MethodPost, "/challenge", nil)
	challengeRequest.TLS = &tls.ConnectionState{PeerCertificates: []*x509.Certificate{clientCert}}
	challengeResponse := httptest.NewRecorder()
	handler.ServeHTTP(challengeResponse, challengeRequest)
	if challengeResponse.Code != http.StatusOK {
		t.Fatalf("challenge status = %d", challengeResponse.Code)
	}
	var issued map[string]string
	if err := json.Unmarshal(challengeResponse.Body.Bytes(), &issued); err != nil {
		t.Fatal(err)
	}

	body, _ := json.Marshal(fetchRequest{
		Repo:       "owner/workload",
		SecretRefs: []string{"PRIVATE_MODEL_KEY"},
		Nonce:      issued["nonce"],
		Document:   json.RawMessage(`{"development":"unverified"}`),
	})
	fetch := func() *httptest.ResponseRecorder {
		request := httptest.NewRequest(http.MethodPost, "/fetch", bytes.NewReader(body))
		request.TLS = &tls.ConnectionState{PeerCertificates: []*x509.Certificate{clientCert}}
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		return response
	}
	first := fetch()
	if first.Code != http.StatusOK || first.Body.String() != "{\"PRIVATE_MODEL_KEY\":\"secret\"}\n" {
		t.Fatalf("fetch = %d %q", first.Code, first.Body.String())
	}
	if replay := fetch(); replay.Code != http.StatusForbidden {
		t.Fatalf("replay status = %d", replay.Code)
	}
}

func TestChallengeBoundToClientCertificate(t *testing.T) {
	handler := newHandler("owner/workload", map[string]string{"KEY": "secret"})
	challengeRequest := httptest.NewRequest(http.MethodPost, "/challenge", nil)
	challengeRequest.TLS = &tls.ConnectionState{PeerCertificates: []*x509.Certificate{{RawSubjectPublicKeyInfo: []byte("client-a")}}}
	challengeResponse := httptest.NewRecorder()
	handler.ServeHTTP(challengeResponse, challengeRequest)
	var issued map[string]string
	if err := json.Unmarshal(challengeResponse.Body.Bytes(), &issued); err != nil {
		t.Fatal(err)
	}

	body, _ := json.Marshal(fetchRequest{
		Repo:       "owner/workload",
		SecretRefs: []string{"KEY"},
		Nonce:      issued["nonce"],
		Document:   json.RawMessage(`{}`),
	})
	request := httptest.NewRequest(http.MethodPost, "/fetch", bytes.NewReader(body))
	request.TLS = &tls.ConnectionState{PeerCertificates: []*x509.Certificate{{RawSubjectPublicKeyInfo: []byte("client-b")}}}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("cross-client status = %d", response.Code)
	}
}

func TestLoopbackAddress(t *testing.T) {
	for _, address := range []string{"127.0.0.1:8443", "[::1]:8443", "localhost:8443"} {
		if !loopbackAddress(address) {
			t.Fatalf("rejected %s", address)
		}
	}
	for _, address := range []string{":8443", "0.0.0.0:8443", "192.0.2.1:8443", "invalid"} {
		if loopbackAddress(address) {
			t.Fatalf("accepted %s", address)
		}
	}
}

func TestRestrictSource(t *testing.T) {
	_, allowed, err := net.ParseCIDR("192.0.2.0/24")
	if err != nil {
		t.Fatal(err)
	}
	handler := restrictSource(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}), allowed)
	for _, test := range []struct {
		remote string
		want   int
	}{
		{remote: "192.0.2.10:1234", want: http.StatusNoContent},
		{remote: "127.0.0.1:1234", want: http.StatusNoContent},
		{remote: "198.51.100.11:1234", want: http.StatusForbidden},
	} {
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request.RemoteAddr = test.remote
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != test.want {
			t.Fatalf("%s status = %d, want %d", test.remote, response.Code, test.want)
		}
	}
}
