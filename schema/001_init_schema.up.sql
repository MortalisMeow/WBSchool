CREATE TABLE IF NOT EXISTS orders (
                                      order_uid          VARCHAR(255) PRIMARY KEY,
                                      track_number       VARCHAR(255) NOT NULL,
                                      entry              VARCHAR(255) NOT NULL,
                                      locale             VARCHAR(10) NOT NULL,
                                      internal_signature VARCHAR(255),
                                      customer_id        VARCHAR(255) NOT NULL,
                                      delivery_service   VARCHAR(255) NOT NULL,
                                      shardkey           VARCHAR(10) NOT NULL,
                                      sm_id              INTEGER NOT NULL,
                                      date_created       TIMESTAMP WITH TIME ZONE NOT NULL,
                                      oof_shard          VARCHAR(10) NOT NULL
);


CREATE TABLE IF NOT EXISTS payments (
                                        payment_id     SERIAL PRIMARY KEY,
                                        transaction   VARCHAR(255) NOT NULL,
                                        request_id    VARCHAR(255),
                                        currency      VARCHAR(10) NOT NULL,
                                        provider      VARCHAR(255) NOT NULL,
                                        amount        INTEGER NOT NULL,
                                        payment_dt    BIGINT NOT NULL,
                                        bank          VARCHAR(255) NOT NULL,
                                        delivery_cost BIGINT NOT NULL,
                                        goods_total   INTEGER NOT NULL,
                                        custom_fee    INTEGER NOT NULL DEFAULT 0,
                                        order_uid     VARCHAR(255) NOT NULL REFERENCES orders(order_uid) ON DELETE CASCADE
);


CREATE TABLE IF NOT EXISTS deliveries (
                                          delivery_id SERIAL PRIMARY KEY,
                                          order_uid   VARCHAR(255) NOT NULL REFERENCES orders(order_uid) ON DELETE CASCADE,
                                          name        VARCHAR(255) NOT NULL,
                                          phone       VARCHAR(50) NOT NULL,
                                          zip         VARCHAR(50) NOT NULL,
                                          city        VARCHAR(255) NOT NULL,
                                          address     TEXT NOT NULL,
                                          region      VARCHAR(255) NOT NULL,
                                          email       VARCHAR(255) NOT NULL
);


CREATE TABLE IF NOT EXISTS items (
                                     chrt_id      BIGINT NOT NULL,
                                     order_uid    VARCHAR(255) NOT NULL REFERENCES orders(order_uid) ON DELETE CASCADE,
                                     price        BIGINT NOT NULL,
                                     rid          VARCHAR(255) NOT NULL,
                                     name         VARCHAR(255) NOT NULL,
                                     sale         INTEGER NOT NULL,
                                     size         VARCHAR(10) NOT NULL,
                                     total_price  BIGINT NOT NULL,
                                     nm_id        BIGINT NOT NULL,
                                     brand        VARCHAR(255) NOT NULL,
                                     status       INTEGER NOT NULL,
                                     PRIMARY KEY (chrt_id, order_uid)
);
