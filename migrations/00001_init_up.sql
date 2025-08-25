BEGIN;

SET statement_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = ON;
SET check_function_bodies = FALSE;
SET client_min_messages = WARNING;
SET search_path = public, extensions;
SET default_tablespace = '';
SET default_with_oids = FALSE;

-- EXTENSIONS --

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- TABLES -- 

CREATE TABLE public.currency
(
    id SERIAL PRIMARY KEY,
    name TEXT,
    symbol TEXT
);

CREATE TABLE public.category
(
    id SERIAL PRIMARY KEY,
    name TEXT
);

CREATE TABLE public.product
(
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            TEXT NOT NULL,
    description     TEXT NOT NULL,
    image_id        UUID,
    price           BIGINT,
    currency_id     INT REFERENCES public.currency (id),
    rating          INT,
    category_id     INT REFERENCES public.category (id),
    specification   JSONB,
    created_at      TIMESTAMPTZ,
    updated_at      TIMESTAMPTZ,
    CONSTRAINT valid_rating CHECK (rating<= 5)
);

-- DATA --

INSERT INTO public.currency (name, symbol)
VALUES ('рубль','₽');
INSERT INTO public.currency (name, symbol)
VALUES ('dollar', '$');

INSERT INTO public.category (name)
VALUES ('купоны');
INSERT INTO public.category (name)
VALUES ('цифровые билеты');

INSERT INTO public.product (
    name,
    description,
    price,
    currency_id,
    rating,
    category_id,
    specification,
    created_at
)
VALUES (
    'a',                               -- строка
    'b',                               -- строка
    100,                               -- число без кавычек
    1,                                 -- внешний ключ, число
    3,                                 -- рейтинг, число
    1,                                 -- категория, число
    '{"A": "ok"}',                     -- JSON-строка
    '2006-01-02 15:04:05+00'           -- дата/время
);


COMMIT;