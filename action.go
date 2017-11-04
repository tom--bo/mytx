package main

import (
	"bufio"
	// "database/sql"
	"fmt"
	"go/build"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/codegangsta/cli"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var db *sqlx.DB
var txs []*sqlx.Tx

func initDB() {
	user := "root"
	passwd := "mysql"
	dbname := "sample"
	host := "localhost"
	port := 3306

	var err error
	db, err = sqlx.Open("mysql", user+":"+passwd+"@tcp("+host+":"+strconv.Itoa(port)+")/"+dbname+"?loc=Local&parseTime=true")
	if err != nil {
		log.Fatalf("Filed to connect to DB: %s.", err.Error())
	}
}

func createTx() {
	tmp, err := db.Beginx()
	if err != nil {
		log.Fatalf("Filed to create tx: %s.", err.Error())
	}
	txs = append(txs, tmp)
}

func execTx(n int, sql string) {
	if len(txs) < n {
		for i := len(txs); i < n; i++ {
			createTx()
		}
	}
	n -= 1

	if sql == "COMMIT" {
		err := txs[n].Commit()
		if err != nil {
			log.Fatal(err)
		}
	} else if sql == "ROLLBACK" {
		err := txs[n].Rollback()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		stmt, err := txs[n].Preparex(sql)
		if err != nil {
			log.Fatal(err)
		}
		_, err = stmt.Exec()
		if err != nil {
			log.Fatal(err)
		}
		stmt.Close()
	}
}

func readPlan(planPath string) []string {
	var ret []string

	fp, err := os.Open(planPath)
	checkError(err)
	defer fp.Close()

	scanner := bufio.NewScanner(fp)
	for scanner.Scan() {
		ret = append(ret, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}

	return ret
}

func mainAction(c *cli.Context) error {
	initDB()
	stdin := bufio.NewScanner(os.Stdin)

	planPath := build.Default.GOPATH + "/src/github.com/tom--bo/mytx/samples/plan.txt"
	if c.NArg() > 0 {
		planPath = c.Args().Get(0)
	}
	lines := readPlan(planPath)
	checkSQL := "SELECT * FROM performance_schema.data_locks d INNER JOIN information_schema.innodb_trx i WHERE d.ENGINE_TRANSACTION_ID = i.trx_id"

	for _, line := range lines {
		l := strings.Split(line, ",")
		n, err := strconv.Atoi(l[0])
		checkError(err)
		fmt.Printf("%s: %s\n> ", l[0], l[1])

		for stdin.Scan() {
			t := stdin.Text()
			if t == "s" {
				fmt.Println("Skiped")
				break
			} else if t == "c" {
				rows, err := db.Queryx(checkSQL)
				checkError(err)

				cols, err := rows.Columns()
				checkError(err)
				cnt := 1
				for rows.Next() {
					fmt.Printf("====================== row: %2d ====================== \n", cnt)
					ret := make(map[string]interface{})
					err = rows.MapScan(ret)
					for _, cname := range cols {
						fmt.Printf("%26v: ", cname) // 26
						r := reflect.ValueOf(ret[cname])
						// Sliceなら[]uint8にassersionでinterface{}を[]uint8に変換してさらにstringまで変換
						if r.Kind() == reflect.Slice {
							fmt.Println(string([]byte(ret[cname].([]uint8))))
						} else {
							fmt.Println(ret[cname])
						}
					}
					cnt++
				}
				fmt.Printf("\n%s: %s\n> ", l[0], l[1])
			} else {
				execTx(n, l[1])
				break
			}
		}
	}
	return nil
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
