/*
    OutCodes:
        -1: Success

        0: Record exists

        1: Record not found

        2: Db Error
*/

create table if not exists portfolio.message (
    id uuid primary key default uuidv7(),
    userId int not null,
    name text not null,
    email text not null,
    messages text[3],
    messages_count int not null default 1
);

create or replace function portfolio.delete_messages(
    in p_id uuid, 
    in p_userId int, 
    out outcode smallint
)
as $$
begin
    if exists (select 1 from portfolio.message where id = p_id and userId = p_userId for update) then
        delete from portfolio.message where id = p_id and userId = p_userId;
        outcode := -1;
    else
        outcode := 1;
    end if;
exception 
    when others then outcode := 2;
end;
$$ language plpgsql security definer;