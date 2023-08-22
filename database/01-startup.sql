create table if not exists job (
    id integer,
    user_id text,
    data text,
    status text not null,
    created_at timestamp with time zone not null,
    published_at timestamp with time zone,
    started_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    PRIMARY KEY(id)
);

CREATE SEQUENCE public.job_id_seq START 1;