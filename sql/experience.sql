/*
    OutCodes:
        -1: Success

        0: Record exists

        1: Record not found

        2: Db Error
*/

/*
create table if not exists portfolio.technology_experience(
    technologyId int not null,
    experienceId int not null,
    primary key(experienceId, technologyId)
)
*/

/*
create table if not exists portfolio.project_position(
    projectId int not null,
    positionId int not null,
    primary key(positionId, projectId)
);
*/

create table if not exists portfolio.position(
    id int generated always as identity primary key,
    experienceId int not null,
    role varchar(150) not null,
    start varchar(100) not null,
    end varchar(100),
    workDone text not null
);
create index idx_position_experienceId on portfolio.position(experienceId);

create table if not exists portfolio.experience(
    id int generated always as identity primary key,
    userId int not null,
    company varchar(200) not null,
    start varchar(100) not null,
    end varchar(100),
    tenure varchar(100),
    updatedAt timestamptz default now()
);
create index idx_experience_userId on portfolio.experience(userId);

create type portfolio.position_create as(
    id int,
    role varchar(150),
    start varchar(100),
    end varchar(100),
    workDone text,
    projects int[]
);

create or replace function portfolio.edit_experience(
    in p_id int,
    in p_userId int,
    in p_company varchar,
    in p_positions portfolio.position_create[],
    in p_technologies int[], 
    out outcode smallint
) as $$
declare
    v_experience_id int;
    v_position_id int;
    pos portfolio.position_create;
    
    v_earliest_start date;
    v_latest_end date;
    v_final_end varchar := '';
    v_tenure varchar := '';
    
    v_parsed_start date;
    v_parsed_end date;
    v_display_end varchar;
    
    v_interval interval;
    v_years int;
    v_months int;
begin 
    if p_technologies is not null and array_length(p_technologies, 1) > 0 then
        if (select count(id) from portfolio.technology where id = any(p_technologies) and userId = p_userId) != array_length(p_technologies, 1) then
            outcode := 1;
            return;
        end if;
    end if;

    foreach pos in array p_positions loop
        if pos.id > 0 then
            if p_id <= 0 then
                outcode := 1;
                return;
            end if;
            if not exists (select 1 from portfolio.position where id = pos.id and experienceId = p_id) then
                outcode := 1;
                return;
            end if;
        end if;
        
        if pos.projects is not null and array_length(pos.projects, 1) > 0 then
            if (select count(id) from portfolio.projects where id = any(pos.projects) and userId = p_userId) != array_length(pos.projects, 1) then
                outcode := 1;
                return;
            end if;
        end if;

        v_parsed_start := pos.start::date;
        
        if pos.end is null or pos.end = '' or pos.end ilike 'present' then
            v_parsed_end := current_date;
            v_display_end := 'Present';
            v_final_end := 'Present';
        else
            v_parsed_end := pos.end::date;
            v_display_end := pos.end;
        end if;
        
        if v_earliest_start is null or v_parsed_start < v_earliest_start then
            v_earliest_start := v_parsed_start;
        end if;
        
        if v_latest_end is null or v_parsed_end > v_latest_end then
            v_latest_end := v_parsed_end;
        end if;
    end loop;
    
    if v_final_end != 'Present' then
        v_final_end := v_latest_end::varchar;
        v_interval := age(v_latest_end, v_earliest_start);
        v_years := extract(year from v_interval);
        v_months := extract(month from v_interval);
        
        if v_years > 0 then
            v_tenure := v_years || ' yrs ';
        end if;
        if v_months > 0 then
            v_tenure := v_tenure || v_months || ' mos';
        end if;
        v_tenure := trim(v_tenure);
    end if;

    if p_id > 0 then
        if exists (select 1 from portfolio.experience where id = p_id and userId = p_userId for update) then
            v_experience_id := p_id;
            update portfolio.experience 
            set company = p_company, 
                start = v_earliest_start::varchar, 
                "end" = v_final_end, 
                tenure = v_tenure,
                updatedAt = now()
            where id = v_experience_id;
        else
            outcode := 1;
            return;
        end if;
    else
        insert into portfolio.experience (userId, company, start, "end", tenure) 
        values (p_userId, p_company, v_earliest_start::varchar, v_final_end, v_tenure) 
        returning id into v_experience_id;
    end if;

    if p_id > 0 then
        delete from portfolio.project_position 
        where positionId in (
            select id from portfolio.position 
            where experienceId = v_experience_id 
            and id <> all( array(select p.id from unnest(p_positions) as p where p.id > 0) )
        );
        
        delete from portfolio.position 
        where experienceId = v_experience_id 
        and id <> all( array(select p.id from unnest(p_positions) as p where p.id > 0) );
    end if;

    foreach pos in array p_positions loop
        if pos.end is null or pos.end = '' or pos.end ilike 'present' then
            v_display_end := 'Present';
        else
            v_display_end := pos.end;
        end if;
        
        if pos.id = -1 then
            insert into portfolio.position (experienceId, role, start, "end", workDone)
            values (v_experience_id, pos.role, pos.start, v_display_end, pos.workDone)
            returning id into v_position_id;
        else
            v_position_id := pos.id;
            update portfolio.position 
            set role = pos.role, start = pos.start, "end" = v_display_end, workDone = pos.workDone
            where id = v_position_id and experienceId = v_experience_id;
            
            delete from portfolio.project_position where positionId = v_position_id;
        end if;
        
        if pos.projects is not null and array_length(pos.projects, 1) > 0 then
            insert into portfolio.project_position (positionId, projectId)
            select v_position_id, t_id from unnest(pos.projects) as t_id;
        end if;
    end loop;
    
    delete from portfolio.technology_experience where experienceId = v_experience_id;
    
    if p_technologies is not null and array_length(p_technologies, 1) > 0 then
        insert into portfolio.technology_experience (experienceId, technologyId)
        select v_experience_id, t_id from unnest(p_technologies) as t_id;
    end if;

    outcode := -1;
exception 
    when others then outcode := 2;
end;
$$ language plpgsql security definer;

create or replace function portfolio.delete_experience(
    in p_id int,
    in p_userId int,
    out outcode smallint
) as $$
begin 
    if exists (select 1 from portfolio.experience where id = p_id and userId = p_userId for update) then
        delete from portfolio.project_position where positionId in (select id from portfolio.position where experienceId = p_id);
        delete from portfolio.technology_experience where experienceId = p_id;
        delete from portfolio.position where experienceId = p_id;
        delete from portfolio.experience where id = p_id;
    else
        outcode := 1;
        return;
    end if;
    outcode := -1;
exception 
    when others then outcode := 2;
end;
$$ language plpgsql security definer;
