CREATE SCHEMA epay;

CREATE EXTENSION pgcrypto SCHEMA epay;

ALTER SCHEMA epay OWNER TO postgres;

CREATE FUNCTION epay.generate_id(size INT) RETURNS TEXT AS $$
DECLARE
  characters TEXT := 'abcdefghijklmnopqrstuvwxyz0123456789';
  bytes BYTEA := epay.gen_random_bytes(size);
  l INT := length(characters);
  i INT := 0;
  output TEXT := '';
BEGIN
  WHILE i < size LOOP
    output := output || substr(characters, get_byte(bytes, i) % l + 1, 1);
    i := i + 1;
  END LOOP;
  RETURN output;
END;
$$ LANGUAGE plpgsql VOLATILE;

CREATE TABLE epay.transactions (
    transaction_id integer NOT NULL,
    from_id text,
    to_id text NOT NULL,
    date timestamp with time zone NOT NULL,
    value integer NOT NULL,
    CONSTRAINT transactions_from_and_to_check CHECK ((from_id <> to_id)),
    CONSTRAINT transactions_value_check CHECK ((value > 0))
);

ALTER TABLE epay.transactions OWNER TO postgres;

CREATE SEQUENCE epay.transactions_transaction_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER TABLE epay.transactions_transaction_id_seq OWNER TO postgres;

ALTER SEQUENCE epay.transactions_transaction_id_seq OWNED BY epay.transactions.transaction_id;

CREATE TABLE epay.wallets (
    wallet_id TEXT PRIMARY KEY DEFAULT epay.generate_id(32),
    firstname text NOT NULL,
    lastname text NOT NULL,
    balance integer NOT NULL,
    CONSTRAINT wallets_balance_check CHECK ((balance >= 0)),
    CONSTRAINT wallets_id_check CHECK ((wallet_id ~ '^[0-9a-z]{32}$'::text))
);


ALTER TABLE epay.wallets OWNER TO postgres;

ALTER TABLE ONLY epay.transactions ALTER COLUMN transaction_id SET DEFAULT nextval('epay.transactions_transaction_id_seq'::regclass);

SELECT pg_catalog.setval('epay.transactions_transaction_id_seq', 1, false);

ALTER TABLE ONLY epay.transactions
    ADD CONSTRAINT transactions_pkey PRIMARY KEY (transaction_id);

ALTER TABLE ONLY epay.transactions
    ADD CONSTRAINT transactions_from_id_fkey FOREIGN KEY (from_id) REFERENCES epay.wallets(wallet_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE ONLY epay.transactions
    ADD CONSTRAINT transactions_to_id_fkey FOREIGN KEY (to_id) REFERENCES epay.wallets(wallet_id) ON UPDATE CASCADE ON DELETE RESTRICT;

