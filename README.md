# MySQL Synchronization

## Usage
Data synchronization from one MySQL table to another MySQL table.

## Specification
Currently, the program only works on MySQL tables with 3 columns.

## How to use
Create an empty table with 3 columns in one MySQL database. (Can refer to `db.sql`).

Create another empty table with 3 columns in another MySQL database. (Can refer to `db.sql`).

Change the code in line 123 of `./cmd/go-mysqlbinlog/main.go` to the IP address, port number, and database name of the second MySQL database you want to use.

Run `SHOW MASTER STATUS` in the first MySQL and get the binlog file name and position. Assume they are binlog.000033 and 28124, respectively.

Then run:
```
go run ./cmd/go-mysqlbinlog/main.go --password=xxxx --port=3306 --file=binlog.000033 --pos=28124
```
where you need to replace the password, port, file name, and pos correctly.

Now the program is waiting in blocking state and you can insert/update/delete in the first MySQL table.

## Output
The modifications in the first MySQL table will be synchronized to the second MySQL table.

The description of all the modifications will also be printed out in the terminal.


## Reference 
Go-MySQL-Driver:
```
https://github.com/go-mysql-org/go-mysql
```
go-mysql:
```
https://github.com/go-sql-driver/mysql
```
