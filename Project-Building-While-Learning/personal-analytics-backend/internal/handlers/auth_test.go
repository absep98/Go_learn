package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestValidRegisteration(t *testing.T) {

	body := strings.NewReader(`{"email":"aks@gmail.com","password":"Juniper@20202"}`)

	req := httptest.NewRequest(http.MethodPost, "/register", body)
	rec := httptest.NewRecorder()

	Register(rec, req)

	var resp RegisterResponse
	json.NewDecoder(rec.Body).Decode(&resp)

	if !resp.Success {
		t.Errorf("expected succes true, got false")
	}

	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}

}

func TestRegisteration_EmailMissing(t *testing.T) {

	body := strings.NewReader(`{"email":"","password":"Juniper@20202"}`)

	req := httptest.NewRequest(http.MethodPost, "/register", body)
	rec := httptest.NewRecorder()

	Register(rec, req)

	var resp RegisterResponse
	json.NewDecoder(rec.Body).Decode(&resp)

	if resp.Success == true {
		t.Errorf("expected succes true, got false")
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}

}

func TestRegisteration_ShortPassword(t *testing.T) {

	body := strings.NewReader(`{"email":"test@gmail.com","password":"Junip"}`)

	req := httptest.NewRequest(http.MethodPost, "/register", body)
	rec := httptest.NewRecorder()

	Register(rec, req)

	var resp RegisterResponse
	json.NewDecoder(rec.Body).Decode(&resp)

	if resp.Success == true {
		t.Errorf("expected succes true, got false")
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}

}

func TestValidRegisterationDuplicate(t *testing.T) {

	body := strings.NewReader(`{"email":"aks@gmail.com","password":"Juniper@20202"}`)

	req := httptest.NewRequest(http.MethodPost, "/register", body)
	rec := httptest.NewRecorder()

	Register(rec, req)

	var resp RegisterResponse
	json.NewDecoder(rec.Body).Decode(&resp)

	if resp.Success {
		t.Errorf("expected succes false for duplicate email")
	}

	if rec.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", rec.Code)
	}

}

func TestValidLogin(t *testing.T) {

	body := strings.NewReader(`{"email":"aks1@gmail.com","password":"Dynasty@Jan2020"}`)

	req := httptest.NewRequest(http.MethodPost, "/register", body)
	rec := httptest.NewRecorder()

	Register(rec, req)

	var resp RegisterResponse
	json.NewDecoder(rec.Body).Decode(&resp)

	if !resp.Success {
		t.Errorf("expected success true, got false")
	}

	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}

	body = strings.NewReader(`{"email":"aks1@gmail.com","password":"Dynasty@Jan2020"}`)

	req = httptest.NewRequest(http.MethodPost, "/login", body)
	rec = httptest.NewRecorder()

	os.Setenv("JWT_SECRET", "test-secret")

	Login(rec, req)

	var loginResp LoginResponse
	json.NewDecoder(rec.Body).Decode(&loginResp)

	if !loginResp.Success {
		t.Errorf("expected success true, got false")
	}

	if loginResp.Token == "" {
		t.Errorf("expected token didn't get")
	}

}

func TestLogin_WrongPassword(t *testing.T) {
	// Register a fresh user for this test (unique email to avoid conflict)
	body := strings.NewReader(`{"email":"wrongpass@gmail.com","password":"Dynasty@Jan2020"}`)
	req := httptest.NewRequest(http.MethodPost, "/register", body)
	rec := httptest.NewRecorder()
	Register(rec, req)

	var resp RegisterResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if !resp.Success {
		t.Fatalf("setup failed: could not register user: %s", resp.Message)
	}

	// Now login with wrong password
	body = strings.NewReader(`{"email":"wrongpass@gmail.com","password":"WrongPassword"}`)
	req = httptest.NewRequest(http.MethodPost, "/login", body)
	rec = httptest.NewRecorder()
	os.Setenv("JWT_SECRET", "test-secret")
	Login(rec, req)

	var loginResp LoginResponse
	json.NewDecoder(rec.Body).Decode(&loginResp)

	if loginResp.Success {
		t.Errorf("expected success false for wrong password")
	}
	if loginResp.Token != "" {
		t.Errorf("expected empty token, got one")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestLogin_NonExistentEmail(t *testing.T) {
	// No registration — email never exists in DB
	body := strings.NewReader(`{"email":"ghost_user@gmail.com","password":"somepassword"}`)
	req := httptest.NewRequest(http.MethodPost, "/login", body)
	rec := httptest.NewRecorder()
	os.Setenv("JWT_SECRET", "test-secret")
	Login(rec, req)

	var loginResp LoginResponse
	json.NewDecoder(rec.Body).Decode(&loginResp)

	if loginResp.Success {
		t.Errorf("expected success false for non-existent email")
	}
	if loginResp.Token != "" {
		t.Errorf("expected empty token, got one")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}
