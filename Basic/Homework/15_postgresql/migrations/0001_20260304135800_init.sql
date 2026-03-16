-- +goose Up
CREATE table IF NOT EXISTS tasks (
    id              serial primary key,
    task   			jsonb UNIQUE not null,
    created_at      timestamptz not null default now(),
    updated_at      timestamptz
);

CREATE table IF NOT EXISTS notes (
    id              serial primary key,
    note   			jsonb UNIQUE not null,
    created_at      timestamptz not null default now(),
    updated_at      timestamptz
);

CREATE table IF NOT EXISTS remindables_log (
    id              serial primary key,
    description   	text not null,
    created_at      timestamptz not null default now()
);


-- +goose Down
drop table tasks;
drop table notes;
drop table remindables_log;
