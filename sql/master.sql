create user manager_service with password 'test@12345';
revoke all on database infra from manager_service;
grant connect on database infra to manager_service;

create schema if not exists portfolio;

grant usage on schema public, portfolio to manager_service;