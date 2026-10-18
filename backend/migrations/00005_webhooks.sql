-- Webhooks per toko: langganan event untuk sistem luar + catatan pengiriman.
-- Secret disimpan apa adanya karena server memerlukannya untuk menandatangani
-- setiap pengiriman (HMAC-SHA256); ia bukan kredensial akses ke NOJPOS.

-- +goose Up
CREATE SEQUENCE public.webhooks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE public.webhooks (
    id bigint NOT NULL DEFAULT nextval('public.webhooks_id_seq'::regclass),
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    store_id bigint NOT NULL,
    url text NOT NULL,
    events text NOT NULL,
    secret text NOT NULL
);

ALTER SEQUENCE public.webhooks_id_seq OWNED BY public.webhooks.id;

ALTER TABLE ONLY public.webhooks
    ADD CONSTRAINT webhooks_pkey PRIMARY KEY (id);

CREATE INDEX idx_webhooks_store_id ON public.webhooks USING btree (store_id);
CREATE INDEX idx_webhooks_deleted_at ON public.webhooks USING btree (deleted_at);

CREATE SEQUENCE public.webhook_deliveries_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE public.webhook_deliveries (
    id bigint NOT NULL DEFAULT nextval('public.webhook_deliveries_id_seq'::regclass),
    created_at timestamp with time zone,
    webhook_id bigint NOT NULL,
    store_id bigint NOT NULL,
    event text NOT NULL,
    payload text NOT NULL,
    response_status integer,
    success boolean NOT NULL DEFAULT false,
    error text NOT NULL DEFAULT ''
);

ALTER SEQUENCE public.webhook_deliveries_id_seq OWNED BY public.webhook_deliveries.id;

ALTER TABLE ONLY public.webhook_deliveries
    ADD CONSTRAINT webhook_deliveries_pkey PRIMARY KEY (id);

CREATE INDEX idx_webhook_deliveries_webhook_id ON public.webhook_deliveries USING btree (webhook_id);
CREATE INDEX idx_webhook_deliveries_store_id ON public.webhook_deliveries USING btree (store_id);

-- +goose Down
DROP TABLE IF EXISTS public.webhook_deliveries;
DROP SEQUENCE IF EXISTS public.webhook_deliveries_id_seq;
DROP TABLE IF EXISTS public.webhooks;
DROP SEQUENCE IF EXISTS public.webhooks_id_seq;
