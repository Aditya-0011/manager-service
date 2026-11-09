/*
    OutCodes:
        -1: Success

        0: Record exists

        1: Record not found

        2: Db Error
*/

create table if not exists portfolio.user (
    id int primary key,
    about text not null,
    coverImage varchar(250) not null,
    chatModel varchar(100) not null default '',
    updatedAt timestamptz default now()
);

create or replace function portfolio.edit_user(
    in p_id int, 
    in p_about text, 
    in p_coverImage varchar, 
    in p_chatModel varchar, 
    out outcode smallint
)
as $$
begin
    insert into portfolio.user (id, about, coverImage, chatModel) 
    values (p_id, p_about, p_coverImage, p_chatModel)
    on conflict (id) do update 
    set about = excluded.about, 
        coverImage = excluded.coverImage, 
        chatModel = excluded.chatModel, 
        updatedAt = now();
        
    outcode := -1;
exception 
    when others then 
        raise warning 'DB Error: %', sqlerrm;
        outcode := 2;
end;
$$ language plpgsql security definer set search_path = '';

revoke insert, update, delete on table 
    portfolio.user
from manager_service;

grant select on table 
    portfolio.user
to manager_service;

revoke execute on function 
    portfolio.edit_user(int, text, varchar, varchar)
from public;

grant execute on function 
    portfolio.edit_user(int, text, varchar, varchar)
to manager_service;