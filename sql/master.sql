do $$
begin
    if not exists (select 1 from pg_roles where rolname = 'manager_service') then
        create user manager_service with password 'test@12345';
    end if;
end
$$;
revoke all on database infra from manager_service;
grant connect on database infra to manager_service;

create schema if not exists portfolio;

grant usage on schema public, portfolio to manager_service;