create user manager_service with password 'test@12345';
revoke all on database infra from manager_service;
grant connect on database infra to manager_service;

create schema if not exists portfolio;

grant usage on schema public, portfolio to manager_service;

revoke insert, update, delete on table 
    portfolio.user, 
    portfolio.message, 
    portfolio.technology, 
    portfolio.technology_project, 
    portfolio.technology_experience,
    portfolio.project,
    portfolio.project_position,
    portfolio.position,
    portfolio.experience 
from manager_service;

grant select on table 
    portfolio.user, 
    portfolio.message, 
    portfolio.technology, 
    portfolio.technology_project, 
    portfolio.technology_experience,
    portfolio.project,
    portfolio.project_position,
    portfolio.position,
    portfolio.experience
to manager_service;

grant execute on function 
    portfolio.edit_user(int, text, varchar), 
    portfolio.delete_messages(uuid, int), 
    portfolio.edit_technology(int, int, varchar, varchar, varchar, smallint),
    portfolio.edit_project(int, int, varchar, text, varchar, varchar, varchar, boolean, int[]),
    portfolio.edit_experience(int, int, varchar, portfolio.position_create[], int[]),
    portfolio.delete_experience(int, int),
    portfolio.delete_project(int, int),
    portfolio.delete_technology(int, int)
to manager_service;