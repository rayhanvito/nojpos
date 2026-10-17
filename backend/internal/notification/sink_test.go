package notification

import (
	"context"
	"path/filepath"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// sinkBizRow adalah baris "bisnis lain" milik transaksi pemanggil pada
// tes semantik sink §4: ia harus tetap commit walaupun insert
// notifikasi di transaksi yang sama gagal.
type sinkBizRow struct {
	ID   uint   `gorm:"primarykey"`
	Name string `gorm:"not null"`
}

func (sinkBizRow) TableName() string { return "sink_biz_rows" }

// Judul penanda yang insert-nya selalu gagal di basis tes.
const failTitle = "__fail_insert__"

// newSinkTestDB membuka SQLite sementara dengan skema notifikasi +
// baris bisnis, plus trigger yang menggagalkan insert notifikasi
// berjudul failTitle. Trigger RAISE(ABORT) adalah padanan deterministik
// di SQLite untuk kegagalan insert tingkat basis data (mis. pelanggaran
// NOT NULL / constraint di Postgres): kolom string Go tidak dapat
// bernilai NULL lewat struct, jadi pelanggaran disimulasikan pada level
// basis data tempat semantik savepoint benar-benar bekerja.
func newSinkTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "notification-sink.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&Notification{}, &sinkBizRow{}); err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`CREATE TRIGGER fail_notification_insert
		BEFORE INSERT ON notifications
		WHEN NEW.title = '__fail_insert__'
		BEGIN SELECT RAISE(ABORT, 'forced notification insert failure'); END`).Error; err != nil {
		t.Fatal(err)
	}
	return db
}

func countSinkRows(t *testing.T, db *gorm.DB, model any) int64 {
	t.Helper()
	var count int64
	if err := db.Model(model).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	return count
}

// TestNotificationCreateInTxFailureSwallowed membuktikan semantik sink
// §4: di dalam transaksi luar, Create yang insert-nya gagal TIDAK
// mengembalikan error, transaksi luar tetap commit (baris bisnis
// tersimpan), dan baris notifikasi yang gagal tidak ada.
func TestNotificationCreateInTxFailureSwallowed(t *testing.T) {
	db := newSinkTestDB(t)
	repo := NewRepo(db)
	ctx := context.Background()

	var createErr error
	err := repo.WithinTx(ctx, func(ctx context.Context) error {
		if err := repo.DB(ctx).Create(&sinkBizRow{Name: "bisnis"}).Error; err != nil {
			return err
		}
		bad := &Notification{
			StoreID: 1, Title: failTitle, Message: "pasti gagal",
			Category: CategorySistem, Type: "info",
		}
		createErr = repo.Create(ctx, bad)
		return nil
	})
	if err != nil {
		t.Fatalf("transaksi luar harus commit, dapat error: %v", err)
	}
	if createErr != nil {
		t.Fatalf("Create di dalam tx harus menelan error insert, dapat: %v", createErr)
	}
	if got := countSinkRows(t, db, &sinkBizRow{}); got != 1 {
		t.Fatalf("baris bisnis harus tetap tersimpan sesudah commit, dapat %d", got)
	}
	if got := countSinkRows(t, db, &Notification{}); got != 0 {
		t.Fatalf("baris notifikasi yang gagal tidak boleh ada, dapat %d", got)
	}
}

// TestNotificationCreateOutsideTx adalah kontrol di luar transaksi:
// Create valid tersimpan dengan error nil, dan Create yang gagal
// TETAP mengembalikan error (perilaku lama di luar tx tidak berubah).
func TestNotificationCreateOutsideTx(t *testing.T) {
	db := newSinkTestDB(t)
	repo := NewRepo(db)
	ctx := context.Background()

	good := &Notification{
		StoreID: 1, Title: "Notifikasi sah", Message: "halo",
		Category: CategorySistem, Type: "info",
	}
	if err := repo.Create(ctx, good); err != nil {
		t.Fatalf("Create valid di luar tx harus berhasil: %v", err)
	}
	if got := countSinkRows(t, db, &Notification{}); got != 1 {
		t.Fatalf("notifikasi valid harus tersimpan, dapat %d baris", got)
	}

	bad := &Notification{
		StoreID: 1, Title: failTitle, Message: "pasti gagal",
		Category: CategorySistem, Type: "info",
	}
	if err := repo.Create(ctx, bad); err == nil {
		t.Fatal("Create yang gagal di luar tx harus mengembalikan error seperti semula")
	}
	if got := countSinkRows(t, db, &Notification{}); got != 1 {
		t.Fatalf("notifikasi gagal tidak boleh tersimpan, dapat %d baris", got)
	}
}
