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
    name varchar(150) not null,
    email varchar(150) not null,
    messages varchar(500)[3],
    messages_count int
);

create index idx_message_userId on portfolio.message(userId);
create index idx_message_email on portfolio.message(email);

create or replace function portfolio.add_message(
    in p_userId int,
    in p_name varchar,
    in p_email varchar,
    in p_message varchar,
    out outcode smallint
)
as $$
declare
    v_record int;
begin
    select messages_count into v_record from portfolio.message where userId = p_userId and email = p_email for update;
    
    if found then
        if (v_record + 1) <= 3 then
            update portfolio.message set messages = array_append(messages, p_message), messages_count = messages_count + 1 where userId = p_userId and email = p_email;
            outcode := -1;
        else
            outcode := 0;
        end if;
    else
        insert into portfolio.message (userId, name, email, messages, messages_count) values (p_userId, p_name, p_email, array[p_message], 1);
        outcode := -1;
    end if;
exception 
    when others then 
        raise warning 'DB Error: %', sqlerrm;
        outcode := 2;
end;
$$ language plpgsql security definer;

create or replace function portfolio.delete_messages(
    in p_id uuid, 
    in p_userId int, 
    out outcode smallint
)
as $$
begin
    delete from portfolio.message where id = p_id and userId = p_userId;
    if found then
        outcode := -1;
    else
        outcode := 1;
    end if;
exception 
    when others then 
        raise warning 'DB Error: %', sqlerrm;
        outcode := 2;
end;
$$ language plpgsql security definer;

revoke insert, update, delete on table 
    portfolio.message
from manager_service;

grant select on table 
    portfolio.message
to manager_service;

grant execute on function 
    portfolio.add_message(int, varchar, varchar, varchar),
    portfolio.delete_messages(uuid, int)
to manager_service;