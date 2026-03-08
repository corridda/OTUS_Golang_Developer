-- +goose Up
-- +goose StatementBegin
CREATE INDEX index_task ON tasks USING GIN (task);
CREATE UNIQUE INDEX index_task_name ON tasks (((task->>'name')::text));
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX index_note ON notes USING GIN (note);
CREATE UNIQUE INDEX index_note_name ON notes (((note->>'name')::text));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop index index_task;
-- +goose StatementEnd

-- +goose StatementBegin
drop index index_note;
-- +goose StatementEnd
