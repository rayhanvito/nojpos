package auth

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/rayhanvito/nojpos/backend/internal/outlet"
	"github.com/rayhanvito/nojpos/backend/internal/platform/middleware"
)

// Handler adalah Handler lama di handler/auth.go (BE-F1c), plus
// Heartbeat dari handler/presence.go dan SwitchOutlet dari rump
// handler/outlet.go — badan metode verbatim.
type Handler struct {
	auth *Service
}

func NewHandler(auth *Service) *Handler { return &Handler{auth: auth} }

type registerReq struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	StoreName string `json:"storeName"`
}

type otpSendReq struct {
	Email   string `json:"email"`
	Purpose string `json:"purpose,omitempty"`
}

type otpVerifyReq struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type forgotPasswordSendReq struct {
	Email string `json:"email"`
}

type forgotPasswordResetReq struct {
	Email       string `json:"email"`
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Passcode string `json:"passcode,omitempty"`
	Otp      string `json:"otp,omitempty"`
}

type refreshReq struct {
	RefreshToken string `json:"refresh_token"`
}

type logoutReq struct {
	RefreshToken string `json:"refresh_token"`
}

type switchReq struct {
	TargetUserID uint   `json:"target_user_id"`
	Passcode     string `json:"passcode,omitempty"`
	Role         string `json:"role,omitempty"`
}

type googleReq struct {
	IDToken   string `json:"id_token"`
	StoreName string `json:"storeName"`
	Passcode  string `json:"passcode,omitempty"`
}

type changePasswordReq struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type authResponse struct {
	User PublicUser `json:"user"`
	TokenPair
}

func (h *Handler) SendOTP(c *gin.Context) {
	var req otpSendReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body JSON tidak valid"})
		return
	}
	var err error
	if req.Purpose == "login" {
		err = h.auth.SendLoginOTP(c.Request.Context(), req.Email)
	} else {
		err = h.auth.SendOTP(c.Request.Context(), req.Email)
	}
	if err != nil {
		if errors.Is(err, ErrEmailTaken) {
			c.JSON(http.StatusOK, gin.H{"message": "Jika email belum terdaftar, kode OTP telah dikirim."})
			return
		}
		respondOTPErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Kode OTP terkirim ke email Anda."})
}

func (h *Handler) VerifyOTP(c *gin.Context) {
	var req otpVerifyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body JSON tidak valid"})
		return
	}
	err := h.auth.VerifyOTP(c.Request.Context(), req.Email, req.Code)
	if err != nil {
		respondOTPErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"verified": true, "message": "Email berhasil diverifikasi."})
}

func (h *Handler) SendPasswordResetOTP(c *gin.Context) {
	var req forgotPasswordSendReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body JSON tidak valid"})
		return
	}
	err := h.auth.SendPasswordResetOTP(c.Request.Context(), req.Email)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusOK, gin.H{"message": "Jika email terdaftar, kode OTP telah dikirim."})
			return
		}
		respondOTPErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Jika email terdaftar, kode OTP telah dikirim."})
}

func (h *Handler) ResetPassword(c *gin.Context) {
	var req forgotPasswordResetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body JSON tidak valid"})
		return
	}
	err := h.auth.ResetPassword(c.Request.Context(), req.Email, req.Code, req.NewPassword)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Akun tidak ditemukan."})
		case errors.Is(err, ErrOtpWrong):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Kode OTP salah."})
		case errors.Is(err, ErrOtpExpired):
			c.JSON(http.StatusGone, gin.H{"error": "Kode OTP sudah kedaluwarsa. Kirim ulang."})
		case errors.Is(err, ErrOtpMaxAttempts):
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Terlalu banyak percobaan. Kirim ulang kode OTP."})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Kata sandi berhasil diubah. Silakan masuk dengan kata sandi baru."})
}

func (h *Handler) Register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body JSON tidak valid"})
		return
	}
	user, pair, err := h.auth.Register(c.Request.Context(), req.Name, req.Email, req.Password, req.StoreName)
	if err != nil {
		respondAuthErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, authResponse{User: user.Public(), TokenPair: *pair})
}

func (h *Handler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body JSON tidak valid"})
		return
	}
	user, pair, err := h.auth.Login(c.Request.Context(), req.Email, req.Password, req.Passcode, req.Otp)
	if err != nil {
		respondAuthErr(c, err)
		return
	}
	c.JSON(http.StatusOK, authResponse{User: user.Public(), TokenPair: *pair})
}

func (h *Handler) Google(c *gin.Context) {
	var req googleReq
	if err := c.ShouldBindJSON(&req); err != nil || req.IDToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id_token wajib diisi"})
		return
	}
	user, pair, err := h.auth.GoogleLogin(c.Request.Context(), req.IDToken, req.StoreName, req.Passcode)
	if err != nil {
		respondGoogleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, authResponse{User: user.Public(), TokenPair: *pair})
}

func (h *Handler) Refresh(c *gin.Context) {
	var req refreshReq
	if err := c.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refresh_token wajib diisi"})
		return
	}
	pubUser, pair, err := h.auth.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, authResponse{User: *pubUser, TokenPair: *pair})
}

func (h *Handler) Logout(c *gin.Context) {
	var req logoutReq
	_ = c.ShouldBindJSON(&req)
	claims := middleware.ClaimsFrom(c)
	if claims == nil {
		if authHeader := c.GetHeader("Authorization"); strings.HasPrefix(authHeader, "Bearer ") {
			if parsed, err := h.auth.ParseAccess(strings.TrimPrefix(authHeader, "Bearer ")); err == nil {
				claims = parsed
			}
		}
	}
	h.auth.Logout(c.Request.Context(), req.RefreshToken, claims)
	c.JSON(http.StatusOK, gin.H{"message": "keluar berhasil"})
}

func (h *Handler) Me(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	pubUser, err := h.auth.Me(c.Request.Context(), claims)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "sesi tidak valid, silakan masuk kembali"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": pubUser})
}

func (h *Handler) ChangePassword(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	var req changePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body JSON tidak valid"})
		return
	}
	if err := h.auth.ChangePassword(c.Request.Context(), claims, req.OldPassword, req.NewPassword); err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Kata sandi lama salah."})
			return
		}
		if err.Error() == "kasir tidak memiliki kata sandi" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Kasir tidak memiliki kata sandi. Gunakan passcode."})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Kata sandi berhasil diperbarui. Silakan masuk kembali."})
}

func (h *Handler) Switch(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	var req switchReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body JSON tidak valid"})
		return
	}
	if req.TargetUserID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target_user_id wajib diisi"})
		return
	}
	pubUser, pair, err := h.auth.Switch(c.Request.Context(), claims, req.TargetUserID, req.Passcode, req.Role)
	if err != nil {
		respondSwitchErr(c, err)
		return
	}
	c.JSON(http.StatusOK, authResponse{User: *pubUser, TokenPair: *pair})
}

func respondSwitchErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrSwitchSelf):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tidak dapat beralih ke akun sendiri."})
	case errors.Is(err, ErrPasscodeRequired):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "passcode_required", "code": "PASSCODE_REQUIRED"})
	case errors.Is(err, ErrPasscodeWrong):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Passcode salah. Coba lagi.", "code": "PASSCODE_WRONG"})
	case errors.Is(err, ErrAccountInactive):
		c.JSON(http.StatusForbidden, gin.H{"error": "Akun dinonaktifkan."})
	case errors.Is(err, ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Akun tidak ditemukan."})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

func respondOTPErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalidEmail):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email tidak valid."})
	case errors.Is(err, ErrOtpCooldown):
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Terlalu sering meminta kode. Coba lagi dalam 60 detik."})
	case errors.Is(err, ErrOtpWrong):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Kode OTP salah."})
	case errors.Is(err, ErrOtpExpired):
		c.JSON(http.StatusGone, gin.H{"error": "Kode OTP sudah kedaluwarsa. Kirim ulang."})
	case errors.Is(err, ErrOtpMaxAttempts):
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Terlalu banyak percobaan. Kirim ulang kode OTP."})
	case errors.Is(err, ErrEmailTaken):
		c.JSON(http.StatusConflict, gin.H{"error": "Email sudah terdaftar. Silakan masuk."})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

func respondGoogleErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrPasscodeRequired):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "passcode_required", "code": "PASSCODE_REQUIRED"})
	case errors.Is(err, ErrPasscodeWrong):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Passcode salah. Coba lagi.", "code": "PASSCODE_WRONG"})
	case errors.Is(err, ErrGoogleInvalid):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "login Google tidak valid"})
	case errors.Is(err, ErrGoogleNotConfigured):
		c.JSON(http.StatusInternalServerError, gin.H{"error": "login Google belum dikonfigurasi di server"})
	case errors.Is(err, ErrEmailNotVerified):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email Google belum diverifikasi."})
	case errors.Is(err, ErrEmailTaken):
		c.JSON(http.StatusConflict, gin.H{"error": "Email sudah terdaftar. Silakan masuk."})
	case errors.Is(err, ErrAccountInactive):
		c.JSON(http.StatusForbidden, gin.H{"error": "Akun dinonaktifkan. Hubungi admin toko."})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

func respondAuthErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrOTPRequired):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "otp_required", "code": "OTP_REQUIRED"})
	case errors.Is(err, ErrPasscodeRequired):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "passcode_required", "code": "PASSCODE_REQUIRED"})
	case errors.Is(err, ErrPasscodeWrong):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Passcode salah. Coba lagi.", "code": "PASSCODE_WRONG"})
	case errors.Is(err, ErrOtpWrong):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Kode OTP salah.", "code": "OTP_WRONG"})
	case errors.Is(err, ErrOtpExpired):
		c.JSON(http.StatusGone, gin.H{"error": "Kode OTP sudah kedaluwarsa. Kirim ulang.", "code": "OTP_EXPIRED"})
	case errors.Is(err, ErrOtpMaxAttempts):
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Terlalu banyak percobaan. Kirim ulang kode OTP.", "code": "OTP_MAX_ATTEMPTS"})
	case errors.Is(err, ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email atau kata sandi tidak cocok. Coba lagi."})
	case errors.Is(err, ErrEmailTaken):
		c.JSON(http.StatusConflict, gin.H{"error": "Email sudah terdaftar. Silakan masuk."})
	case errors.Is(err, ErrEmailNotVerified):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email belum diverifikasi. Silakan verifikasi kode OTP terlebih dahulu."})
	case errors.Is(err, ErrAccountInactive):
		c.JSON(http.StatusForbidden, gin.H{"error": "Akun dinonaktifkan. Hubungi admin toko."})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

// Heartbeat adalah PresenceHandler.Heartbeat lama (handler/presence.go).
func (h *Handler) Heartbeat(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tidak terotentikasi"})
		return
	}

	now := time.Now().UTC()
	if err := h.auth.RecordHeartbeat(c.Request.Context(), claims, now); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal mencatat heartbeat"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// SwitchOutlet adalah OutletHandler.Switch lama (rump handler/outlet.go):
// perpindahan outlet sesi lewat Service.SwitchOutlet.
func (h *Handler) SwitchOutlet(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	var req struct {
		OutletID uint `json:"outletId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.OutletID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Outlet tujuan tidak valid."})
		return
	}
	user, pair, err := h.auth.SwitchOutlet(c.Request.Context(), claims, req.OutletID)
	if err != nil {
		respondOutletErr(c, err)
		return
	}
	c.JSON(http.StatusOK, authResponse{User: *user, TokenPair: *pair})
}

func respondOutletErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrOutletForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "Kasir hanya dapat menggunakan outlet yang ditetapkan."})
	case errors.Is(err, ErrOutletNotFound), errors.Is(err, ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Outlet tidak ditemukan di toko Anda."})
	case errors.Is(err, outlet.ErrOutletDefault):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Outlet default tidak dapat dinonaktifkan."})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}
