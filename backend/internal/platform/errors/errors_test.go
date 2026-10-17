package apierr

import (
	"errors"
	"fmt"
	"testing"
)

func TestNewAndOf(t *testing.T) {
	err := New(409, "DUPLICATE", "data sudah ada")
	if err.Error() != "data sudah ada" {
		t.Fatalf("Error() = %q", err.Error())
	}
	if err.Status != 409 || err.Code != "DUPLICATE" || err.Message != "data sudah ada" {
		t.Fatalf("field salah: %+v", err)
	}
	got, ok := Of(err)
	if !ok || got != err {
		t.Fatalf("Of harus menemukan *Error yang sama, ok=%v", ok)
	}
	if _, ok := Of(errors.New("asing")); ok {
		t.Fatal("Of tidak boleh cocok untuk error asing")
	}
	if _, ok := Of(nil); ok {
		t.Fatal("Of(nil) harus false")
	}
}

func TestOfThroughWrapChain(t *testing.T) {
	base := New(404, "NOT_FOUND", "tidak ditemukan")
	wrapped := fmt.Errorf("konteks: %w", base)
	got, ok := Of(wrapped)
	if !ok || got != base {
		t.Fatalf("Of harus menembus fmt wrap, ok=%v", ok)
	}
	if !errors.Is(wrapped, base) {
		t.Fatal("errors.Is harus tetap cocok lewat rantai wrap")
	}
}

func TestWrap(t *testing.T) {
	cause := errors.New("penyebab asli")
	err := Wrap(cause, 400, "STOCK_INSUFFICIENT")
	if err.Message != "penyebab asli" || err.Status != 400 || err.Code != "STOCK_INSUFFICIENT" {
		t.Fatalf("Wrap salah: %+v", err)
	}
	if !errors.Is(err, cause) {
		t.Fatal("errors.Is harus menemukan cause lewat Unwrap")
	}
	if got := Wrap(nil, 500, "INTERNAL"); got != nil {
		t.Fatalf("Wrap(nil) harus nil, dapat %v", got)
	}
}

func TestSentinelErrorsIsCompatibility(t *testing.T) {
	// Pola sentinel: satu variabel *Error dipakai di banyak tempat;
	// errors.Is(err, sentinel) harus cocok untuk instance yang sama
	// maupun yang dibungkus.
	sentinel := New(409, "SKU_TAKEN", "SKU sudah digunakan di toko ini.")
	if !errors.Is(sentinel, sentinel) {
		t.Fatal("errors.Is(sentinel, sentinel) harus true")
	}
	if !errors.Is(fmt.Errorf("balut: %w", sentinel), sentinel) {
		t.Fatal("errors.Is harus cocok untuk sentinel yang dibungkus")
	}
	other := New(409, "SKU_TAKEN", "SKU sudah digunakan di toko ini.")
	if errors.Is(sentinel, other) {
		t.Fatal("instance berbeda tidak boleh dianggap sama oleh errors.Is")
	}
}
