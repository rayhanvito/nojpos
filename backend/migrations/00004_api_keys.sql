-- API keys per store; the secret itself is never persisted.

-- +goose Up
CREATE SEQUENCE public.api_keys_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE public.api_keys (
    id bigint NOT NULL DEFAULT nextval('public.api_keys_id_seq'::regclass),
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    store_id bigint NOT NULL,
    name text NOT NULL,
    key_prefix varchar(11) NOT NULL,
    key_hash char(64) NOT NULL,
    created_by bigint NOT NULL,
    last_used_at timestamp with time zone
);

ALTER SEQUENCE public.api_keys_id_seq OWNED BY public.api_keys.id;

ALTER TABLE ONLY public.api_keys
    ADD CONSTRAINT api_keys_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.api_keys
    ADD CONSTRAINT api_keys_key_hash_key UNIQUE (key_hash);

CREATE INDEX idx_api_keys_store_id ON public.api_keys USING btree (store_id);
CREATE INDEX idx_api_keys_deleted_at ON public.api_keys USING btree (deleted_at);

-- +goose Down
DROP TABLE IF EXISTS public.api_keys;
DROP SEQUENCE IF EXISTS public.api_keys_id_seq;
