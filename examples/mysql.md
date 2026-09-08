# MySQL Examples

db=db-mysql|name=createtable
```sql
CREATE TABLE todo (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255),
    is_done BOOLEAN
);
```

result:
---
|  |
||
---

db=db-mysql|name=inserttodo
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

db=db-mysql|name=todoschema
```sql
SELECT 
    column_name, 
    data_type, 
    character_maximum_length AS max_length, 
    is_nullable
FROM 
    information_schema.columns
WHERE 
    table_schema = DATABASE() -- Replace with your database name if different
    AND table_name = 'todo' -- Replace with your table name
ORDER BY 
    ordinal_position;
```

result:
---
| COLUMN_NAME | DATA_TYPE | max_length | IS_NULLABLE |
|-------------|-----------|------------|-------------|
| id          | int       |            | NO          |
| name        | varchar   | 255        | YES         |
| is_done     | tinyint   |            | YES         |
---

db=db-mysql|name=getalltodo
```sql
SELECT * FROM todo;
```

result:
---
| id | name        | is_done |
|----|-------------|---------|
| 1  | first todo  | 1       |
| 2  | second todo | 0       |
---

