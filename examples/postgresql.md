# Postgres Examples

db=db-postgres|name=createtable
```sql
CREATE TABLE todo (
    id serial not null,
    name varchar,
    is_done boolean
);
```

result:
---
|  |
||
---

db=db-postgres|name=inserttodo
```sql
INSERT INTO todo (name, is_done) VALUES 
('first todo', true),
('second todo', false);
```

result:
---
|  |
||
---

db=db-postgres|name=todoschema
```sql
SELECT 
    column_name, 
    data_type, 
    character_maximum_length AS max_length, 
    is_nullable
FROM 
    information_schema.columns
WHERE 
    table_schema = 'public' -- Replace with your schema name if different
    AND table_name = 'todo' -- Replace with your table name
ORDER BY 
    ordinal_position;
```

result:
---
| column_name | data_type         | max_length | is_nullable |
|-------------|-------------------|------------|-------------|
| id          | integer           |            | NO          |
| name        | character varying |            | YES         |
| is_done     | boolean           |            | YES         |
---

db=db-postgres|name=getalltodo 
```sql
SELECT * from todo; 
```

result:
---
| id | name        | is_done |
|----|-------------|---------|
| 1  | first todo  | true    |
| 2  | second todo | false   |
---

