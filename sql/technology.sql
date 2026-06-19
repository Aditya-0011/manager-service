/*
    OutCodes:
        -1: Success

        0: Record exists

        1: Record not found

        2: Db Error
*/

create table if not exists portfolio.technology (
    id int generated always as identity primary key,
    userId int not null,
    name varchar(255) not null,
    imageUrl varchar(255) not null,
    fallbackImageUrl varchar(255) not null,
    category smallint not null,
    updatedAt timestamptz default now()
);
create index idx_technology_userId on portfolio.technology(userId);

create table if not exists portfolio.technology_project(
    technologyId int not null,
    projectId int not null,
    primary key(projectId, technologyId)
);
create index idx_technology_project_technologyId on portfolio.technology_project(technologyId);

create table if not exists portfolio.technology_experience(
    technologyId int not null,
    experienceId int not null,
    primary key(experienceId, technologyId)
);
create index idx_technology_experience_technologyId on portfolio.technology_experience(technologyId);

create or replace function portfolio.edit_technology(
    in p_id int, 
    in p_userId int, 
    in p_name varchar, 
    in p_imageUrl varchar, 
    in p_fallbackImageUrl varchar, 
    in p_category int, 
    out outcode smallint
)
as $$
begin 
    if p_id > 0 then
        update portfolio.technology set name = p_name, imageUrl = p_imageUrl, fallbackImageUrl = p_fallbackImageUrl, category = p_category::smallint, updatedAt = now()
        where id = p_id and userId = p_userId;
        
        if not found then
            outcode := 1;
            return;
        end if;

    else
        insert into portfolio.technology (userId, name, imageUrl, fallbackImageUrl, category) values (p_userId, p_name, p_imageUrl, p_fallbackImageUrl, p_category::smallint);
    end if;
    outcode := -1;
exception 
    when others then 
        raise warning 'DB Error: %', sqlerrm;
        outcode := 2;
end;
$$ language plpgsql security definer;

create or replace function portfolio.delete_technology(
    in p_id int, 
    in p_userId int, 
    out outcode smallint
) as $$
begin 
    perform 1 from portfolio.technology where id = p_id and userId = p_userId for update;
    if not found then
        outcode := 1;
        return;
    end if;

    delete from portfolio.technology_project where technologyId = p_id;
    delete from portfolio.technology_experience where technologyId = p_id;
    delete from portfolio.technology where id = p_id;
    outcode := -1;
exception 
    when others then 
        raise warning 'DB Error: %', sqlerrm;
        outcode := 2;
end;
$$ language plpgsql security definer;

revoke insert, update, delete on table
    portfolio.technology, 
    portfolio.technology_project, 
    portfolio.technology_experience
from manager_service;

grant select on table
    portfolio.technology, 
    portfolio.technology_project, 
    portfolio.technology_experience
to manager_service;

grant execute on function 
    portfolio.edit_technology(int, int, varchar, varchar, varchar, int),
    portfolio.delete_technology(int, int)
to manager_service;
