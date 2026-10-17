// Package apierr adalah tipe error terstruktur NOJPOS: membawa status HTTP,
// kode mesin, dan pesan Bahasa Indonesia yang aman ditampilkan ke pengguna.
// Sentinel lama yang sudah dikonversi adalah *Error, sehingga errors.Is /
// errors.As di seluruh kode yang ada tetap bekerja seperti sebelumnya.
package apierr

import "errors"

// Error adalah error aplikasi dengan status HTTP dan kode mesin.
type Error struct {
	Status  int
	Code    string
	Message string
	cause   error
}

// New membuat *Error baru dengan pesan yang diberikan.
func New(status int, code, message string) *Error {
	return &Error{Status: status, Code: code, Message: message}
}

// Wrap membungkus err yang sudah ada; Message diambil dari err.Error().
func Wrap(err error, status int, code string) *Error {
	if err == nil {
		return nil
	}
	return &Error{Status: status, Code: code, Message: err.Error(), cause: err}
}

// Error mengembalikan pesan yang aman untuk pengguna.
func (e *Error) Error() string { return e.Message }

// Unwrap mengembalikan penyebab yang dibungkus Wrap, bila ada.
func (e *Error) Unwrap() error { return e.cause }

// Of mengekstrak *Error dari err (atau rantai bungkusnya).
func Of(err error) (*Error, bool) {
	var e *Error
	if errors.As(err, &e) {
		return e, true
	}
	return nil, false
}
