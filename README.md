# PG App CLI

Generate a Go Application that can be released as Postgres Plugin.

### Install

Download the executable and copy it to a Folder that can the system found. For Example /usr/bin.

If you want to build it from Source simple checkout the Project and install Go 1.19 and run following commands:

```bash
cd {pgapp_cli folder}
go build -o /usr/bin/pgapp_cli main.go
```

### Usage

```bash
# switch to the Go Module Path in this Example {user_folder}/go/src/githib.com/nodejayes
cd /home/{user}/go/src/githib.com/nodejayes
# for the used Parameter Values look in the Table below
pgapp_cli create {go_module_path} {go_version} {postgres_version} {description}
cd pg_todo_app
# generate the Postgres Plugin Files in ./bin folder
go generate main.go
```

| Parameter | Description | Example                          |
|-----------|-------------|----------------------------------|
| go_module_path | a valid module path that is used for a go module | github.com/nodejayes/pg_todo_app |
| go_version | a valid Go runtime Version that was used | 1.19 |
| pg_version | a valid Postgres Version that was used | 14 |

### Requirements
1. Go in the Version you give the pgapp_cli
2. a valid cpp compiler for example install build-essential
3. Postgres Development Files postgresql-server-dev-14 (the version must match with the version you're giving the pgapp_cli command!)

### Use Plugin in Postgres

```bash
# Copy the Files into Postgres Folders
cp /home/{user}/go/src/githib.com/nodejayes/pg_todo_app/bin/pg_todo_app.so /usr/lib/postgresql/14/lib
cp /home/{user}/go/src/githib.com/nodejayes/pg_todo_app/bin/pg_todo_app.control /usr/share/postgresql/14/extension
cp /home/{user}/go/src/githib.com/nodejayes/pg_todo_app/bin/pg_todo_app--1.0.0.sql /usr/share/postgresql/14/extension
# install Extension in Database simple execute following SQL
create extension pg_todo_app;

# after that you have a schema pg_todo_app that has a dispatch function
```