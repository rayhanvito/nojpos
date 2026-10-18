-- Migrasi 00002: pengaturan per-outlet (Unit C).
-- Tabel outlet_settings menyimpan subset pengaturan toko yang boleh
-- berbeda per outlet: tarif pajak (tax_pct), footer struk
-- (receipt_footer), dan jam operasional (hours — JSON, representasi
-- sama seperti stores.hours). Kolom override NULLABLE: NULL berarti
-- outlet mewarisi nilai toko. Maksimal satu baris per (store_id,
-- outlet_id), ditegakkan constraint unik di bawah.

-- +goose Up
CREATE SEQUENCE public.outlet_settings_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE public.outlet_settings (
    id bigint NOT NULL DEFAULT nextval('public.outlet_settings_id_seq'::regclass),
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    store_id bigint NOT NULL,
    outlet_id bigint NOT NULL,
    tax_pct numeric,
    receipt_footer text,
    hours text
);

ALTER SEQUENCE public.outlet_settings_id_seq OWNED BY public.outlet_settings.id;

ALTER TABLE ONLY public.outlet_settings
    ADD CONSTRAINT outlet_settings_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.outlet_settings
    ADD CONSTRAINT outlet_settings_store_id_outlet_id_key UNIQUE (store_id, outlet_id);

CREATE INDEX idx_outlet_settings_deleted_at ON public.outlet_settings USING btree (deleted_at);

-- +goose Down
DROP TABLE IF EXISTS public.outlet_settings;
DROP SEQUENCE IF EXISTS public.outlet_settings_id_seq;
