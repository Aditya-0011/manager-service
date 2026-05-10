/*
    OutCodes:
        -1: Success

        0: Record exists

        1: Record not found

        2: Db Error
*/

/*
create table if not exists portfolio.technology_project(
    technologyId int not null,
    projectId int not null,
    primary key(projectId, technologyId)
)
*/

create table if not exists portfolio.project(
    id int generated always as identity primary key,
    userId int not null,
    name varchar(255) not null,
    description text not null,
    imageUrl varchar(255) not null,
    projectUrl varchar(255),
    githubUrl varchar(255),
    featured boolean not null,
    updatedAt timestamptz default now()
);
create index idx_project_userId on portfolio.project(userId);

create table if not exists portfolio.project_position(
    projectId int not null,
    positionId int not null,
    primary key(positionId, projectId)
);
create index idx_project_position_projectId on portfolio.project_position(projectId);

create or replace function portfolio.edit_project(
    in p_id int, 
    in p_userId int, 
    in p_name varchar, 
    in p_description text, 
    in p_imageUrl varchar, 
    in p_projectUrl varchar, 
    in p_githubUrl varchar, 
    in p_featured boolean, 
    in p_technologies int[], 
    out outcode smallint
)
as $$
declare
    v_project_id int;
begin 
    if p_technologies is not null and array_length(p_technologies, 1) > 0 then
        if (select count(id) from portfolio.technology where id = any(p_technologies) and userId = p_userId) != array_length(p_technologies, 1) then
            outcode := 1;
            return;
        end if;
    end if;

    if p_id > 0 then
        if exists (select 1 from portfolio.project where id = p_id and userId = p_userId for update) then
            v_project_id := p_id;

            update portfolio.project set name = p_name, description = p_description, imageUrl = p_imageUrl, projectUrl = p_projectUrl, githubUrl = p_githubUrl, featured = p_featured, updatedAt = now()
            where id = v_project_id;

            delete from portfolio.technology_project 
            where projectId = v_project_id 
            and technologyId <> all(coalesce(p_technologies, '{}'::int[]));

            if p_technologies is not null then
                insert into portfolio.technology_project (projectId, technologyId) 
                select v_project_id, t_id from unnest(p_technologies) as t_id
                on conflict (projectId, technologyId) do nothing;
            end if;
        else
            outcode := 1;
            return;
        end if;
    else
        insert into portfolio.project (userId, name, description, imageUrl, projectUrl, githubUrl, featured) 
        values (p_userId, p_name, p_description, p_imageUrl, p_projectUrl, p_githubUrl, p_featured)
        returning id into v_project_id;

        if p_technologies is not null then
            insert into portfolio.technology_project (projectId, technologyId) 
            select v_project_id, t_id from unnest(p_technologies) as t_id;
        end if;
    end if;
    outcode := -1;
exception 
    when others then outcode := 2;
end;
$$ language plpgsql security definer;

create or replace function portfolio.delete_project(
    in p_id int, 
    in p_userId int, 
    out outcode smallint
) as $$
begin 
    if exists (select 1 from portfolio.project where id = p_id and userId = p_userId for update) then
        delete from portfolio.technology_project where projectId = p_id;
        delete from portfolio.project_position where projectId = p_id;
        delete from portfolio.project where id = p_id;
    else
        outcode := 1;
        return;
    end if;
    outcode := -1;
exception 
    when others then outcode := 2;
end;
$$ language plpgsql security definer;