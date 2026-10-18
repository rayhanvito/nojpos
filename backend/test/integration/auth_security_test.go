package integration

// Unit A — keamanan auth di level HTTP:
//   - SendOTP tanpa purpose berfungsi untuk pendaftaran.
//   - Purpose "login" (warisan) tidak lagi punya cabang khusus: permintaan
//     diperlakukan persis seperti pendaftaran.
//   - Login murni email + kata sandi; field warisan passcode/otp yang ikut
//     terkirim diabaikan (bukan ditantang, bukan pula jalan pintas).

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/rayhanvito/nojpos/backend/internal/app"
	"github.com/rayhanvito/nojpos/backend/internal/auth"
)

func TestAuthSecurityLoginDanOTP(t *testing.T) {
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DATABASE_URL", filepath.Join(t.TempDir(), "auth_security.db"))
	t.Setenv("JWT_SECRET", "auth-security-secret")

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	srv, err := app.New(ctx)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if srv.Cleanup != nil {
		defer srv.Cleanup()
	}

	doReq := func(method, path string, body any, token string) (*httptest.ResponseRecorder, map[string]any) {
		var bodyBuf []byte
		if body != nil {
			bodyBuf, _ = json.Marshal(body)
		}
		req := httptest.NewRequest(method, "/api/v1"+path, bytes.NewReader(bodyBuf))
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		srv.Handler.ServeHTTP(w, req)
		var respMap map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &respMap)
		return w, respMap
	}

	suffix := time.Now().Format("20060102150405")
	email := "unit_a_" + suffix + "@tokosaya.com"
	emailPurpose := "unit_a_purpose_" + suffix + "@tokosaya.com"
	codes := map[string]string{}
	auth.TestOnOTPSent = func(e, c string) { codes[e] = c }
	defer func() { auth.TestOnOTPSent = nil }()

	// 1) SendOTP tanpa purpose → alur pendaftaran: 200 dan kode tertangkap.
	w, _ := doReq("POST", "/auth/otp/send", map[string]string{"email": email}, "")
	if w.Code != http.StatusOK {
		t.Fatalf("SendOTP tanpa purpose: mau 200, dapat %d, body=%s", w.Code, w.Body.String())
	}
	if codes[email] == "" {
		t.Fatalf("SendOTP tanpa purpose: kode OTP tidak tertangkap")
	}

	// 2) SendOTP dengan purpose "login" untuk email baru: cabang login sudah
	// dihapus, jadi diperlakukan sebagai pendaftaran — 200 dan kode terkirim.
	w, _ = doReq("POST", "/auth/otp/send", map[string]any{"email": emailPurpose, "purpose": "login"}, "")
	if w.Code != http.StatusOK {
		t.Fatalf("SendOTP purpose=login: mau 200 (diperlakukan pendaftaran), dapat %d, body=%s", w.Code, w.Body.String())
	}
	if codes[emailPurpose] == "" {
		t.Fatalf("SendOTP purpose=login: kode OTP tidak tertangkap — purpose tidak boleh menghalangi alur pendaftaran")
	}

	// 3) Verifikasi + daftar akun utama.
	w, _ = doReq("POST", "/auth/otp/verify", map[string]string{"email": email, "code": codes[email]}, "")
	if w.Code != http.StatusOK {
		t.Fatalf("VerifyOTP: mau 200, dapat %d, body=%s", w.Code, w.Body.String())
	}
	w, _ = doReq("POST", "/auth/register", map[string]string{
		"name": "Pemilik Unit A", "email": email, "password": "password123", "storeName": "Toko Unit A",
	}, "")
	if w.Code != http.StatusCreated {
		t.Fatalf("Register: mau 201, dapat %d, body=%s", w.Code, w.Body.String())
	}

	// 4) Login hanya email + kata sandi → 200 dengan token.
	w, resp := doReq("POST", "/auth/login", map[string]string{"email": email, "password": "password123"}, "")
	if w.Code != http.StatusOK {
		t.Fatalf("Login polos: mau 200, dapat %d, body=%s", w.Code, w.Body.String())
	}
	if _, ok := resp["access_token"]; !ok {
		t.Fatalf("Login polos: access_token tidak ada di respons: %s", w.Body.String())
	}

	// 5) Login dengan field warisan passcode/otp ikut terkirim → tetap 200:
	// field diabaikan, BUKAN ditantang dan BUKAN memicu verifikasi apa pun.
	w, resp = doReq("POST", "/auth/login", map[string]any{
		"email": email, "password": "password123", "passcode": "99999", "otp": "000000",
	}, "")
	if w.Code != http.StatusOK {
		t.Fatalf("Login + field warisan: mau 200, dapat %d, body=%s", w.Code, w.Body.String())
	}
	if _, ok := resp["access_token"]; !ok {
		t.Fatalf("Login + field warisan: access_token tidak ada di respons: %s", w.Body.String())
	}

	// 6) Kontrol: kata sandi salah tetap ditolak (401) walau field warisan
	// ikut dikirim — field warisan bukan jalan pintas.
	w, _ = doReq("POST", "/auth/login", map[string]any{
		"email": email, "password": "bukan-passwordnya", "passcode": "99999", "otp": "000000",
	}, "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("Login sandi salah + field warisan: mau 401, dapat %d, body=%s", w.Code, w.Body.String())
	}
}
