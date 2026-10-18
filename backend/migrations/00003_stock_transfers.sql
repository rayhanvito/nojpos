-- Migrasi 00003: transfer stok antar outlet (Unit D).
-- stock_transfers adalah header pemindahan stok antar outlet dalam satu
-- toko; stock_transfer_items adalah baris barangnya. Mengikuti konvensi
-- skema proyek: TIDAK ada constraint foreign key di level database (sama
-- seperti stock_opname_items.session_id di baseline) — relasi ditegakkan
-- layer aplikasi, kolom relasi terindeks.

-- +goose Up
CREATE SEQUENCE public.stock_transfers_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE public.stock_transfers (
    id bigint NOT NULL DEFAULT nextval('public.stock_transfers_id_seq'::regclass),
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    store_id bigint NOT NULL,
    from_outlet_id bigint NOT NULL,
    to_outlet_id bigint NOT NULL,
    note text NOT NULL DEFAULT '',
    created_by bigint NOT NULL DEFAULT 0,
    created_by_name text NOT NULL DEFAULT ''
);

ALTER SEQUENCE public.stock_transfers_id_seq OWNED BY public.stock_transfers.id;

ALTER TABLE ONLY public.stock_transfers
    ADD CONSTRAINT stock_transfers_pkey PRIMARY KEY (id);

CREATE INDEX idx_stock_transfers_store_id ON public.stock_transfers USING btree (store_id);
CREATE INDEX idx_stock_transfers_from_outlet_id ON public.stock_transfers USING btree (from_outlet_id);
CREATE INDEX idx_stock_transfers_to_outlet_id ON public.stock_transfers USING btree (to_outlet_id);
CREATE INDEX idx_stock_transfers_deleted_at ON public.stock_transfers USING btree (deleted_at);

CREATE SEQUENCE public.stock_transfer_items_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE public.stock_transfer_items (
    id bigint NOT NULL DEFAULT nextval('public.stock_transfer_items_id_seq'::regclass),
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    transfer_id bigint NOT NULL,
    product_id bigint NOT NULL,
    variant_id bigint,
    qty bigint NOT NULL
);

ALTER SEQUENCE public.stock_transfer_items_id_seq OWNED BY public.stock_transfer_items.id;

ALTER TABLE ONLY public.stock_transfer_items
    ADD CONSTRAINT stock_transfer_items_pkey PRIMARY KEY (id);

CREATE INDEX idx_stock_transfer_items_transfer_id ON public.stock_transfer_items USING btree (transfer_id);
CREATE INDEX idx_stock_transfer_items_product_id ON public.stock_transfer_items USING btree (product_id);
CREATE INDEX idx_stock_transfer_items_deleted_at ON public.stock_transfer_items USING btree (deleted_at);

-- +goose Down
DROP TABLE IF EXISTS public.stock_transfer_items;
DROP SEQUENCE IF EXISTS public.stock_transfer_items_id_seq;
DROP TABLE IF EXISTS public.stock_transfers;
DROP SEQUENCE IF EXISTS public.stock_transfers_id_seq;
