# Migrasi Database NOJPOS (goose)

## Alat dan alasan

Migrasi skema Postgres memakai **goose v3** (`github.com/pressly/goose/v3`)
sebagai library lewat Provider API, dijalankan oleh CLI kecil di
`cmd/migrate`. Satu file SQL per migrasi, dengan anotasi `-- +goose Up`
dan `-- +goose Down`, di-embed ke binary lewat `migrations.go`
(`//go:embed *.sql`).

Alasan: skema production adalah sumber kebenaran. Mulai langkah ini,
perubahan skema **hanya** lewat file migrasi baru yang berversi —
`AutoMigrate` GORM tidak lagi berjalan untuk Postgres saat aplikasi boot
(lihat gerbang di `router/router.go`: AutoMigrate hanya untuk SQLite,
yaitu tes dan dev lokal berbasis file).

Baseline (`00001_baseline.sql`) adalah DDL dari dump skema production
2026-10-17 (`pg_dump --schema-only`), dibersihkan dari artefak pg_dump
(`\restrict`, `SET`, `set_config`, `OWNER TO`, header dump). Skrip
pembangkitnya tersimpan di server di
`/opt/pos-pakesigap/backups/restruktur-20261017/build-baseline.sh`
bersama file acuannya `prod-schema.sql`.

## Menambah migrasi baru

1. Buat file baru di direktori ini: `NNNNN_nama_singkat.sql`
   (nomor urut 5 digit berikutnya, mis. `00002_tambah_kolom_x.sql`).
2. Isi dengan dua bagian:

   ```sql
   -- +goose Up
   ALTER TABLE public.contoh ADD COLUMN kolom_baru text NOT NULL DEFAULT '';

   -- +goose Down
   ALTER TABLE public.contoh DROP COLUMN kolom_baru;
   ```

3. Jangan pernah mengedit migrasi yang sudah terapan — buat migrasi baru.
4. Uji di database scratch dulu (`up`, lalu `down`, lalu `up` lagi),
   baru terapkan ke dev.

## Menjalankan cmd/migrate di dev (VPS 2)

`DATABASE_URL` diambil dari env container backend yang sedang berjalan
(tanpa pernah menulis nilai rahasianya), lalu `cmd/migrate` dijalankan
dari container sementara di network yang sama:

```bash
cd /opt/pos-pakesigap/backend
DATABASE_URL=$(docker inspect -f '{{range .Config.Env}}{{println .}}{{end}}' pos-pakesigap-backend | grep '^DATABASE_URL=' | cut -d= -f2-)
docker run --rm --network pos-pakesigap_default \
  -v /opt/pos-pakesigap/backend:/src -w /src \
  -v /opt/.cache-go-build:/root/.cache/go-build -v /root/go:/go -v /opt/.cache-go-tmp:/gotmp \
  -e GOPATH=/go -e TMPDIR=/gotmp -e GOFLAGS=-mod=mod -e CGO_ENABLED=1 \
  -e DATABASE_URL="$DATABASE_URL" \
  nojpos-gotest:1.26 go run ./cmd/migrate status
```

Ganti `status` dengan `up`, `down`, atau `version` sesuai kebutuhan.

## Stamp baseline untuk database yang sudah ada

Database yang skemanya sudah terbentuk (lewat AutoMigrate lama) **tidak**
menjalankan DDL baseline — ia di-*stamp*: ditandai bahwa versi 1 sudah
terapan.

1. Jalankan `status` sekali (membuat tabel `goose_db_version` bila belum
   ada; baseline akan tampil `pending`):

   ```bash
   # cmd/migrate status, seperti contoh di atas
   ```

2. Sisipkan baris stamp:

   ```bash
   docker exec pos-pakesigap-db psql -U openpos -d openpos \
     -c "INSERT INTO goose_db_version (version_id, is_applied) VALUES (1, true);"
   ```

3. Jalankan `status` lagi — `00001_baseline.sql` harus tampil `applied`,
   dan `up` harus melaporkan tidak ada migrasi tertunda.

Prosedur ini sudah diterapkan ke DB dev. **Production akan di-stamp
saat rilis, hanya atas perintah Rayhan** — jangan stamp production
sendiri.

## Aturan

- Perubahan skema Postgres **hanya** lewat file migrasi baru di sini.
- AutoMigrate untuk Postgres mati; AutoMigrate tetap dipakai SQLite
  (tes + dev lokal file) lewat `db.Migrate`.
- Setiap migrasi wajib punya bagian `Down` yang membatalkan `Up`-nya.
