package catalog

import (
	"context"
	"errors"
	"strconv"
	"strings"
)

// F2.8 — Import produk CSV (create-only). Kolom mengikuti format ekspor F0.2 persis:
// id (diabaikan), nama, sku, barcode, kategori, harga beli, harga jual, stok, satuan, aktif.
// Header tanpa kolom id juga diterima. Baris tidak valid DITOLAK dan dilaporkan per baris;
// SKU yang sudah terdaftar (termasuk soft-deleted) atau duplikat di dalam file tidak pernah menimpa data lama.

type ImportRowFailure struct {
	Baris  int    `json:"baris"`
	Alasan string `json:"alasan"`
}

type ImportResult struct {
	Total     int                `json:"total"`
	Berhasil  int                `json:"berhasil"`
	Gagal     int                `json:"gagal"`
	Kesalahan []ImportRowFailure `json:"kesalahan"`
}

const MaxImportRows = 1000

func parseImportInt(raw, field string, emptyAsZero bool) (int64, string) {
	s := strings.TrimSpace(raw)
	if s == "" {
		if emptyAsZero {
			return 0, ""
		}
		return 0, field + " wajib diisi angka"
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, field + " harus bilangan bulat"
	}
	if n < 0 {
		return 0, field + " tidak boleh negatif"
	}
	return n, ""
}

func parseImportAktif(raw string) (bool, string) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "true", "1", "ya":
		return true, ""
	case "false", "0", "tidak":
		return false, ""
	}
	return true, "kolom aktif harus true/false"
}

// records adalah baris data TANPA header, dari file CSV yang header-nya sudah divalidasi handler.
// Nomor baris pada laporan mengikuti nomor baris file (header = baris 1, data pertama = baris 2).
func (s *Service) ImportProducts(ctx context.Context, storeID, outletID uint, actorName string, records [][]string, hasID bool) (*ImportResult, error) {
	width := 9
	offset := 0
	if hasID {
		width = 10
		offset = 1
	}
	cats, err := s.cats.ListByStore(ctx, storeID)
	if err != nil {
		return nil, err
	}
	catByName := make(map[string]uint, len(cats))
	for _, c := range cats {
		catByName[strings.ToLower(strings.TrimSpace(c.Name))] = c.ID
	}
	result := &ImportResult{Kesalahan: []ImportRowFailure{}}
	seenSKU := make(map[string]bool)
	for i, rec := range records {
		baris := i + 2
		// Baris yang seluruhnya kosong dilewati tanpa dihitung.
		blank := true
		for _, cell := range rec {
			if strings.TrimSpace(cell) != "" {
				blank = false
				break
			}
		}
		if blank {
			continue
		}
		result.Total++
		fail := func(alasan string) {
			result.Gagal++
			result.Kesalahan = append(result.Kesalahan, ImportRowFailure{Baris: baris, Alasan: alasan})
		}
		if len(rec) != width {
			fail("jumlah kolom tidak sesuai format (harus " + strconv.Itoa(width) + " kolom)")
			continue
		}
		cell := func(idx int) string { return strings.TrimSpace(rec[offset+idx]) }
		nama, sku, barcode, kategori := cell(0), cell(1), cell(2), cell(3)
		buyPrice, msg := parseImportInt(cell(4), "harga beli", true)
		if msg != "" {
			fail(msg)
			continue
		}
		sellPrice, msg := parseImportInt(cell(5), "harga jual", false)
		if msg != "" {
			fail(msg)
			continue
		}
		stok, msg := parseImportInt(cell(6), "stok", true)
		if msg != "" {
			fail(msg)
			continue
		}
		satuan := cell(7)
		aktif, msg := parseImportAktif(cell(8))
		if msg != "" {
			fail(msg)
			continue
		}
		if nama == "" {
			fail("nama produk wajib diisi")
			continue
		}
		if sku == "" {
			fail("SKU wajib diisi")
			continue
		}
		key := strings.ToLower(sku)
		if seenSKU[key] {
			fail("SKU duplikat di dalam file")
			continue
		}
		seenSKU[key] = true
		var categoryID *uint
		if kategori != "" {
			k := strings.ToLower(kategori)
			if id, ok := catByName[k]; ok {
				categoryID = &id
			} else {
				created, err := s.cats.Create(ctx, storeID, kategori)
				if err != nil {
					fail("gagal membuat kategori: " + kategori)
					continue
				}
				catByName[k] = created.ID
				categoryID = &created.ID
			}
		}
		product, err := s.CreateProduct(ctx, storeID, outletID, actorName, ProductInput{
			Name: nama, SKU: sku, Barcode: barcode, CategoryID: categoryID,
			BuyPrice: buyPrice, SellPrice: sellPrice, Stock: 0, Unit: satuan,
		})
		if err != nil {
			if errors.Is(err, ErrSkuTaken) {
				fail("SKU sudah terdaftar")
			} else {
				fail(err.Error())
			}
			continue
		}
		if stok > 0 {
			if _, err := s.inv.AdjustStock(ctx, storeID, outletID, product.ID, nil, actorName, "plus", stok, "import"); err != nil {
				fail("gagal mencatat stok awal: " + err.Error())
				continue
			}
		}
		if !aktif {
			if err := s.SetProductActive(ctx, storeID, product.ID, false); err != nil {
				fail("gagal menonaktifkan produk: " + err.Error())
				continue
			}
		}
		result.Berhasil++
	}
	return result, nil
}
