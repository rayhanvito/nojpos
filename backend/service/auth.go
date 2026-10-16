package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/idtoken"

	"github.com/rayhanvito/nojpos/backend/model"
	"github.com/rayhanvito/nojpos/backend/repo"
)

var (
	ErrInvalidCredentials  = errors.New("email atau kata sandi tidak cocok")
	ErrPasscodeRequired    = errors.New("passcode_required")
	ErrPasscodeWrong       = errors.New("passcode salah")
	ErrEmailTaken          = errors.New("email sudah terdaftar")
	ErrAccountInactive     = errors.New("akun dinonaktifkan")
	ErrTokenInvalid        = errors.New("sesi tidak valid, silakan masuk kembali")
	ErrSwitchSelf          = errors.New("tidak dapat beralih ke akun sendiri")
	ErrInvalidEmail        = errors.New("Email tidak valid.")
	ErrOtpCooldown         = errors.New("Terlalu sering meminta kode. Coba lagi dalam 60 detik.")
	ErrOtpWrong            = errors.New("Kode OTP salah.")
	ErrOtpExpired          = errors.New("Kode OTP sudah kedaluwarsa. Kirim ulang.")
	ErrOtpMaxAttempts      = errors.New("Terlalu banyak percobaan. Kirim ulang kode OTP.")
	ErrEmailNotVerified    = errors.New("Email belum diverifikasi. Silakan verifikasi kode OTP terlebih dahulu.")
	ErrGoogleInvalid       = errors.New("login Google tidak valid")
	ErrGoogleNotConfigured = errors.New("login Google belum dikonfigurasi di server")
	ErrOTPRequired         = errors.New("otp_required")
)

var TestOnOTPSent func(email, code string)

// VerifyGoogleToken validates a GIS ID token. Package-level var (same
// pattern as TestOnOTPSent) so tests can stub it — no network needed.
var VerifyGoogleToken = idtoken.Validate

type AuthService struct {
	users          UserRepository
	cashiers       CashierRepository
	refresh        RefreshRepository
	otps           OtpRepository
	jwtSecret      []byte
	accessTTL      time.Duration
	refreshTTL     time.Duration
	googleClientID string
	activity       *ActivityService
	outlets        OutletRepository
}

func (s *AuthService) SetActivity(a *ActivityService) {
	s.activity = a
}

func (s *AuthService) SetOutlets(outlets OutletRepository) {
	s.outlets = outlets
}

func (s *AuthService) logActivity(ctx context.Context, storeID uint, actorID, actorName, action, detail, refType, refID string) {
	if s.activity == nil {
		return
	}
	if len(detail) > 80 {
		detail = detail[:80]
	}
	s.activity.Log(ctx, storeID, actorID, actorName, action, detail, refType, refID)
}

func NewAuthService(users UserRepository, cashiers CashierRepository, refresh RefreshRepository, otps OtpRepository, jwtSecret string, accessTTL time.Duration, refreshTTL time.Duration, googleClientID ...string) *AuthService {
	svc := &AuthService{
		users:      users,
		cashiers:   cashiers,
		refresh:    refresh,
		otps:       otps,
		jwtSecret:  []byte(jwtSecret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
	if len(googleClientID) > 0 {
		svc.googleClientID = googleClientID[0]
	}
	return svc
}

func (s *AuthService) SendOTP(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	if !isEmail(email) {
		return ErrInvalidEmail
	}

	_, err := s.users.GetByEmail(ctx, email)
	if err == nil {
		return ErrEmailTaken
	} else if !errors.Is(err, repo.ErrNotFound) {
		return err
	}

	existing, err := s.otps.GetOTP(ctx, email)
	if err == nil && existing != nil {
		if time.Since(existing.LastSentAt) < 60*time.Second {
			return ErrOtpCooldown
		}
	}

	code, err := generate6DigitOTP()
	if err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	expiresAt := time.Now().Add(10 * time.Minute)
	if err := s.otps.UpsertOTP(ctx, email, string(hash), expiresAt); err != nil {
		return err
	}

	if TestOnOTPSent != nil {
		TestOnOTPSent(email, code)
	}

	if err := sendOTPEmail(email, code); err != nil {
		return err
	}

	return nil
}

func (s *AuthService) SendLoginOTP(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	if !isEmail(email) {
		return ErrInvalidEmail
	}
	existing, err := s.otps.GetOTP(ctx, email)
	if err == nil && existing != nil {
		if time.Since(existing.LastSentAt) < 60*time.Second {
			return ErrOtpCooldown
		}
	}
	code, err := generate6DigitOTP()
	if err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	expiresAt := time.Now().Add(10 * time.Minute)
	if err := s.otps.UpsertOTP(ctx, email, string(hash), expiresAt); err != nil {
		return err
	}
	if TestOnOTPSent != nil {
		TestOnOTPSent(email, code)
	}
	if err := sendOTPEmail(email, code); err != nil {
		return err
	}
	return nil
}

func (s *AuthService) VerifyOTP(ctx context.Context, email, code string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	code = strings.TrimSpace(code)

	o, err := s.otps.GetOTP(ctx, email)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrOtpWrong
		}
		return err
	}

	if o.Attempts >= 3 {
		return ErrOtpMaxAttempts
	}

	if time.Now().After(o.ExpiresAt) {
		return ErrOtpExpired
	}

	if bcrypt.CompareHashAndPassword([]byte(o.CodeHash), []byte(code)) != nil {
		attempts, _ := s.otps.IncrementAttempts(ctx, email)
		if attempts >= 3 {
			return ErrOtpMaxAttempts
		}
		return ErrOtpWrong
	}

	if err := s.otps.MarkVerified(ctx, email); err != nil {
		return err
	}

	return nil
}

func (s *AuthService) SendPasswordResetOTP(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	if !isEmail(email) {
		return ErrInvalidEmail
	}

	_, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return repo.ErrNotFound
		}
		return err
	}

	existing, err := s.otps.GetOTP(ctx, email)
	if err == nil && existing != nil {
		if time.Since(existing.LastSentAt) < 60*time.Second {
			return ErrOtpCooldown
		}
	}

	code, err := generate6DigitOTP()
	if err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	expiresAt := time.Now().Add(10 * time.Minute)
	if err := s.otps.UpsertOTP(ctx, email, string(hash), expiresAt); err != nil {
		return err
	}

	if TestOnOTPSent != nil {
		TestOnOTPSent(email, code)
	}

	if err := sendPasswordResetOTPEmail(email, code); err != nil {
		return err
	}

	return nil
}

func (s *AuthService) ResetPassword(ctx context.Context, email, code, newPassword string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	code = strings.TrimSpace(code)
	newPassword = strings.TrimSpace(newPassword)

	if len(newPassword) < 8 {
		return fmt.Errorf("kata sandi minimal 8 karakter")
	}

	o, err := s.otps.GetOTP(ctx, email)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrOtpWrong
		}
		return err
	}

	if o.Attempts >= 3 {
		return ErrOtpMaxAttempts
	}

	if time.Now().After(o.ExpiresAt) {
		return ErrOtpExpired
	}

	if bcrypt.CompareHashAndPassword([]byte(o.CodeHash), []byte(code)) != nil {
		attempts, _ := s.otps.IncrementAttempts(ctx, email)
		if attempts >= 3 {
			return ErrOtpMaxAttempts
		}
		return ErrOtpWrong
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err := s.users.UpdatePassword(ctx, email, string(hash)); err != nil {
		return err
	}

	if u, err := s.users.GetByEmail(ctx, email); err == nil {
		_ = s.refresh.RevokeAllForUser(ctx, u.ID)
	}
	_ = s.otps.RevokeOTP(ctx, email)

	return nil
}

func (s *AuthService) ChangePassword(ctx context.Context, claims *Claims, oldPassword, newPassword string) error {
	if claims == nil {
		return ErrTokenInvalid
	}
	if claims.ActingAsCashierID != nil {
		return fmt.Errorf("kasir tidak memiliki kata sandi")
	}
	oldPassword = strings.TrimSpace(oldPassword)
	newPassword = strings.TrimSpace(newPassword)
	if len(newPassword) < 8 {
		return fmt.Errorf("kata sandi minimal 8 karakter")
	}
	if oldPassword == "" {
		return fmt.Errorf("kata sandi lama wajib diisi")
	}
	user, err := s.users.GetByID(ctx, claims.UserID)
	if err != nil {
		return err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)) != nil {
		return ErrInvalidCredentials
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := s.users.UpdatePasswordByID(ctx, user.ID, string(hash)); err != nil {
		return err
	}
	_ = s.refresh.RevokeAllForUser(ctx, user.ID)
	return nil
}

func sendPasswordResetOTPEmail(toEmail, otpCode string) error {
	subject := "Kode Pemulihan Kata Sandi NOJPOS"
	body := fmt.Sprintf("Halo,\n\nKode pemulihan kata sandi NOJPOS Anda adalah: %s\n\nKode ini berlaku selama 10 menit. Jangan berikan kode ini kepada siapa pun.\n\nSalam,\nTim NOJPOS", otpCode)
	return deliverOTPEmail(toEmail, subject, body, "Reset OTP Code", otpCode)
}

func sendOTPEmail(toEmail, otpCode string) error {
	subject := "Kode Verifikasi OTP NOJPOS"
	body := fmt.Sprintf("Halo,\n\nKode verifikasi OTP NOJPOS Anda adalah: %s\n\nKode ini berlaku selama 10 menit. Jangan berikan kode ini kepada siapa pun.\n\nSalam,\nTim NOJPOS", otpCode)
	return deliverOTPEmail(toEmail, subject, body, "OTP Code", otpCode)
}

// deliverOTPEmail mengirim email OTP lewat API Brevo (HTTPS 443) bila
// BREVO_API_KEY terisi; jika tidak, lewat SMTP bila SMTP_EMAIL/SMTP_PASSWORD
// terisi; bila keduanya kosong, kode ditulis ke log (mode dev/mock).
// Kegagalan kirim selalu jatuh ke log agar kode tetap bisa dibaca admin.
func deliverOTPEmail(toEmail, subject, body, logLabel, otpCode string) error {
	if apiKey := getEnv("BREVO_API_KEY", ""); apiKey != "" {
		sender := getEnv("BREVO_SENDER", getEnv("SMTP_FROM", ""))
		if sender == "" {
			log.Printf("[BREVO WARNING] BREVO_API_KEY terisi tetapi BREVO_SENDER/SMTP_FROM kosong; email ke %s tidak dikirim.", toEmail)
			log.Printf("[DEV/MOCK EMAIL FALLBACK] To: %s | %s: %s", toEmail, logLabel, otpCode)
			return nil
		}
		if err := sendViaBrevoAPI(apiKey, sender, toEmail, subject, body); err != nil {
			log.Printf("[BREVO WARNING] Gagal kirim email ke %s: %v. Jatuh ke log.", toEmail, err)
			log.Printf("[DEV/MOCK EMAIL FALLBACK] To: %s | %s: %s", toEmail, logLabel, otpCode)
			return nil
		}
		return nil
	}

	host := getEnv("SMTP_HOST", "smtp.gmail.com")
	port := getEnv("SMTP_PORT", "587")
	user := getEnv("SMTP_EMAIL", getEnv("SMTP_USER", ""))
	pass := getEnv("SMTP_PASSWORD", getEnv("SMTP_PASS", ""))
	from := getEnv("SMTP_FROM", user)

	if user == "" || pass == "" {
		log.Printf("[DEV/MOCK EMAIL] To: %s | %s: %s", toEmail, logLabel, otpCode)
		return nil
	}

	msg := "From: " + from + "\r\n" +
		"To: " + toEmail + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n" +
		body

	auth := smtp.PlainAuth("", user, pass, host)
	addr := host + ":" + port
	if err := smtp.SendMail(addr, auth, from, []string{toEmail}, []byte(msg)); err != nil {
		log.Printf("[SMTP WARNING] Failed to send email to %s: %v. Falling back to console log. %s: %s", toEmail, err, logLabel, otpCode)
		log.Printf("[DEV/MOCK EMAIL FALLBACK] To: %s | %s: %s", toEmail, logLabel, otpCode)
		return nil
	}
	return nil
}

// sendViaBrevoAPI mengirim satu email teks lewat Brevo Transactional API.
func sendViaBrevoAPI(apiKey, sender, toEmail, subject, body string) error {
	payload := map[string]any{
		"sender":      map[string]string{"email": sender, "name": "NOJPOS"},
		"to":          []map[string]string{{"email": toEmail}},
		"subject":     subject,
		"textContent": body,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, "https://api.brevo.com/v3/smtp/email", bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("api-key", apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(resp.Body)
		return fmt.Errorf("Brevo API status %d: %s", resp.StatusCode, strings.TrimSpace(buf.String()))
	}
	return nil
}

func generate6DigitOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return "", err
	}
	val := n.Int64() + 100000
	return fmt.Sprintf("%d", val), nil
}

func (s *AuthService) Register(ctx context.Context, name, email, password, storeName string) (*model.User, *model.TokenPair, error) {
	name = strings.TrimSpace(name)
	email = strings.ToLower(strings.TrimSpace(email))
	storeName = strings.TrimSpace(storeName)

	if name == "" {
		return nil, nil, fmt.Errorf("nama wajib diisi")
	}
	if !isEmail(email) {
		return nil, nil, ErrInvalidEmail
	}
	if len(password) < 8 {
		return nil, nil, fmt.Errorf("kata sandi minimal 8 karakter")
	}
	if storeName == "" {
		return nil, nil, fmt.Errorf("nama toko wajib diisi")
	}

	verified, err := s.otps.IsEmailVerified(ctx, email)
	if err != nil || !verified {
		return nil, nil, ErrEmailNotVerified
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, err
	}

	user, err := s.users.RegisterTx(ctx, storeName, email, name, string(hash))
	if err != nil {
		if errors.Is(err, repo.ErrDuplicate) {
			return nil, nil, ErrEmailTaken
		}
		return nil, nil, err
	}

	pair, err := s.issueTokens(ctx, user.ID, nil)
	if err != nil {
		return nil, nil, err
	}
	s.setSessionOutlet(user, pair)
	now := time.Now().UTC()
	_ = s.users.UpdateLastSeenAt(ctx, user.ID, &now)
	user.LastSeenAt = &now
	s.logActivity(ctx, user.StoreID, fmt.Sprintf("%d", user.ID), user.Name, ActivityLogin, fmt.Sprintf("%s login ke sistem", user.Name), "", "")
	return user, pair, nil
}

func (s *AuthService) Login(ctx context.Context, email, password, passcode, otp string) (*model.User, *model.TokenPair, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, nil, ErrInvalidCredentials
		}
		return nil, nil, err
	}
	if !user.Active {
		return nil, nil, ErrAccountInactive
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return nil, nil, ErrInvalidCredentials
	}
	// Passcode bukan syarat login (keputusan produk 2026-10-15): passcode hanya
	// diminta saat ganti akun lewat Switch, bukan saat login email/Google.

	pair, err := s.issueTokens(ctx, user.ID, nil)
	if err != nil {
		return nil, nil, err
	}
	s.setSessionOutlet(user, pair)
	now2 := time.Now().UTC()
	_ = s.users.UpdateLastSeenAt(ctx, user.ID, &now2)
	user.LastSeenAt = &now2
	s.logActivity(ctx, user.StoreID, fmt.Sprintf("%d", user.ID), user.Name, ActivityLogin, fmt.Sprintf("%s login ke sistem", user.Name), "", "")
	return user, pair, nil
}

// GoogleLogin verifies a GIS ID token and returns our own token pair.
// Existing emails auto-link (password keeps working); new emails get a
// fresh store + admin, mirroring Register without the OTP gate or password.
// Login Google mengikuti aturan yang sama: tanpa passcode.
func (s *AuthService) GoogleLogin(ctx context.Context, idToken, storeName, passcode string) (*model.User, *model.TokenPair, error) {
	if s.googleClientID == "" {
		return nil, nil, ErrGoogleNotConfigured
	}
	if strings.TrimSpace(idToken) == "" {
		return nil, nil, ErrGoogleInvalid
	}

	payload, err := VerifyGoogleToken(ctx, idToken, s.googleClientID)
	if err != nil {
		return nil, nil, ErrGoogleInvalid
	}

	email, _ := payload.Claims["email"].(string)
	email = strings.ToLower(strings.TrimSpace(email))
	if !isEmail(email) {
		return nil, nil, ErrGoogleInvalid
	}
	if !googleEmailVerified(payload.Claims["email_verified"]) {
		return nil, nil, ErrEmailNotVerified
	}
	name := strings.TrimSpace(stringClaim(payload.Claims, "name"))
	if name == "" {
		name = strings.Split(email, "@")[0]
	}

	user, err := s.users.GetByEmail(ctx, email)
	if err == nil {
		if !user.Active {
			return nil, nil, ErrAccountInactive
		}
		pair, err := s.issueTokens(ctx, user.ID, nil)
		if err != nil {
			return nil, nil, err
		}
		s.setSessionOutlet(user, pair)
		now := time.Now().UTC()
		_ = s.users.UpdateLastSeenAt(ctx, user.ID, &now)
		user.LastSeenAt = &now
		s.logActivity(ctx, user.StoreID, fmt.Sprintf("%d", user.ID), user.Name, ActivityLogin, fmt.Sprintf("%s login ke sistem", user.Name), "", "")
		return user, pair, nil
	} else if !errors.Is(err, repo.ErrNotFound) {
		return nil, nil, err
	}

	storeName = strings.TrimSpace(storeName)
	if storeName == "" {
		storeName = name + "'s Store"
	}

	// Empty password hash: column stays NOT NULL, and bcrypt can never
	// match it, so password-login for Google users safely 401s.
	user, err = s.users.RegisterTx(ctx, storeName, email, name, "")
	if err != nil {
		if errors.Is(err, repo.ErrDuplicate) {
			return nil, nil, ErrEmailTaken
		}
		return nil, nil, err
	}
	pair, err := s.issueTokens(ctx, user.ID, nil)
	if err != nil {
		return nil, nil, err
	}
	s.setSessionOutlet(user, pair)
	now2 := time.Now().UTC()
	_ = s.users.UpdateLastSeenAt(ctx, user.ID, &now2)
	user.LastSeenAt = &now2
	s.logActivity(ctx, user.StoreID, fmt.Sprintf("%d", user.ID), user.Name, ActivityLogin, fmt.Sprintf("%s login ke sistem", user.Name), "", "")
	return user, pair, nil
}

func stringClaim(claims map[string]interface{}, key string) string {
	v, _ := claims[key].(string)
	return v
}

// Google sends email_verified as bool; accept "true"/1 defensively.
func googleEmailVerified(v interface{}) bool {
	switch t := v.(type) {
	case bool:
		return t
	case string:
		return strings.EqualFold(strings.TrimSpace(t), "true")
	case float64:
		return t == 1
	default:
		return false
	}
}

func (s *AuthService) Me(ctx context.Context, claims *Claims) (*model.PublicUser, error) {
	if claims == nil {
		return nil, ErrTokenInvalid
	}
	if err := s.populateActiveOutlet(ctx, claims); err != nil {
		return nil, ErrTokenInvalid
	}
	if claims.ActingAsCashierID != nil {
		c, err := s.cashiers.GetByID(ctx, *claims.ActingAsCashierID)
		if err != nil {
			return nil, ErrTokenInvalid
		}
		if !c.Active {
			return nil, ErrAccountInactive
		}
		owner, err := s.users.GetByID(ctx, claims.UserID)
		storeName := ""
		if err == nil {
			storeName = owner.StoreName
		}
		pub := c.Public(storeName)
		setPublicOutlet(&pub, claims.OutletID)
		return &pub, nil
	}

	user, err := s.users.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	if !user.Active {
		return nil, ErrAccountInactive
	}
	pub := user.Public()
	setPublicOutlet(&pub, claims.OutletID)
	return &pub, nil
}

func (s *AuthService) populateActiveOutlet(ctx context.Context, claims *Claims) error {
	var owner *model.User
	var err error
	if claims.ActingAsCashierID != nil {
		owner, err = s.users.GetByID(ctx, claims.UserID)
		if err != nil {
			return err
		}
		cashier, err := s.cashiers.GetByID(ctx, *claims.ActingAsCashierID)
		if err != nil || cashier.StoreID != owner.StoreID || !cashier.Active {
			return ErrTokenInvalid
		}
		if cashier.OutletID == nil {
			return ErrTokenInvalid
		}
		if claims.OutletID != 0 && claims.OutletID != *cashier.OutletID {
			return ErrTokenInvalid
		}
		claims.OutletID = *cashier.OutletID
	} else if claims.OutletID == 0 {
		owner, err = s.users.GetByID(ctx, claims.UserID)
		if err != nil {
			return err
		}
		if owner.OutletID != nil {
			claims.OutletID = *owner.OutletID
		}
	}
	if claims.OutletID == 0 && s.outlets != nil {
		outlet, err := s.outlets.GetDefault(ctx, claims.StoreID)
		if err != nil {
			return err
		}
		claims.OutletID = outlet.ID
	}
	if claims.OutletID != 0 && s.outlets != nil {
		if _, err := s.outlets.GetByID(ctx, claims.StoreID, claims.OutletID); err != nil {
			return err
		}
	}
	return nil
}

func setPublicOutlet(user *model.PublicUser, outletID uint) {
	if outletID == 0 {
		return
	}
	user.OutletID = &outletID
}

func (s *AuthService) SwitchOutlet(ctx context.Context, claims *Claims, outletID uint) (*model.PublicUser, *model.TokenPair, error) {
	if claims == nil || s.outlets == nil {
		return nil, nil, ErrTokenInvalid
	}
	if _, err := s.outlets.GetByID(ctx, claims.StoreID, outletID); err != nil {
		return nil, nil, ErrOutletNotFound
	}
	owner, err := s.users.GetByID(ctx, claims.UserID)
	if err != nil || !owner.Active {
		return nil, nil, ErrTokenInvalid
	}
	if claims.ActingAsCashierID != nil {
		cashier, err := s.cashiers.GetByID(ctx, *claims.ActingAsCashierID)
		if err != nil || !cashier.Active || cashier.StoreID != owner.StoreID {
			return nil, nil, ErrTokenInvalid
		}
		if cashier.OutletID == nil || *cashier.OutletID != outletID {
			return nil, nil, ErrOutletForbidden
		}
	}
	pair, err := s.issueTokens(ctx, owner.ID, claims.ActingAsCashierID, outletID)
	if err != nil {
		return nil, nil, err
	}
	nextClaims := *claims
	nextClaims.OutletID = outletID
	user, err := s.Me(ctx, &nextClaims)
	if err != nil {
		return nil, nil, err
	}
	return user, pair, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*model.PublicUser, *model.TokenPair, error) {
	hash := hashToken(refreshToken)
	rt, err := s.refresh.GetActiveByHash(ctx, hash)
	if err != nil || rt.Revoked || time.Now().After(rt.ExpiresAt) {
		return nil, nil, ErrTokenInvalid
	}

	owner, err := s.users.GetByID(ctx, rt.UserID)
	if err != nil || !owner.Active {
		return nil, nil, ErrTokenInvalid
	}

	// Konteks sesi wajib bertahan lintas refresh: kasir tetap kasir,
	// admin tetap admin. Tanpa ini sesi kasir berubah jadi admin.
	var pub model.PublicUser
	if rt.ActingAsCashierID != nil {
		c, err := s.cashiers.GetByID(ctx, *rt.ActingAsCashierID)
		if err != nil || !c.Active || c.StoreID != owner.StoreID {
			return nil, nil, ErrTokenInvalid
		}
		pub = c.Public(owner.StoreName)
	} else {
		pub = owner.Public()
	}

	if err := s.refresh.Revoke(ctx, hash); err != nil {
		return nil, nil, err
	}
	var pair *model.TokenPair
	if rt.OutletID != nil {
		pair, err = s.issueTokens(ctx, owner.ID, rt.ActingAsCashierID, *rt.OutletID)
	} else {
		pair, err = s.issueTokens(ctx, owner.ID, rt.ActingAsCashierID)
	}
	if err != nil {
		return nil, nil, err
	}
	if active, err := s.ParseAccess(pair.AccessToken); err == nil {
		setPublicOutlet(&pub, active.OutletID)
	}
	return &pub, pair, nil
}

func (s *AuthService) RecordHeartbeat(ctx context.Context, claims *Claims, now time.Time) error {
	if claims.ActingAsCashierID != nil {
		return s.cashiers.UpdateLastSeenAt(ctx, *claims.ActingAsCashierID, &now)
	}
	return s.users.UpdateLastSeenAt(ctx, claims.UserID, &now)
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string, claims ...*Claims) {
	logged := false
	if refreshToken != "" {
		hash := hashToken(refreshToken)
		if rt, err := s.refresh.GetActiveByHash(ctx, hash); err == nil && !rt.Revoked && !time.Now().After(rt.ExpiresAt) {
			// Tandai entity sesi yang benar offline (kasir vs owner).
			if rt.ActingAsCashierID != nil {
				_ = s.cashiers.UpdateLastSeenAt(ctx, *rt.ActingAsCashierID, nil)
				if owner, err := s.users.GetByID(ctx, rt.UserID); err == nil {
					if c, err := s.cashiers.GetByID(ctx, *rt.ActingAsCashierID); err == nil {
						s.logActivity(ctx, owner.StoreID, fmt.Sprintf("%d", c.ID), c.Name, ActivityLogout, fmt.Sprintf("%s keluar dari sistem (kasir)", c.Name), "", "")
						logged = true
					}
				}
			} else {
				_ = s.users.UpdateLastSeenAt(ctx, rt.UserID, nil)
				if owner, err := s.users.GetByID(ctx, rt.UserID); err == nil {
					s.logActivity(ctx, owner.StoreID, fmt.Sprintf("%d", owner.ID), owner.Name, ActivityLogout, fmt.Sprintf("%s keluar dari sistem", owner.Name), "", "")
					logged = true
				}
			}
		}
		_ = s.refresh.Revoke(ctx, hash)
	}
	// Fallback via Authorization header (frontend baru kirim header tanpa body) — gabungan dengan rt di atas.
	if len(claims) > 0 && claims[0] != nil && claims[0].ActingAsCashierID != nil {
		_ = s.cashiers.UpdateLastSeenAt(ctx, *claims[0].ActingAsCashierID, nil)
		if !logged {
			if owner, err := s.users.GetByID(ctx, claims[0].UserID); err == nil {
				if c, err := s.cashiers.GetByID(ctx, *claims[0].ActingAsCashierID); err == nil {
					s.logActivity(ctx, owner.StoreID, fmt.Sprintf("%d", c.ID), c.Name, ActivityLogout, fmt.Sprintf("%s keluar dari sistem (kasir)", c.Name), "", "")
				}
			}
		}
	} else if len(claims) > 0 && claims[0] != nil {
		_ = s.users.UpdateLastSeenAt(ctx, claims[0].UserID, nil)
		if !logged {
			if owner, err := s.users.GetByID(ctx, claims[0].UserID); err == nil {
				s.logActivity(ctx, owner.StoreID, fmt.Sprintf("%d", owner.ID), owner.Name, ActivityLogout, fmt.Sprintf("%s keluar dari sistem", owner.Name), "", "")
			}
		}
	}
}

func (s *AuthService) switchSource(ctx context.Context, claims *Claims, owner *model.User) (actorID, actorName string) {
	if claims.ActingAsCashierID != nil {
		if c, err := s.cashiers.GetByID(ctx, *claims.ActingAsCashierID); err == nil {
			return fmt.Sprintf("%d", c.ID), c.Name
		}
		return fmt.Sprintf("%d", *claims.ActingAsCashierID), claims.Name
	}
	return fmt.Sprintf("%d", owner.ID), owner.Name
}

func (s *AuthService) Switch(ctx context.Context, claims *Claims, targetID uint, passcode string, roleHint ...string) (*model.PublicUser, *model.TokenPair, error) {
	owner, err := s.users.GetByID(ctx, claims.UserID)
	if err != nil || !owner.Active {
		return nil, nil, ErrTokenInvalid
	}

	hint := ""
	if len(roleHint) > 0 {
		hint = roleHint[0]
	}

	now := time.Now().UTC()
	offlineCurrent := func() {
		if claims.ActingAsCashierID != nil {
			_ = s.cashiers.UpdateLastSeenAt(ctx, *claims.ActingAsCashierID, nil)
		} else {
			_ = s.users.UpdateLastSeenAt(ctx, claims.UserID, nil)
		}
	}

	srcID, srcName := s.switchSource(ctx, claims, owner)

	if hint == "cashier" {
		if claims.ActingAsCashierID != nil && targetID == *claims.ActingAsCashierID {
			return nil, nil, ErrSwitchSelf
		}
		if c, err := s.cashiers.GetByID(ctx, targetID); err == nil && c.StoreID == owner.StoreID {
			if !c.Active {
				return nil, nil, ErrAccountInactive
			}
			if err := checkPasscode(c.PasscodeHash, passcode); err != nil {
				return nil, nil, err
			}
			pair, err := s.issueTokens(ctx, owner.ID, &c.ID)
			if err != nil {
				return nil, nil, err
			}
			offlineCurrent()
			_ = s.cashiers.UpdateLastSeenAt(ctx, c.ID, &now)
			c.LastSeenAt = &now
			pub := c.Public(owner.StoreName)
			s.logActivity(ctx, owner.StoreID, srcID, srcName, ActivitySwitch, fmt.Sprintf("%s beralih ke %s (kasir)", srcName, c.Name), "user", fmt.Sprintf("%d", c.ID))
			return &pub, pair, nil
		}
		return nil, nil, repo.ErrNotFound
	}

	if hint == "admin" {
		if owner.PasscodeHash == nil || *owner.PasscodeHash == "" {
			return nil, nil, ErrPasscodeRequired
		}
		if err := checkPasscode(owner.PasscodeHash, passcode); err != nil {
			return nil, nil, err
		}
		pair, err := s.issueTokens(ctx, owner.ID, nil)
		if err != nil {
			return nil, nil, err
		}
		offlineCurrent()
		_ = s.users.UpdateLastSeenAt(ctx, owner.ID, &now)
		owner.LastSeenAt = &now
		pub := owner.Public()
		setPublicOutlet(&pub, s.sessionOutletID(pair))
		s.logActivity(ctx, owner.StoreID, srcID, srcName, ActivitySwitch, fmt.Sprintf("%s beralih ke %s (admin)", srcName, owner.Name), "user", fmt.Sprintf("%d", owner.ID))
		return &pub, pair, nil
	}

	// Default: check cashier first, then owner
	if claims.ActingAsCashierID != nil && targetID == *claims.ActingAsCashierID {
		return nil, nil, ErrSwitchSelf
	}
	if c, err := s.cashiers.GetByID(ctx, targetID); err == nil && c.StoreID == owner.StoreID {
		if !c.Active {
			return nil, nil, ErrAccountInactive
		}
		if err := checkPasscode(c.PasscodeHash, passcode); err != nil {
			return nil, nil, err
		}
		pair, err := s.issueTokens(ctx, owner.ID, &c.ID)
		if err != nil {
			return nil, nil, err
		}
		offlineCurrent()
		_ = s.cashiers.UpdateLastSeenAt(ctx, c.ID, &now)
		c.LastSeenAt = &now
		pub := c.Public(owner.StoreName)
		s.logActivity(ctx, owner.StoreID, srcID, srcName, ActivitySwitch, fmt.Sprintf("%s beralih ke %s (kasir)", srcName, c.Name), "user", fmt.Sprintf("%d", c.ID))
		return &pub, pair, nil
	}

	if targetID == owner.ID {
		if owner.PasscodeHash == nil || *owner.PasscodeHash == "" {
			return nil, nil, ErrPasscodeRequired
		}
		if err := checkPasscode(owner.PasscodeHash, passcode); err != nil {
			return nil, nil, err
		}
		pair, err := s.issueTokens(ctx, owner.ID, nil)
		if err != nil {
			return nil, nil, err
		}
		offlineCurrent()
		_ = s.users.UpdateLastSeenAt(ctx, owner.ID, &now)
		owner.LastSeenAt = &now
		pub := owner.Public()
		setPublicOutlet(&pub, s.sessionOutletID(pair))
		s.logActivity(ctx, owner.StoreID, srcID, srcName, ActivitySwitch, fmt.Sprintf("%s beralih ke %s (admin)", srcName, owner.Name), "user", fmt.Sprintf("%d", owner.ID))
		return &pub, pair, nil
	}

	return nil, nil, repo.ErrNotFound
}

func (s *AuthService) switchToOwner(ctx context.Context, owner *model.User, passcode string) (*model.PublicUser, *model.TokenPair, error) {
	if owner.PasscodeHash == nil || *owner.PasscodeHash == "" {
		return nil, nil, ErrPasscodeRequired
	}
	if err := checkPasscode(owner.PasscodeHash, passcode); err != nil {
		return nil, nil, err
	}
	pair, err := s.issueTokens(ctx, owner.ID, nil)
	if err != nil {
		return nil, nil, err
	}
	now := time.Now().UTC()
	_ = s.users.UpdateLastSeenAt(ctx, owner.ID, &now)
	owner.LastSeenAt = &now
	pub := owner.Public()
	setPublicOutlet(&pub, s.sessionOutletID(pair))
	return &pub, pair, nil
}

func (s *AuthService) setSessionOutlet(user *model.User, pair *model.TokenPair) {
	if user == nil || pair == nil {
		return
	}
	if outletID := s.sessionOutletID(pair); outletID != 0 {
		user.OutletID = &outletID
	}
}

func (s *AuthService) sessionOutletID(pair *model.TokenPair) uint {
	if pair == nil {
		return 0
	}
	claims, err := s.ParseAccess(pair.AccessToken)
	if err != nil {
		return 0
	}
	return claims.OutletID
}

func (s *AuthService) switchToCashier(ctx context.Context, owner *model.User, targetID uint, passcode string) (*model.PublicUser, *model.TokenPair, error) {
	cashier, err := s.cashiers.GetByID(ctx, targetID)
	if err != nil {
		return nil, nil, repo.ErrNotFound
	}
	if cashier.StoreID != owner.StoreID {
		return nil, nil, repo.ErrNotFound
	}
	if !cashier.Active {
		return nil, nil, ErrAccountInactive
	}
	if err := checkPasscode(cashier.PasscodeHash, passcode); err != nil {
		return nil, nil, err
	}
	pair, err := s.issueTokens(ctx, owner.ID, &cashier.ID)
	if err != nil {
		return nil, nil, err
	}
	pub := cashier.Public(owner.StoreName)
	return &pub, pair, nil
}

func checkPasscode(hash *string, passcode string) error {
	if hash == nil || *hash == "" {
		return nil
	}
	if passcode == "" {
		return ErrPasscodeRequired
	}
	if bcrypt.CompareHashAndPassword([]byte(*hash), []byte(passcode)) != nil {
		return ErrPasscodeWrong
	}
	return nil
}

type Claims struct {
	UserID            uint
	StoreID           uint
	OutletID          uint
	Name              string
	Email             string
	ActingAsCashierID *uint
	Role              model.Role
}

func (c *Claims) ActiveRole() model.Role {
	if c.ActingAsCashierID != nil {
		return model.RoleCashier
	}
	return model.RoleAdmin
}

func (s *AuthService) issueTokens(ctx context.Context, ownerID uint, actingAsCashierID *uint, selectedOutletID ...uint) (*model.TokenPair, error) {
	owner, err := s.users.GetByID(ctx, ownerID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	var selected *uint
	if len(selectedOutletID) > 0 && selectedOutletID[0] > 0 {
		selected = &selectedOutletID[0]
	}
	activeOutletID := s.resolveOutletID(ctx, owner, actingAsCashierID, selected)
	if s.outlets != nil {
		if activeOutletID == 0 {
			return nil, ErrOutletNotFound
		}
		if _, err := s.outlets.GetByID(ctx, owner.StoreID, activeOutletID); err != nil {
			return nil, ErrOutletNotFound
		}
		if actingAsCashierID != nil {
			cashier, err := s.cashiers.GetByID(ctx, *actingAsCashierID)
			if err != nil || cashier.StoreID != owner.StoreID || !cashier.Active || cashier.OutletID == nil || *cashier.OutletID != activeOutletID {
				return nil, ErrOutletForbidden
			}
		}
	}
	mc := jwt.MapClaims{
		"sub":   strconv.FormatUint(uint64(owner.ID), 10),
		"sid":   strconv.FormatUint(uint64(owner.StoreID), 10),
		"name":  owner.Name,
		"email": owner.Email,
		"iat":   now.Unix(),
		"exp":   now.Add(s.accessTTL).Unix(),
	}
	if activeOutletID != 0 {
		mc["outletId"] = strconv.FormatUint(uint64(activeOutletID), 10)
	}
	if actingAsCashierID != nil {
		mc["acting_as"] = strconv.FormatUint(uint64(*actingAsCashierID), 10)
	}

	access, err := jwt.NewWithClaims(jwt.SigningMethodHS256, mc).SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}

	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return nil, err
	}
	refresh := hex.EncodeToString(raw)

	var outletID *uint
	if activeOutletID != 0 {
		outletID = &activeOutletID
	}
	if err := s.refresh.Create(ctx, owner.ID, actingAsCashierID, outletID, hashToken(refresh), now.Add(s.refreshTTL)); err != nil {
		return nil, err
	}

	return &model.TokenPair{AccessToken: access, RefreshToken: refresh}, nil
}

func (s *AuthService) resolveOutletID(ctx context.Context, owner *model.User, actingAsCashierID *uint, selected *uint) uint {
	if selected != nil {
		return *selected
	}
	if actingAsCashierID != nil {
		if cashier, err := s.cashiers.GetByID(ctx, *actingAsCashierID); err == nil && cashier.OutletID != nil {
			return *cashier.OutletID
		}
	}
	if s.outlets != nil {
		if outlet, err := s.outlets.GetDefault(ctx, owner.StoreID); err == nil {
			return outlet.ID
		}
	}
	if owner.OutletID != nil {
		return *owner.OutletID
	}
	return 0
}

func (s *AuthService) ParseAccess(tokenStr string) (*Claims, error) {
	tok, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("metode tanda tangan tak terduga: %v", t.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil || !tok.Valid {
		return nil, ErrTokenInvalid
	}
	mc, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrTokenInvalid
	}
	sub, _ := mc["sub"].(string)
	sid, _ := mc["sid"].(string)
	name, _ := mc["name"].(string)
	email, _ := mc["email"].(string)
	actingAs, _ := mc["acting_as"].(string)
	outletIDText, _ := mc["outletId"].(string)
	if sub == "" {
		return nil, ErrTokenInvalid
	}
	uid, err := strconv.ParseUint(sub, 10, 64)
	if err != nil {
		return nil, ErrTokenInvalid
	}
	sidNum, err := strconv.ParseUint(sid, 10, 64)
	if err != nil {
		return nil, ErrTokenInvalid
	}
	var outletID uint
	if outletIDText != "" {
		parsedOutletID, err := strconv.ParseUint(outletIDText, 10, 64)
		if err != nil || parsedOutletID == 0 {
			return nil, ErrTokenInvalid
		}
		outletID = uint(parsedOutletID)
	}

	c := &Claims{
		UserID:   uint(uid),
		StoreID:  uint(sidNum),
		OutletID: outletID,
		Name:     name,
		Email:    email,
		Role:     model.RoleAdmin,
	}
	if actingAs != "" {
		actingNum, err := strconv.ParseUint(actingAs, 10, 64)
		if err != nil {
			return nil, ErrTokenInvalid
		}
		actingID := uint(actingNum)
		c.ActingAsCashierID = &actingID
		c.Role = model.RoleCashier
	}
	return c, nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func isEmail(s string) bool {
	at := strings.IndexByte(s, '@')
	return at > 0 && at < len(s)-1 && strings.Contains(s[at:], ".") && !strings.ContainsAny(s, " \t")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
