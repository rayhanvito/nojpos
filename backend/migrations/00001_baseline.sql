-- Baseline skema NOJPOS.
-- Sumber: pg_dump --schema-only production 2026-10-17
-- File acuan: /opt/pos-pakesigap/backups/restruktur-20261017/prod-schema.sql
-- Dibangkitkan oleh: build-baseline.sh di folder backup yang sama
-- Artefak pg_dump dibuang; lihat build-baseline.sh untuk daftar filternya.

-- +goose Up
--
--






--
-- Name: activity_logs; Type: TABLE; Schema: public; Owner: openpos
--

CREATE TABLE public.activity_logs (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    store_id bigint NOT NULL,
    actor_id text DEFAULT ''::text NOT NULL,
    actor_name text DEFAULT ''::text NOT NULL,
    action text NOT NULL,
    detail text NOT NULL,
    reference_type text DEFAULT ''::text NOT NULL,
    reference_id text DEFAULT ''::text NOT NULL
);



--
-- Name: activity_logs_id_seq; Type: SEQUENCE; Schema: public; Owner: openpos
--

CREATE SEQUENCE public.activity_logs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



--
-- Name: activity_logs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: openpos
--

ALTER SEQUENCE public.activity_logs_id_seq OWNED BY public.activity_logs.id;


--
-- Name: cashier_shifts; Type: TABLE; Schema: public; Owner: openpos
--

CREATE TABLE public.cashier_shifts (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    store_id bigint NOT NULL,
    outlet_id bigint,
    cashier_id bigint NOT NULL,
    cashier_name text NOT NULL,
    started_at timestamp with time zone NOT NULL,
    closed_at timestamp with time zone,
    opening_cash bigint DEFAULT 0 NOT NULL,
    closing_cash_actual bigint,
    expected_cash bigint,
    cash_variance bigint,
    close_note text,
    closed_by_name text
);



--
-- Name: cashier_shifts_id_seq; Type: SEQUENCE; Schema: public; Owner: openpos
--

CREATE SEQUENCE public.cashier_shifts_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



--
-- Name: cashier_shifts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: openpos
--

ALTER SEQUENCE public.cashier_shifts_id_seq OWNED BY public.cashier_shifts.id;


--
-- Name: cashiers; Type: TABLE; Schema: public; Owner: openpos
--

CREATE TABLE public.cashiers (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    store_id bigint NOT NULL,
    outlet_id bigint,
    name text NOT NULL,
    passcode_hash text,
    active boolean DEFAULT true NOT NULL,
    last_seen_at timestamp with time zone
);



--
-- Name: cashiers_id_seq; Type: SEQUENCE; Schema: public; Owner: openpos
--

CREATE SEQUENCE public.cashiers_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



--
-- Name: cashiers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: openpos
--

ALTER SEQUENCE public.cashiers_id_seq OWNED BY public.cashiers.id;


--
-- Name: categories; Type: TABLE; Schema: public; Owner: openpos
--

CREATE TABLE public.categories (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    store_id bigint NOT NULL,
    name text NOT NULL,
    active boolean DEFAULT true NOT NULL
);



--
-- Name: categories_id_seq; Type: SEQUENCE; Schema: public; Owner: openpos
--

CREATE SEQUENCE public.categories_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



--
-- Name: categories_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: openpos
--

ALTER SEQUENCE public.categories_id_seq OWNED BY public.categories.id;


--
-- Name: customers; Type: TABLE; Schema: public; Owner: openpos
--

CREATE TABLE public.customers (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    store_id bigint NOT NULL,
    name text NOT NULL,
    phone text DEFAULT ''::text NOT NULL,
    email text DEFAULT ''::text NOT NULL,
    address text DEFAULT ''::text NOT NULL,
    points bigint DEFAULT 0 NOT NULL,
    CONSTRAINT chk_customers_points CHECK ((points >= 0))
);



--
-- Name: customers_id_seq; Type: SEQUENCE; Schema: public; Owner: openpos
--

CREATE SEQUENCE public.customers_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



--
-- Name: customers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: openpos
--

ALTER SEQUENCE public.customers_id_seq OWNED BY public.customers.id;


--
-- Name: email_otps; Type: TABLE; Schema: public; Owner: openpos
--

CREATE TABLE public.email_otps (
    email text NOT NULL,
    code_hash text NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    attempts bigint DEFAULT 0 NOT NULL,
    last_sent_at timestamp with time zone NOT NULL,
    verified_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL
);



--
-- Name: expenses; Type: TABLE; Schema: public; Owner: openpos
--

CREATE TABLE public.expenses (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    store_id bigint NOT NULL,
    outlet_id bigint,
    date date NOT NULL,
    category text NOT NULL,
    note text DEFAULT ''::text NOT NULL,
    amount bigint NOT NULL,
    method text NOT NULL,
    source text DEFAULT 'outside'::text NOT NULL,
    shift_id bigint,
    created_by_id bigint DEFAULT 0 NOT NULL,
    created_by_name text DEFAULT ''::text NOT NULL
);



--
-- Name: expenses_id_seq; Type: SEQUENCE; Schema: public; Owner: openpos
--

CREATE SEQUENCE public.expenses_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



--
-- Name: expenses_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: openpos
--

ALTER SEQUENCE public.expenses_id_seq OWNED BY public.expenses.id;


--
-- Name: notifications; Type: TABLE; Schema: public; Owner: openpos
--

CREATE TABLE public.notifications (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    store_id bigint NOT NULL,
    dedup_key text,
    title text NOT NULL,
    message text NOT NULL,
    category text DEFAULT 'sistem'::text NOT NULL,
    type text DEFAULT 'info'::text NOT NULL,
    actor_id text DEFAULT ''::text NOT NULL,
    actor_name text DEFAULT ''::text NOT NULL,
    reference_type text DEFAULT ''::text NOT NULL,
    reference_id text DEFAULT ''::text NOT NULL,
    read boolean DEFAULT false NOT NULL
);



--
-- Name: notifications_id_seq; Type: SEQUENCE; Schema: public; Owner: openpos
--

CREATE SEQUENCE public.notifications_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



--
-- Name: notifications_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: openpos
--

ALTER SEQUENCE public.notifications_id_seq OWNED BY public.notifications.id;


--
-- Name: outlets; Type: TABLE; Schema: public; Owner: openpos
--

CREATE TABLE public.outlets (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    store_id bigint NOT NULL,
    name text NOT NULL,
    address text DEFAULT ''::text NOT NULL,
    is_default boolean DEFAULT false NOT NULL
);



--
-- Name: outlets_id_seq; Type: SEQUENCE; Schema: public; Owner: openpos
--

CREATE SEQUENCE public.outlets_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



--
-- Name: outlets_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: openpos
--

ALTER SEQUENCE public.outlets_id_seq OWNED BY public.outlets.id;


--
-- Name: payable_payments; Type: TABLE; Schema: public; Owner: openpos
--

CREATE TABLE public.payable_payments (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    store_id bigint NOT NULL,
    payable_id bigint NOT NULL,
    amount bigint NOT NULL,
    method text NOT NULL,
    created_by text DEFAULT ''::text NOT NULL,
    CONSTRAINT chk_payable_payments_amount CHECK ((amount > 0))
);



--
-- Name: payable_payments_id_seq; Type: SEQUENCE; Schema: public; Owner: openpos
--

CREATE SEQUENCE public.payable_payments_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



--
-- Name: payable_payments_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: openpos
--

ALTER SEQUENCE public.payable_payments_id_seq OWNED BY public.payable_payments.id;


--
-- Name: payables; Type: TABLE; Schema: public; Owner: openpos
--

CREATE TABLE public.payables (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    store_id bigint NOT NULL,
    supplier_id bigint NOT NULL,
    purchase_order_id bigint NOT NULL,
    amount bigint NOT NULL,
    status text DEFAULT 'unpaid'::text NOT NULL,
    CONSTRAINT chk_payables_amount CHECK ((amount >= 0))
);



--
-- Name: payables_id_seq; Type: SEQUENCE; Schema: public; Owner: openpos
--

CREATE SEQUENCE public.payables_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



--
-- Name: payables_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: openpos
--

ALTER SEQUENCE public.payables_id_seq OWNED BY public.payables.id;


--
-- Name: price_tiers; Type: TABLE; Schema: public; Owner: openpos
--

CREATE TABLE public.price_tiers (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    store_id bigint NOT NULL,
    product_id bigint NOT NULL,
    variant_id bigint,
    min_qty bigint NOT NULL,
    sell_price bigint NOT NULL,
    CONSTRAINT chk_price_tiers_min_qty CHECK ((min_qty > 0)),
    CONSTRAINT chk_price_tiers_sell_price CHECK ((sell_price >= 0))
);



--
-- Name: price_tiers_id_seq; Type: SEQUENCE; Schema: public; Owner: openpos
--

CREATE SEQUENCE public.price_tiers_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



--
-- Name: price_tiers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: openpos
--

ALTER SEQUENCE public.price_tiers_id_seq OWNED BY public.price_tiers.id;


--
-- Name: product_stocks; Type: TABLE; Schema: public; Owner: openpos
--

CREATE TABLE public.product_stocks (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    product_id bigint NOT NULL,
    variant_id bigint,
    outlet_id bigint NOT NULL,
    qty bigint DEFAULT 0 NOT NULL,
    CONSTRAINT chk_product_stocks_qty CHECK ((qty >= 0))
);



--
-- Name: product_stocks_id_seq; Type: SEQUENCE; Schema: public; Owner: openpos
--

CREATE SEQUENCE public.product_stocks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



--
-- Name: product_stocks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: openpos
--

ALTER SEQUENCE public.product_stocks_id_seq OWNED BY public.product_stocks.id;


--
-- Name: product_variants; Type: TABLE; Schema: public; Owner: openpos
--

CREATE TABLE public.product_variants (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    store_id bigint NOT NULL,
    outlet_id bigint,
    product_id bigint NOT NULL,
    name text NOT NULL,
    sku text NOT NULL,
    buy_price bigint DEFAULT 0 NOT NULL,
    sell_price bigint DEFAULT 0 NOT NULL,
    stock bigint DEFAULT 0 NOT NULL,
    min_stock bigint DEFAULT 0 NOT NULL,
    CONSTRAINT chk_product_variants_buy_price CHECK ((buy_price >= 0)),
    CONSTRAINT chk_product_variants_min_stock CHECK ((min_stock >= 0)),
    CONSTRAINT chk_product_variants_sell_price CHECK ((sell_price >= 0)),
    CONSTRAINT chk_product_variants_stock CHECK ((stock >= 0))
);



--
-- Name: product_variants_id_seq; Type: SEQUENCE; Schema: public; Owner: openpos
--

CREATE SEQUENCE public.product_variants_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



--
-- Name: product_variants_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: openpos
--

ALTER SEQUENCE public.product_variants_id_seq OWNED BY public.product_variants.id;


--
-- Name: products; Type: TABLE; Schema: public; Owner: openpos
--

CREATE TABLE public.products (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    store_id bigint NOT NULL,
    outlet_id bigint,
    category_id bigint,
    name text NOT NULL,
    sku text NOT NULL,
    barcode text DEFAULT ''::text NOT NULL,
    buy_price bigint DEFAULT 0 NOT NULL,
    sell_price bigint DEFAULT 0 NOT NULL,
    stock bigint DEFAULT 0 NOT NULL,
    min_stock bigint DEFAULT 0 NOT NULL,
    unit text DEFAULT 'pcs'::text NOT NULL,
    active boolean DEFAULT true NOT NULL,
    CONSTRAINT chk_products_buy_price CHECK ((buy_price >= 0)),
    CONSTRAINT chk_products_min_stock CHECK ((min_stock >= 0)),
    CONSTRAINT chk_products_sell_price CHECK ((sell_price >= 0)),
    CONSTRAINT chk_products_stock CHECK ((stock >= 0))
);



--
-- Name: products_id_seq; Type: SEQUENCE; Schema: public; Owner: openpos
--

CREATE SEQUENCE public.products_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



--
-- Name: products_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: openpos
--

ALTER SEQUENCE public.products_id_seq OWNED BY public.products.id;


--
-- Name: promos; Type: TABLE; Schema: public; Owner: openpos
--

CREATE TABLE public.promos (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    store_id bigint NOT NULL,
    name text NOT NULL,
    type text NOT NULL,
    value numeric DEFAULT 0 NOT NULL,
    amount_value bigint DEFAULT 0 NOT NULL,
    scope text NOT NULL,
    category_id bigint,
    product_id bigint,
    voucher_code text,
    starts_at timestamp with time zone NOT NULL,
    ends_at timestamp with time zone NOT NULL,
    happy_start text,
    happy_end text,
    min_purchase bigint DEFAULT 0 NOT NULL,
    max_discount bigint DEFAULT 0 NOT NULL,
    combinable boolean DEFAULT false NOT NULL,
    active boolean DEFAULT true NOT NULL,
    CONSTRAINT chk_promos_amount_value CHECK ((amount_value >= 0)),
    CONSTRAINT chk_promos_max_discount CHECK ((max_discount >= 0)),
    CONSTRAINT chk_promos_min_purchase CHECK ((min_purchase >= 0))
);



--
-- Name: promos_id_seq; Type: SEQUENCE; Schema: public; Owner: openpos
--

CREATE SEQUENCE public.promos_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



--
-- Name: promos_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: openpos
--

ALTER SEQUENCE public.promos_id_seq OWNED BY public.promos.id;


--
-- Name: purchase_order_items; Type: TABLE; Schema: public; Owner: openpos
--

CREATE TABLE public.purchase_order_items (
    id bigint NOT NULL,
    purchase_order_id bigint NOT NULL,
    product_id bigint NOT NULL,
    variant_id bigint,
    qty_ordered bigint NOT NULL,
    qty_received bigint DEFAULT 0 NOT NULL,
    buy_price bigint NOT NULL,
    CONSTRAINT chk_purchase_order_items_buy_price CHECK ((buy_price >= 0)),
    CONSTRAINT chk_purchase_order_items_qty_ordered CHECK ((qty_ordered > 0)),
    CONSTRAINT chk_purchase_order_items_qty_received CHECK ((qty_received >= 0))
);



--
-- Name: purchase_order_items_id_seq; Type: SEQUENCE; Schema: public; Owner: openpos
--

CREATE SEQUENCE public.purchase_order_items_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



--
-- Name: purchase_order_items_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: openpos
--

ALTER SEQUENCE public.purchase_order_items_id_seq OWNED BY public.purchase_order_items.id;


--
-- Name: purchase_orders; Type: TABLE; Schema: public; Owner: openpos
--

CREATE TABLE public.purchase_orders (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    store_id bigint NOT NULL,
    outlet_id bigint,
    supplier_id bigint NOT NULL,
    number text NOT NULL,
    status text DEFAULT 'draft'::text NOT NULL
);



--
-- Name: purchase_orders_id_seq; Type: SEQUENCE; Schema: public; Owner: openpos
--

CREATE SEQUENCE public.purchase_orders_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



--
-- Name: purchase_orders_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: openpos
--

ALTER SEQUENCE public.purchase_orders_id_seq OWNED BY public.purchase_orders.id;


--
-- Name: receivable_payments; Type: TABLE; Schema: public; Owner: openpos
--

CREATE TABLE public.receivable_payments (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    store_id bigint NOT NULL,
    receivable_id bigint NOT NULL,
    amount bigint NOT NULL,
    method text NOT NULL,
    created_by text DEFAULT ''::text NOT NULL,
    CONSTRAINT chk_receivable_payments_amount CHECK ((amount > 0))
);



--
-- Name: receivable_payments_id_seq; Type: SEQUENCE; Schema: public; Owner: openpos
--

CREATE SEQUENCE public.receivable_payments_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



--
-- Name: receivable_payments_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: openpos
--

ALTER SEQUENCE public.receivable_payments_id_seq OWNED BY public.receivable_payments.id;


--
-- Name: receivables; Type: TABLE; Schema: public; Owner: openpos
--

CREATE TABLE public.receivables (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    store_id bigint NOT NULL,
    customer_id bigint NOT NULL,
    transaction_id bigint NOT NULL,
    amount bigint NOT NULL,
    due_date timestamp with time zone NOT NULL,
    status text DEFAULT 'unpaid'::text NOT NULL,
    CONSTRAINT chk_receivables_amount CHECK ((amount >= 0))
);



--
-- Name: receivables_id_seq; Type: SEQUENCE; Schema: public; Owner: openpos
--

CREATE SEQUENCE public.receivables_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



--
-- Name: receivables_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: openpos
--

ALTER SEQUENCE public.receivables_id_seq OWNED BY public.receivables.id;


--
-- Name: refresh_tokens; Type: TABLE; Schema: public; Owner: openpos
--

CREATE TABLE public.refresh_tokens (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    user_id bigint NOT NULL,
    outlet_id bigint,
    token_hash text NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    revoked boolean DEFAULT false NOT NULL,
    acting_as_cashier_id bigint
);



--
-- Name: refresh_tokens_id_seq; Type: SEQUENCE; Schema: public; Owner: openpos
--

CREATE SEQUENCE public.refresh_tokens_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



--
-- Name: refresh_tokens_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: openpos
--

ALTER SEQUENCE public.refresh_tokens_id_seq OWNED BY public.refresh_tokens.id;


--
-- Name: refunds; Type: TABLE; Schema: public; Owner: openpos
--

CREATE TABLE public.refunds (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    store_id bigint NOT NULL,
    trx_id bigint NOT NULL,
    items text,
    variant_items text,
    reason text NOT NULL,
    by_name text NOT NULL
);



--
-- Name: refunds_id_seq; Type: SEQUENCE; Schema: public; Owner: openpos
--

CREATE SEQUENCE public.refunds_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



--
-- Name: refunds_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: openpos
--

ALTER SEQUENCE public.refunds_id_seq OWNED BY public.refunds.id;


--
-- Name: stock_movements; Type: TABLE; Schema: public; Owner: openpos
--

CREATE TABLE public.stock_movements (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    store_id bigint NOT NULL,
    outlet_id bigint,
    product_id bigint NOT NULL,
    variant_id bigint,
    type text NOT NULL,
    qty bigint NOT NULL,
    reason text DEFAULT ''::text NOT NULL,
    actor text DEFAULT ''::text NOT NULL
);



--
-- Name: stock_movements_id_seq; Type: SEQUENCE; Schema: public; Owner: openpos
--

CREATE SEQUENCE public.stock_movements_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



--
-- Name: stock_movements_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: openpos
--

ALTER SEQUENCE public.stock_movements_id_seq OWNED BY public.stock_movements.id;


--
-- Name: stock_opname_items; Type: TABLE; Schema: public; Owner: openpos
--

CREATE TABLE public.stock_opname_items (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    session_id bigint NOT NULL,
    product_id bigint NOT NULL,
    variant_id bigint,
    system_stock bigint NOT NULL,
    counted_stock bigint,
    CONSTRAINT chk_stock_opname_items_counted_stock CHECK (((counted_stock IS NULL) OR (counted_stock >= 0))),
    CONSTRAINT chk_stock_opname_items_system_stock CHECK ((system_stock >= 0))
);



--
-- Name: stock_opname_items_id_seq; Type: SEQUENCE; Schema: public; Owner: openpos
--

CREATE SEQUENCE public.stock_opname_items_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



--
-- Name: stock_opname_items_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: openpos
--

ALTER SEQUENCE public.stock_opname_items_id_seq OWNED BY public.stock_opname_items.id;


--
-- Name: stock_opname_sessions; Type: TABLE; Schema: public; Owner: openpos
--

CREATE TABLE public.stock_opname_sessions (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    store_id bigint NOT NULL,
    outlet_id bigint,
    status text DEFAULT 'counting'::text NOT NULL,
    note text DEFAULT ''::text NOT NULL,
    created_by text DEFAULT ''::text NOT NULL
);



--
-- Name: stock_opname_sessions_id_seq; Type: SEQUENCE; Schema: public; Owner: openpos
--

CREATE SEQUENCE public.stock_opname_sessions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



--
-- Name: stock_opname_sessions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: openpos
--

ALTER SEQUENCE public.stock_opname_sessions_id_seq OWNED BY public.stock_opname_sessions.id;


--
-- Name: stores; Type: TABLE; Schema: public; Owner: openpos
--

CREATE TABLE public.stores (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text NOT NULL,
    address text DEFAULT ''::text NOT NULL,
    phone text DEFAULT ''::text NOT NULL,
    tax_enabled boolean DEFAULT false NOT NULL,
    tax_pct numeric DEFAULT 0 NOT NULL,
    receipt_header text DEFAULT 'Terima kasih sudah berbelanja'::text NOT NULL,
    receipt_footer text DEFAULT 'Barang yang sudah dibeli tidak dapat ditukar'::text NOT NULL,
    paper text DEFAULT '58mm'::text NOT NULL,
    timezone text DEFAULT 'Asia/Makassar'::text NOT NULL,
    business_type text DEFAULT ''::text NOT NULL,
    email text DEFAULT ''::text NOT NULL,
    city text DEFAULT ''::text NOT NULL,
    province text DEFAULT ''::text NOT NULL,
    currency text DEFAULT 'IDR'::text NOT NULL,
    hours text DEFAULT '[]'::text NOT NULL,
    receipt_show_logo boolean DEFAULT true NOT NULL,
    receipt_show_cashier boolean DEFAULT true NOT NULL,
    receipt_show_method boolean DEFAULT true NOT NULL,
    receipt_show_tax boolean DEFAULT true NOT NULL,
    receipt_show_discount boolean DEFAULT true NOT NULL,
    receipt_show_note boolean DEFAULT true NOT NULL,
    tax_name text DEFAULT ''::text NOT NULL,
    tax_inclusive boolean DEFAULT false NOT NULL,
    tax_rounding text DEFAULT 'none'::text NOT NULL,
    tax_apply_to text DEFAULT 'all'::text NOT NULL,
    loyalty_enabled boolean DEFAULT false NOT NULL,
    loyalty_earn_per bigint DEFAULT 10000 NOT NULL,
    loyalty_point_value bigint DEFAULT 100 NOT NULL,
    credit_due_days bigint DEFAULT 14 NOT NULL,
    low_stock_default bigint DEFAULT 0 NOT NULL,
    CONSTRAINT chk_stores_credit_due_days CHECK ((credit_due_days >= 0)),
    CONSTRAINT chk_stores_low_stock_default CHECK ((low_stock_default >= 0))
);



--
-- Name: stores_id_seq; Type: SEQUENCE; Schema: public; Owner: openpos
--

CREATE SEQUENCE public.stores_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



--
-- Name: stores_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: openpos
--

ALTER SEQUENCE public.stores_id_seq OWNED BY public.stores.id;


--
-- Name: suppliers; Type: TABLE; Schema: public; Owner: openpos
--

CREATE TABLE public.suppliers (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    store_id bigint NOT NULL,
    name text NOT NULL,
    contact text DEFAULT ''::text NOT NULL,
    address text DEFAULT ''::text NOT NULL
);



--
-- Name: suppliers_id_seq; Type: SEQUENCE; Schema: public; Owner: openpos
--

CREATE SEQUENCE public.suppliers_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



--
-- Name: suppliers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: openpos
--

ALTER SEQUENCE public.suppliers_id_seq OWNED BY public.suppliers.id;


--
-- Name: transaction_items; Type: TABLE; Schema: public; Owner: openpos
--

CREATE TABLE public.transaction_items (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    trx_id bigint NOT NULL,
    product_id bigint NOT NULL,
    variant_id bigint,
    variant_name text DEFAULT ''::text NOT NULL,
    name text NOT NULL,
    buy_price bigint NOT NULL,
    price bigint NOT NULL,
    qty bigint NOT NULL,
    CONSTRAINT chk_transaction_items_qty CHECK ((qty > 0))
);



--
-- Name: transaction_items_id_seq; Type: SEQUENCE; Schema: public; Owner: openpos
--

CREATE SEQUENCE public.transaction_items_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



--
-- Name: transaction_items_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: openpos
--

ALTER SEQUENCE public.transaction_items_id_seq OWNED BY public.transaction_items.id;


--
-- Name: transactions; Type: TABLE; Schema: public; Owner: openpos
--

CREATE TABLE public.transactions (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    store_id bigint NOT NULL,
    outlet_id bigint,
    client_tx_id text DEFAULT ''::text NOT NULL,
    cashier_id bigint NOT NULL,
    cashier_name text NOT NULL,
    items text,
    subtotal bigint NOT NULL,
    discount bigint DEFAULT 0 NOT NULL,
    promo_id bigint,
    promo_name text DEFAULT ''::text NOT NULL,
    promo_discount bigint DEFAULT 0 NOT NULL,
    voucher_id bigint,
    voucher_code text DEFAULT ''::text NOT NULL,
    voucher_discount bigint DEFAULT 0 NOT NULL,
    tax bigint DEFAULT 0 NOT NULL,
    total bigint NOT NULL,
    method text NOT NULL,
    paid bigint NOT NULL,
    change bigint DEFAULT 0 NOT NULL,
    status text DEFAULT 'completed'::text NOT NULL,
    customer text DEFAULT ''::text NOT NULL,
    customer_id bigint,
    points_earned bigint DEFAULT 0 NOT NULL,
    points_redeemed bigint DEFAULT 0 NOT NULL,
    points_discount bigint DEFAULT 0 NOT NULL,
    CONSTRAINT chk_transactions_change CHECK ((change >= 0)),
    CONSTRAINT chk_transactions_discount CHECK ((discount >= 0)),
    CONSTRAINT chk_transactions_paid CHECK ((paid >= 0)),
    CONSTRAINT chk_transactions_points_discount CHECK ((points_discount >= 0)),
    CONSTRAINT chk_transactions_points_earned CHECK ((points_earned >= 0)),
    CONSTRAINT chk_transactions_points_redeemed CHECK ((points_redeemed >= 0)),
    CONSTRAINT chk_transactions_promo_discount CHECK ((promo_discount >= 0)),
    CONSTRAINT chk_transactions_subtotal CHECK ((subtotal >= 0)),
    CONSTRAINT chk_transactions_tax CHECK ((tax >= 0)),
    CONSTRAINT chk_transactions_total CHECK ((total >= 0)),
    CONSTRAINT chk_transactions_voucher_discount CHECK ((voucher_discount >= 0))
);



--
-- Name: transactions_id_seq; Type: SEQUENCE; Schema: public; Owner: openpos
--

CREATE SEQUENCE public.transactions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



--
-- Name: transactions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: openpos
--

ALTER SEQUENCE public.transactions_id_seq OWNED BY public.transactions.id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: openpos
--

CREATE TABLE public.users (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    store_id bigint NOT NULL,
    outlet_id bigint,
    email text NOT NULL,
    name text NOT NULL,
    password_hash text NOT NULL,
    passcode_hash text,
    role text DEFAULT 'cashier'::text NOT NULL,
    active boolean DEFAULT true NOT NULL,
    last_seen_at timestamp with time zone,
    email_verified_at timestamp with time zone
);



--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: openpos
--

CREATE SEQUENCE public.users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: openpos
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: activity_logs id; Type: DEFAULT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.activity_logs ALTER COLUMN id SET DEFAULT nextval('public.activity_logs_id_seq'::regclass);


--
-- Name: cashier_shifts id; Type: DEFAULT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.cashier_shifts ALTER COLUMN id SET DEFAULT nextval('public.cashier_shifts_id_seq'::regclass);


--
-- Name: cashiers id; Type: DEFAULT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.cashiers ALTER COLUMN id SET DEFAULT nextval('public.cashiers_id_seq'::regclass);


--
-- Name: categories id; Type: DEFAULT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.categories ALTER COLUMN id SET DEFAULT nextval('public.categories_id_seq'::regclass);


--
-- Name: customers id; Type: DEFAULT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.customers ALTER COLUMN id SET DEFAULT nextval('public.customers_id_seq'::regclass);


--
-- Name: expenses id; Type: DEFAULT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.expenses ALTER COLUMN id SET DEFAULT nextval('public.expenses_id_seq'::regclass);


--
-- Name: notifications id; Type: DEFAULT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.notifications ALTER COLUMN id SET DEFAULT nextval('public.notifications_id_seq'::regclass);


--
-- Name: outlets id; Type: DEFAULT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.outlets ALTER COLUMN id SET DEFAULT nextval('public.outlets_id_seq'::regclass);


--
-- Name: payable_payments id; Type: DEFAULT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.payable_payments ALTER COLUMN id SET DEFAULT nextval('public.payable_payments_id_seq'::regclass);


--
-- Name: payables id; Type: DEFAULT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.payables ALTER COLUMN id SET DEFAULT nextval('public.payables_id_seq'::regclass);


--
-- Name: price_tiers id; Type: DEFAULT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.price_tiers ALTER COLUMN id SET DEFAULT nextval('public.price_tiers_id_seq'::regclass);


--
-- Name: product_stocks id; Type: DEFAULT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.product_stocks ALTER COLUMN id SET DEFAULT nextval('public.product_stocks_id_seq'::regclass);


--
-- Name: product_variants id; Type: DEFAULT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.product_variants ALTER COLUMN id SET DEFAULT nextval('public.product_variants_id_seq'::regclass);


--
-- Name: products id; Type: DEFAULT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.products ALTER COLUMN id SET DEFAULT nextval('public.products_id_seq'::regclass);


--
-- Name: promos id; Type: DEFAULT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.promos ALTER COLUMN id SET DEFAULT nextval('public.promos_id_seq'::regclass);


--
-- Name: purchase_order_items id; Type: DEFAULT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.purchase_order_items ALTER COLUMN id SET DEFAULT nextval('public.purchase_order_items_id_seq'::regclass);


--
-- Name: purchase_orders id; Type: DEFAULT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.purchase_orders ALTER COLUMN id SET DEFAULT nextval('public.purchase_orders_id_seq'::regclass);


--
-- Name: receivable_payments id; Type: DEFAULT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.receivable_payments ALTER COLUMN id SET DEFAULT nextval('public.receivable_payments_id_seq'::regclass);


--
-- Name: receivables id; Type: DEFAULT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.receivables ALTER COLUMN id SET DEFAULT nextval('public.receivables_id_seq'::regclass);


--
-- Name: refresh_tokens id; Type: DEFAULT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.refresh_tokens ALTER COLUMN id SET DEFAULT nextval('public.refresh_tokens_id_seq'::regclass);


--
-- Name: refunds id; Type: DEFAULT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.refunds ALTER COLUMN id SET DEFAULT nextval('public.refunds_id_seq'::regclass);


--
-- Name: stock_movements id; Type: DEFAULT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.stock_movements ALTER COLUMN id SET DEFAULT nextval('public.stock_movements_id_seq'::regclass);


--
-- Name: stock_opname_items id; Type: DEFAULT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.stock_opname_items ALTER COLUMN id SET DEFAULT nextval('public.stock_opname_items_id_seq'::regclass);


--
-- Name: stock_opname_sessions id; Type: DEFAULT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.stock_opname_sessions ALTER COLUMN id SET DEFAULT nextval('public.stock_opname_sessions_id_seq'::regclass);


--
-- Name: stores id; Type: DEFAULT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.stores ALTER COLUMN id SET DEFAULT nextval('public.stores_id_seq'::regclass);


--
-- Name: suppliers id; Type: DEFAULT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.suppliers ALTER COLUMN id SET DEFAULT nextval('public.suppliers_id_seq'::regclass);


--
-- Name: transaction_items id; Type: DEFAULT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.transaction_items ALTER COLUMN id SET DEFAULT nextval('public.transaction_items_id_seq'::regclass);


--
-- Name: transactions id; Type: DEFAULT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.transactions ALTER COLUMN id SET DEFAULT nextval('public.transactions_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Name: activity_logs activity_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.activity_logs
    ADD CONSTRAINT activity_logs_pkey PRIMARY KEY (id);


--
-- Name: cashier_shifts cashier_shifts_pkey; Type: CONSTRAINT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.cashier_shifts
    ADD CONSTRAINT cashier_shifts_pkey PRIMARY KEY (id);


--
-- Name: cashiers cashiers_pkey; Type: CONSTRAINT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.cashiers
    ADD CONSTRAINT cashiers_pkey PRIMARY KEY (id);


--
-- Name: categories categories_pkey; Type: CONSTRAINT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.categories
    ADD CONSTRAINT categories_pkey PRIMARY KEY (id);


--
-- Name: customers customers_pkey; Type: CONSTRAINT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.customers
    ADD CONSTRAINT customers_pkey PRIMARY KEY (id);


--
-- Name: email_otps email_otps_pkey; Type: CONSTRAINT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.email_otps
    ADD CONSTRAINT email_otps_pkey PRIMARY KEY (email);


--
-- Name: expenses expenses_pkey; Type: CONSTRAINT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.expenses
    ADD CONSTRAINT expenses_pkey PRIMARY KEY (id);


--
-- Name: notifications notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (id);


--
-- Name: outlets outlets_pkey; Type: CONSTRAINT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.outlets
    ADD CONSTRAINT outlets_pkey PRIMARY KEY (id);


--
-- Name: payable_payments payable_payments_pkey; Type: CONSTRAINT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.payable_payments
    ADD CONSTRAINT payable_payments_pkey PRIMARY KEY (id);


--
-- Name: payables payables_pkey; Type: CONSTRAINT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.payables
    ADD CONSTRAINT payables_pkey PRIMARY KEY (id);


--
-- Name: price_tiers price_tiers_pkey; Type: CONSTRAINT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.price_tiers
    ADD CONSTRAINT price_tiers_pkey PRIMARY KEY (id);


--
-- Name: product_stocks product_stocks_pkey; Type: CONSTRAINT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.product_stocks
    ADD CONSTRAINT product_stocks_pkey PRIMARY KEY (id);


--
-- Name: product_variants product_variants_pkey; Type: CONSTRAINT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.product_variants
    ADD CONSTRAINT product_variants_pkey PRIMARY KEY (id);


--
-- Name: products products_pkey; Type: CONSTRAINT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_pkey PRIMARY KEY (id);


--
-- Name: promos promos_pkey; Type: CONSTRAINT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.promos
    ADD CONSTRAINT promos_pkey PRIMARY KEY (id);


--
-- Name: purchase_order_items purchase_order_items_pkey; Type: CONSTRAINT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.purchase_order_items
    ADD CONSTRAINT purchase_order_items_pkey PRIMARY KEY (id);


--
-- Name: purchase_orders purchase_orders_pkey; Type: CONSTRAINT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.purchase_orders
    ADD CONSTRAINT purchase_orders_pkey PRIMARY KEY (id);


--
-- Name: receivable_payments receivable_payments_pkey; Type: CONSTRAINT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.receivable_payments
    ADD CONSTRAINT receivable_payments_pkey PRIMARY KEY (id);


--
-- Name: receivables receivables_pkey; Type: CONSTRAINT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.receivables
    ADD CONSTRAINT receivables_pkey PRIMARY KEY (id);


--
-- Name: refresh_tokens refresh_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.refresh_tokens
    ADD CONSTRAINT refresh_tokens_pkey PRIMARY KEY (id);


--
-- Name: refunds refunds_pkey; Type: CONSTRAINT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.refunds
    ADD CONSTRAINT refunds_pkey PRIMARY KEY (id);


--
-- Name: stock_movements stock_movements_pkey; Type: CONSTRAINT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.stock_movements
    ADD CONSTRAINT stock_movements_pkey PRIMARY KEY (id);


--
-- Name: stock_opname_items stock_opname_items_pkey; Type: CONSTRAINT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.stock_opname_items
    ADD CONSTRAINT stock_opname_items_pkey PRIMARY KEY (id);


--
-- Name: stock_opname_sessions stock_opname_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.stock_opname_sessions
    ADD CONSTRAINT stock_opname_sessions_pkey PRIMARY KEY (id);


--
-- Name: stores stores_pkey; Type: CONSTRAINT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.stores
    ADD CONSTRAINT stores_pkey PRIMARY KEY (id);


--
-- Name: suppliers suppliers_pkey; Type: CONSTRAINT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.suppliers
    ADD CONSTRAINT suppliers_pkey PRIMARY KEY (id);


--
-- Name: transaction_items transaction_items_pkey; Type: CONSTRAINT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.transaction_items
    ADD CONSTRAINT transaction_items_pkey PRIMARY KEY (id);


--
-- Name: transactions transactions_pkey; Type: CONSTRAINT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_pkey PRIMARY KEY (id);


--
-- Name: refresh_tokens uni_refresh_tokens_token_hash; Type: CONSTRAINT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.refresh_tokens
    ADD CONSTRAINT uni_refresh_tokens_token_hash UNIQUE (token_hash);


--
-- Name: users uni_users_email; Type: CONSTRAINT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT uni_users_email UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: openpos
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: idx_activity_logs_action; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_activity_logs_action ON public.activity_logs USING btree (action);


--
-- Name: idx_activity_logs_deleted_at; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_activity_logs_deleted_at ON public.activity_logs USING btree (deleted_at);


--
-- Name: idx_activity_logs_store_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_activity_logs_store_id ON public.activity_logs USING btree (store_id);


--
-- Name: idx_cashier_shifts_cashier_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_cashier_shifts_cashier_id ON public.cashier_shifts USING btree (cashier_id);


--
-- Name: idx_cashier_shifts_deleted_at; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_cashier_shifts_deleted_at ON public.cashier_shifts USING btree (deleted_at);


--
-- Name: idx_cashier_shifts_outlet_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_cashier_shifts_outlet_id ON public.cashier_shifts USING btree (outlet_id);


--
-- Name: idx_cashier_shifts_store_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_cashier_shifts_store_id ON public.cashier_shifts USING btree (store_id);


--
-- Name: idx_cashiers_deleted_at; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_cashiers_deleted_at ON public.cashiers USING btree (deleted_at);


--
-- Name: idx_cashiers_last_seen_at; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_cashiers_last_seen_at ON public.cashiers USING btree (last_seen_at);


--
-- Name: idx_cashiers_outlet_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_cashiers_outlet_id ON public.cashiers USING btree (outlet_id);


--
-- Name: idx_cashiers_store_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_cashiers_store_id ON public.cashiers USING btree (store_id);


--
-- Name: idx_categories_deleted_at; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_categories_deleted_at ON public.categories USING btree (deleted_at);


--
-- Name: idx_categories_store_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_categories_store_id ON public.categories USING btree (store_id);


--
-- Name: idx_customers_deleted_at; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_customers_deleted_at ON public.customers USING btree (deleted_at);


--
-- Name: idx_email_otps_expires_at; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_email_otps_expires_at ON public.email_otps USING btree (expires_at);


--
-- Name: idx_expenses_date; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_expenses_date ON public.expenses USING btree (date);


--
-- Name: idx_expenses_deleted_at; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_expenses_deleted_at ON public.expenses USING btree (deleted_at);


--
-- Name: idx_expenses_outlet_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_expenses_outlet_id ON public.expenses USING btree (outlet_id);


--
-- Name: idx_expenses_shift_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_expenses_shift_id ON public.expenses USING btree (shift_id);


--
-- Name: idx_expenses_store_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_expenses_store_id ON public.expenses USING btree (store_id);


--
-- Name: idx_notifications_deleted_at; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_notifications_deleted_at ON public.notifications USING btree (deleted_at);


--
-- Name: idx_notifications_store_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_notifications_store_id ON public.notifications USING btree (store_id);


--
-- Name: idx_outlets_deleted_at; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_outlets_deleted_at ON public.outlets USING btree (deleted_at);


--
-- Name: idx_outlets_store_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_outlets_store_id ON public.outlets USING btree (store_id);


--
-- Name: idx_payable_payments_deleted_at; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_payable_payments_deleted_at ON public.payable_payments USING btree (deleted_at);


--
-- Name: idx_payable_payments_payable_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_payable_payments_payable_id ON public.payable_payments USING btree (payable_id);


--
-- Name: idx_payable_payments_store_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_payable_payments_store_id ON public.payable_payments USING btree (store_id);


--
-- Name: idx_payables_deleted_at; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_payables_deleted_at ON public.payables USING btree (deleted_at);


--
-- Name: idx_payables_purchase_order_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_payables_purchase_order_id ON public.payables USING btree (purchase_order_id);


--
-- Name: idx_payables_status; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_payables_status ON public.payables USING btree (status);


--
-- Name: idx_payables_store_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_payables_store_id ON public.payables USING btree (store_id);


--
-- Name: idx_payables_supplier_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_payables_supplier_id ON public.payables USING btree (supplier_id);


--
-- Name: idx_price_tiers_deleted_at; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_price_tiers_deleted_at ON public.price_tiers USING btree (deleted_at);


--
-- Name: idx_price_tiers_product_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_price_tiers_product_id ON public.price_tiers USING btree (product_id);


--
-- Name: idx_price_tiers_store_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_price_tiers_store_id ON public.price_tiers USING btree (store_id);


--
-- Name: idx_price_tiers_variant_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_price_tiers_variant_id ON public.price_tiers USING btree (variant_id);


--
-- Name: idx_product_stocks_deleted_at; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_product_stocks_deleted_at ON public.product_stocks USING btree (deleted_at);


--
-- Name: idx_product_stocks_outlet_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_product_stocks_outlet_id ON public.product_stocks USING btree (outlet_id);


--
-- Name: idx_product_stocks_variant_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_product_stocks_variant_id ON public.product_stocks USING btree (variant_id);


--
-- Name: idx_product_variants_deleted_at; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_product_variants_deleted_at ON public.product_variants USING btree (deleted_at);


--
-- Name: idx_product_variants_outlet_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_product_variants_outlet_id ON public.product_variants USING btree (outlet_id);


--
-- Name: idx_product_variants_product_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_product_variants_product_id ON public.product_variants USING btree (product_id);


--
-- Name: idx_product_variants_store_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_product_variants_store_id ON public.product_variants USING btree (store_id);


--
-- Name: idx_products_category_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_products_category_id ON public.products USING btree (category_id);


--
-- Name: idx_products_deleted_at; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_products_deleted_at ON public.products USING btree (deleted_at);


--
-- Name: idx_products_outlet_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_products_outlet_id ON public.products USING btree (outlet_id);


--
-- Name: idx_products_store_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_products_store_id ON public.products USING btree (store_id);


--
-- Name: idx_promos_category_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_promos_category_id ON public.promos USING btree (category_id);


--
-- Name: idx_promos_deleted_at; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_promos_deleted_at ON public.promos USING btree (deleted_at);


--
-- Name: idx_promos_ends_at; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_promos_ends_at ON public.promos USING btree (ends_at);


--
-- Name: idx_promos_product_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_promos_product_id ON public.promos USING btree (product_id);


--
-- Name: idx_promos_starts_at; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_promos_starts_at ON public.promos USING btree (starts_at);


--
-- Name: idx_promos_store_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_promos_store_id ON public.promos USING btree (store_id);


--
-- Name: idx_purchase_order_items_product_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_purchase_order_items_product_id ON public.purchase_order_items USING btree (product_id);


--
-- Name: idx_purchase_order_items_purchase_order_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_purchase_order_items_purchase_order_id ON public.purchase_order_items USING btree (purchase_order_id);


--
-- Name: idx_purchase_order_items_variant_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_purchase_order_items_variant_id ON public.purchase_order_items USING btree (variant_id);


--
-- Name: idx_purchase_orders_deleted_at; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_purchase_orders_deleted_at ON public.purchase_orders USING btree (deleted_at);


--
-- Name: idx_purchase_orders_outlet_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_purchase_orders_outlet_id ON public.purchase_orders USING btree (outlet_id);


--
-- Name: idx_purchase_orders_status; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_purchase_orders_status ON public.purchase_orders USING btree (status);


--
-- Name: idx_purchase_orders_store_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_purchase_orders_store_id ON public.purchase_orders USING btree (store_id);


--
-- Name: idx_purchase_orders_supplier_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_purchase_orders_supplier_id ON public.purchase_orders USING btree (supplier_id);


--
-- Name: idx_receivable_payments_deleted_at; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_receivable_payments_deleted_at ON public.receivable_payments USING btree (deleted_at);


--
-- Name: idx_receivable_payments_receivable_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_receivable_payments_receivable_id ON public.receivable_payments USING btree (receivable_id);


--
-- Name: idx_receivable_payments_store_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_receivable_payments_store_id ON public.receivable_payments USING btree (store_id);


--
-- Name: idx_receivables_customer_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_receivables_customer_id ON public.receivables USING btree (customer_id);


--
-- Name: idx_receivables_deleted_at; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_receivables_deleted_at ON public.receivables USING btree (deleted_at);


--
-- Name: idx_receivables_due_date; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_receivables_due_date ON public.receivables USING btree (due_date);


--
-- Name: idx_receivables_status; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_receivables_status ON public.receivables USING btree (status);


--
-- Name: idx_receivables_store_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_receivables_store_id ON public.receivables USING btree (store_id);


--
-- Name: idx_receivables_transaction_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE UNIQUE INDEX idx_receivables_transaction_id ON public.receivables USING btree (transaction_id);


--
-- Name: idx_refresh_tokens_acting_as_cashier_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_refresh_tokens_acting_as_cashier_id ON public.refresh_tokens USING btree (acting_as_cashier_id);


--
-- Name: idx_refresh_tokens_deleted_at; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_refresh_tokens_deleted_at ON public.refresh_tokens USING btree (deleted_at);


--
-- Name: idx_refresh_tokens_outlet_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_refresh_tokens_outlet_id ON public.refresh_tokens USING btree (outlet_id);


--
-- Name: idx_refresh_tokens_user_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_refresh_tokens_user_id ON public.refresh_tokens USING btree (user_id);


--
-- Name: idx_refunds_deleted_at; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_refunds_deleted_at ON public.refunds USING btree (deleted_at);


--
-- Name: idx_refunds_trx_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_refunds_trx_id ON public.refunds USING btree (trx_id);


--
-- Name: idx_stock_movements_deleted_at; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_stock_movements_deleted_at ON public.stock_movements USING btree (deleted_at);


--
-- Name: idx_stock_movements_outlet_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_stock_movements_outlet_id ON public.stock_movements USING btree (outlet_id);


--
-- Name: idx_stock_movements_product_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_stock_movements_product_id ON public.stock_movements USING btree (product_id);


--
-- Name: idx_stock_movements_store_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_stock_movements_store_id ON public.stock_movements USING btree (store_id);


--
-- Name: idx_stock_movements_variant_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_stock_movements_variant_id ON public.stock_movements USING btree (variant_id);


--
-- Name: idx_stock_opname_items_deleted_at; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_stock_opname_items_deleted_at ON public.stock_opname_items USING btree (deleted_at);


--
-- Name: idx_stock_opname_items_product_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_stock_opname_items_product_id ON public.stock_opname_items USING btree (product_id);


--
-- Name: idx_stock_opname_items_session_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_stock_opname_items_session_id ON public.stock_opname_items USING btree (session_id);


--
-- Name: idx_stock_opname_items_variant_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_stock_opname_items_variant_id ON public.stock_opname_items USING btree (variant_id);


--
-- Name: idx_stock_opname_sessions_deleted_at; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_stock_opname_sessions_deleted_at ON public.stock_opname_sessions USING btree (deleted_at);


--
-- Name: idx_stock_opname_sessions_outlet_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_stock_opname_sessions_outlet_id ON public.stock_opname_sessions USING btree (outlet_id);


--
-- Name: idx_stock_opname_sessions_status; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_stock_opname_sessions_status ON public.stock_opname_sessions USING btree (status);


--
-- Name: idx_stock_opname_sessions_store_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_stock_opname_sessions_store_id ON public.stock_opname_sessions USING btree (store_id);


--
-- Name: idx_stores_deleted_at; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_stores_deleted_at ON public.stores USING btree (deleted_at);


--
-- Name: idx_suppliers_deleted_at; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_suppliers_deleted_at ON public.suppliers USING btree (deleted_at);


--
-- Name: idx_suppliers_store_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_suppliers_store_id ON public.suppliers USING btree (store_id);


--
-- Name: idx_transaction_items_deleted_at; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_transaction_items_deleted_at ON public.transaction_items USING btree (deleted_at);


--
-- Name: idx_transaction_items_trx_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_transaction_items_trx_id ON public.transaction_items USING btree (trx_id);


--
-- Name: idx_transaction_items_variant_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_transaction_items_variant_id ON public.transaction_items USING btree (variant_id);


--
-- Name: idx_transactions_cashier_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_transactions_cashier_id ON public.transactions USING btree (cashier_id);


--
-- Name: idx_transactions_customer_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_transactions_customer_id ON public.transactions USING btree (customer_id);


--
-- Name: idx_transactions_deleted_at; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_transactions_deleted_at ON public.transactions USING btree (deleted_at);


--
-- Name: idx_transactions_outlet_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_transactions_outlet_id ON public.transactions USING btree (outlet_id);


--
-- Name: idx_transactions_promo_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_transactions_promo_id ON public.transactions USING btree (promo_id);


--
-- Name: idx_transactions_store_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_transactions_store_id ON public.transactions USING btree (store_id);


--
-- Name: idx_transactions_voucher_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_transactions_voucher_id ON public.transactions USING btree (voucher_id);


--
-- Name: idx_users_deleted_at; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_users_deleted_at ON public.users USING btree (deleted_at);


--
-- Name: idx_users_last_seen_at; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_users_last_seen_at ON public.users USING btree (last_seen_at);


--
-- Name: idx_users_outlet_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_users_outlet_id ON public.users USING btree (outlet_id);


--
-- Name: idx_users_store_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE INDEX idx_users_store_id ON public.users USING btree (store_id);


--
-- Name: uq_category_store_name; Type: INDEX; Schema: public; Owner: openpos
--

CREATE UNIQUE INDEX uq_category_store_name ON public.categories USING btree (store_id, name);


--
-- Name: uq_customers_store_phone; Type: INDEX; Schema: public; Owner: openpos
--

CREATE UNIQUE INDEX uq_customers_store_phone ON public.customers USING btree (store_id, phone) WHERE ((phone <> ''::text) AND (deleted_at IS NULL));


--
-- Name: uq_notification_daily_key; Type: INDEX; Schema: public; Owner: openpos
--

CREATE UNIQUE INDEX uq_notification_daily_key ON public.notifications USING btree (store_id, dedup_key);


--
-- Name: uq_price_tier_product_qty; Type: INDEX; Schema: public; Owner: openpos
--

CREATE UNIQUE INDEX uq_price_tier_product_qty ON public.price_tiers USING btree (store_id, product_id, min_qty) WHERE ((variant_id IS NULL) AND (deleted_at IS NULL));


--
-- Name: uq_price_tier_variant_qty; Type: INDEX; Schema: public; Owner: openpos
--

CREATE UNIQUE INDEX uq_price_tier_variant_qty ON public.price_tiers USING btree (store_id, product_id, variant_id, min_qty) WHERE ((variant_id IS NOT NULL) AND (deleted_at IS NULL));


--
-- Name: uq_product_stock_product_outlet; Type: INDEX; Schema: public; Owner: openpos
--

CREATE UNIQUE INDEX uq_product_stock_product_outlet ON public.product_stocks USING btree (product_id, outlet_id) WHERE (variant_id IS NULL);


--
-- Name: uq_product_stock_variant_outlet; Type: INDEX; Schema: public; Owner: openpos
--

CREATE UNIQUE INDEX uq_product_stock_variant_outlet ON public.product_stocks USING btree (product_id, variant_id, outlet_id) WHERE (variant_id IS NOT NULL);


--
-- Name: uq_product_store_sku; Type: INDEX; Schema: public; Owner: openpos
--

CREATE UNIQUE INDEX uq_product_store_sku ON public.products USING btree (store_id, sku);


--
-- Name: uq_promo_store_voucher; Type: INDEX; Schema: public; Owner: openpos
--

CREATE UNIQUE INDEX uq_promo_store_voucher ON public.promos USING btree (store_id, voucher_code);


--
-- Name: uq_purchase_order_store_number; Type: INDEX; Schema: public; Owner: openpos
--

CREATE UNIQUE INDEX uq_purchase_order_store_number ON public.purchase_orders USING btree (store_id, number);


--
-- Name: uq_stock_opname_counting_store; Type: INDEX; Schema: public; Owner: openpos
--

CREATE UNIQUE INDEX uq_stock_opname_counting_store ON public.stock_opname_sessions USING btree (store_id) WHERE ((status = 'counting'::text) AND (deleted_at IS NULL));


--
-- Name: uq_transactions_store_client_tx_id; Type: INDEX; Schema: public; Owner: openpos
--

CREATE UNIQUE INDEX uq_transactions_store_client_tx_id ON public.transactions USING btree (store_id, client_tx_id) WHERE (client_tx_id <> ''::text);


--
-- Name: uq_variant_store_sku; Type: INDEX; Schema: public; Owner: openpos
--

CREATE UNIQUE INDEX uq_variant_store_sku ON public.product_variants USING btree (store_id, sku);


--
--



-- +goose Down
DROP TABLE IF EXISTS public.users CASCADE;
DROP TABLE IF EXISTS public.transactions CASCADE;
DROP TABLE IF EXISTS public.transaction_items CASCADE;
DROP TABLE IF EXISTS public.suppliers CASCADE;
DROP TABLE IF EXISTS public.stores CASCADE;
DROP TABLE IF EXISTS public.stock_opname_sessions CASCADE;
DROP TABLE IF EXISTS public.stock_opname_items CASCADE;
DROP TABLE IF EXISTS public.stock_movements CASCADE;
DROP TABLE IF EXISTS public.refunds CASCADE;
DROP TABLE IF EXISTS public.refresh_tokens CASCADE;
DROP TABLE IF EXISTS public.receivables CASCADE;
DROP TABLE IF EXISTS public.receivable_payments CASCADE;
DROP TABLE IF EXISTS public.purchase_orders CASCADE;
DROP TABLE IF EXISTS public.purchase_order_items CASCADE;
DROP TABLE IF EXISTS public.promos CASCADE;
DROP TABLE IF EXISTS public.products CASCADE;
DROP TABLE IF EXISTS public.product_variants CASCADE;
DROP TABLE IF EXISTS public.product_stocks CASCADE;
DROP TABLE IF EXISTS public.price_tiers CASCADE;
DROP TABLE IF EXISTS public.payables CASCADE;
DROP TABLE IF EXISTS public.payable_payments CASCADE;
DROP TABLE IF EXISTS public.outlets CASCADE;
DROP TABLE IF EXISTS public.notifications CASCADE;
DROP TABLE IF EXISTS public.expenses CASCADE;
DROP TABLE IF EXISTS public.email_otps CASCADE;
DROP TABLE IF EXISTS public.customers CASCADE;
DROP TABLE IF EXISTS public.categories CASCADE;
DROP TABLE IF EXISTS public.cashiers CASCADE;
DROP TABLE IF EXISTS public.cashier_shifts CASCADE;
DROP TABLE IF EXISTS public.activity_logs CASCADE;
