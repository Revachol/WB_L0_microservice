CREATE TABLE orders (
    order_uid TEXT PRIMARY KEY,
    track_number TEXT,
    entry TEXT,
    locale TEXT,
    internal_signature TEXT,
    customer_id TEXT,
    delivery_service TEXT,
    shardkey TEXT,
    sm_id INT,
    date_created TIMESTAMP,
    oof_shard TEXT
);

CREATE TABLE deliveries (
    id SERIAL PRIMARY KEY,
    order_uid TEXT REFERENCES orders(order_uid) ON DELETE CASCADE,
    name TEXT,
    phone TEXT,
    zip TEXT,
    city TEXT,
    address TEXT,
    region TEXT,
    email TEXT
);

CREATE TABLE payments (
    id SERIAL PRIMARY KEY,
    order_uid TEXT REFERENCES orders(order_uid) ON DELETE CASCADE,
    transaction TEXT,
    request_id TEXT,
    currency TEXT,
    provider TEXT,
    amount INT,
    payment_dt BIGINT,
    bank TEXT,
    delivery_cost INT,
    goods_total INT,
    custom_fee INT
);

CREATE TABLE items (
    id SERIAL PRIMARY KEY,
    order_uid TEXT REFERENCES orders(order_uid) ON DELETE CASCADE,
    chrt_id INT,
    track_number TEXT,
    price INT,
    rid TEXT,
    name TEXT,
    sale INT,
    size TEXT,
    total_price INT,
    nm_id INT,
    brand TEXT,
    status INT
);
