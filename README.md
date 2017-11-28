# mytx

This is a transaction tester for MySQL.

## Description

[mytx](https://github.com/tom--bo/mytx) is motivated to simplify the test process of checking lock status in MySQL.

## Usage

$ mytx [options] PLAN-FILE.txt

- PLAN-FILE.txt
  - Specify a transaction number and SQL command in one line
    - Split the transaction-number and SQL by ```,```
    - Need not ```BEGIN``` statement at the begining of commands.
  - like...
```
1,UPDATE t1 SET c2 = 70 WHERE id = 4
2,SELECT * from t1 WHERE id = 4
1,ROLLBACK
2,COMMIT
```

- options
  - ```-c```: to specify a file includes SQLs for internal ```c``` command
  - ```-host(H)``` hostname of your MySQL
  - ```-init(i)``` to specify a .sql file to initialize your MySQL.
  - ```-user(u)``` username of your MySQL
  - ```-password(p)``` password of your MySQL
  - ```-database(db or d)``` database name of your MySQL
  - ```-port(P)``` port number(int) of your MySQL

### To Use

1. Prepare the transactions and arrange in the order you want to execute in some file(PLAN-FILE)
1. If needed, you can prepare the SQLs to check any status in MySQL in another file(CHECK-SQLs)
1. Execute with options you need
1. For each line in the (PLAN-FILE), you can check the check lock status by commands in (CHECK-SQLs)
  - s: skip the command
  - c[n]: execute the [n]th check command
  - (enter): execute the command

## Install

To install, use `go get`:

```bash
$ go get -d github.com/tom--bo/mytx
```

## Contribution

1. Fork ([https://github.com/tom--bo/mytx/fork](https://github.com/tom--bo/mytx/fork))
1. Create a feature branch
1. Commit your changes
1. Rebase your local changes against the master branch
1. Run test suite with the `go test ./...` command and confirm that it passes
1. Run `gofmt -s`
1. Create a new Pull Request

## Author

[tom--bo](https://github.com/tom--bo)
