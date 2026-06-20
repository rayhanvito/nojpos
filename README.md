# NojPOS Final Fix

Monorepo NojPOS untuk Flutter Cashier App dan Laravel API. Admin web dibekukan sampai kontrak Flutter dan backend stabil.

## Struktur

```text
apps/
  cashier/   Flutter Cashier App
  backend/   Laravel API
```

Source of truth produk dan arsitektur: `NOJPOS_POS_FLUTTER_BACKEND_PRD.md`.

## Backend

```bash
cd apps/backend
composer install
cp .env.example .env
php artisan key:generate
php artisan migrate:fresh --seed
php artisan serve --host=0.0.0.0 --port=8000
```

## Cashier

```bash
cd apps/cashier
flutter pub get
flutter run --dart-define=API_BASE_URL=http://10.0.2.2:8000/api/v1
```

## Verification Gates

Backend:

```bash
cd apps/backend
php artisan test
php artisan route:list --path=api/v1
```

Cashier:

```bash
cd apps/cashier
flutter analyze
flutter test
flutter build apk --debug
```
