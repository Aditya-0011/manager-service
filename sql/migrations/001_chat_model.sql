alter table portfolio.user
    add column if not exists chatModel varchar(100) not null default '';

drop function if exists portfolio.edit_user(int, text, varchar);
