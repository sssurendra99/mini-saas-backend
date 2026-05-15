# Mini Saas API

- This is created to polish the existing knowledge on golang.


## Mind map

- revising the go project structure
- creating the crud setup
- connecting relational databases (postgresql)
- setting up authentication
- adding necessary middlewares
- revising the concurrency
- adding caching
- making production 


## Several technical decisions I used.

- used the https://github.com/joho/godotenv for the adding environmental variables.
```bash
https://github.com/joho/godotenv
```
- Had a confusion where which one to use as the posgtresql driver pgx or go lang posgtres driver pure one.
- Added the uuid package go get github.com/google/uuid for using the uuids. (correction: here I'm using pgx built in uuid | Reason: you need the adapter installation for that.)
```bash
github.com/google/uuid
```
- When creating the repository, I thought rather than creating three repositories for taks, project and user it's easier to create a generic repository much easier. (Learning curve is somewhat big so skipping this for now.)



## Things understood

- constructor pattern in golang - why struct pointer over struct value and struct type. (Very important)
- Understanding dependency injection is critical (Refer how did I injected dp connection pool into handlers)
- Dependency Injection for the DB: created a connection pool -> repository gets a connection -> with the repository service deals with business logic ->  Handlers handle the requests/responses.
- for formatting the code I used gofmt -s -w .
```bash
gofmt -s -w .
```

## Things to look Later.

2026/05/26 - Generics 


## Queries

```bash
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,

    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID NOT NULL,

    name TEXT NOT NULL,
    description TEXT,

    created_at TIMESTAMP DEFAULT NOW(),

    CONSTRAINT fk_projects_user
        FOREIGN KEY(user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    project_id UUID NOT NULL,
    user_id UUID NOT NULL,

    title TEXT NOT NULL,
    description TEXT,

    completed BOOLEAN DEFAULT FALSE,

    created_at TIMESTAMP DEFAULT NOW(),

    CONSTRAINT fk_tasks_project
        FOREIGN KEY(project_id)
        REFERENCES projects(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_tasks_user
        FOREIGN KEY(user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);
```

```bash
INSERT INTO users (name, email, password_hash)
VALUES
(
    'John Doe',
    'john@example.com',
    'hashed_password_1'
),
(
    'Alice Smith',
    'alice@example.com',
    'hashed_password_2'
),
(
    'Michael Brown',
    'michael@example.com',
    'hashed_password_3'
);

INSERT INTO projects (user_id, name, description)
VALUES
(
    (SELECT id FROM users WHERE email = 'john@example.com'),
    'Go SaaS Backend',
    'Learning backend engineering with Go'
),
(
    (SELECT id FROM users WHERE email = 'alice@example.com'),
    'Portfolio Website',
    'Building a personal portfolio'
),
(
    (SELECT id FROM users WHERE email = 'michael@example.com'),
    'Task Manager',
    'Simple project management system'
);

INSERT INTO tasks (
    project_id,
    user_id,
    title,
    description,
    completed
)
VALUES
(
    (SELECT id FROM projects WHERE name = 'Go SaaS Backend'),
    (SELECT id FROM users WHERE email = 'john@example.com'),
    'Setup PostgreSQL',
    'Install PostgreSQL and configure pgx',
    FALSE
),

(
    (SELECT id FROM projects WHERE name = 'Go SaaS Backend'),
    (SELECT id FROM users WHERE email = 'john@example.com'),
    'Implement JWT Auth',
    'Create authentication system',
    FALSE
),

(
    (SELECT id FROM projects WHERE name = 'Portfolio Website'),
    (SELECT id FROM users WHERE email = 'alice@example.com'),
    'Design Landing Page',
    'Create homepage UI',
    TRUE
),

(
    (SELECT id FROM projects WHERE name = 'Task Manager'),
    (SELECT id FROM users WHERE email = 'michael@example.com'),
    'Create API Routes',
    'Implement CRUD endpoints',
    FALSE
);
```