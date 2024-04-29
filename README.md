# Go CRUD API with PostgreSQL

A simple API in Go that performs CRUD (Create, Read, Update, Delete) operations on a PostgreSQL database. I created this project to learn how to build a robust and secure API in Go.

## Users Table Structure

| Column     | Type                        | Nullable | Default | Description                                         |
| ---------- | --------------------------- | -------- | ------- | --------------------------------------------------- |
| id         | integer                     | not null |         | Unique identifier for the user                      |
| name       | text                        | not null |         | Name of the user                                    |
| email      | text                        | not null |         | Email address of the user (unique constraint)       |
| created_at | timestamp without time zone | not null | now()   | Timestamp indicating when the user was created      |
| updated_at | timestamp without time zone | not null | now()   | Timestamp indicating when the user was last updated |
