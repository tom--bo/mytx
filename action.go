package main

import (
	"bufio"
	// "database/sql"
	"fmt"
	"go/build"
	"log"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

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

func queryTx(n int, sql string) {
	if len(txs) < n {
		for i := len(txs); i < n; i++ {
			createTx()
		}
	}
	n -= 1

	rows, err := txs[n].Queryx(sql)
	checkError(err)
	printRows(6, rows)
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
	// checkSQL := "SELECT * FROM performance_schema.data_locks d INNER JOIN information_schema.innodb_trx i WHERE d.ENGINE_TRANSACTION_ID = i.trx_id"
	checkSQL := "SELECT PARTITION_NAME,INDEX_NAME,LOCK_TYPE,LOCK_MODE,LOCK_STATUS,LOCK_DATA,trx_id,trx_state,trx_started,trx_requested_lock_id,trx_query,trx_operation_state,trx_tables_in_use,trx_tables_locked,trx_lock_structs,trx_rows_locked,trx_rows_modified,trx_adaptive_hash_latched,trx_autocommit_non_locking FROM performance_schema.data_locks d INNER JOIN information_schema.innodb_trx i WHERE d.ENGINE_TRANSACTION_ID = i.trx_id;"

	for _, line := range lines {
		l := strings.SplitN(line, ",", 2)
		n, err := strconv.Atoi(l[0])
		checkError(err)
		sql := strings.Trim(l[1], " ")
		fmt.Printf("%d: %s\n> ", n, sql)

	FLABEL:
		for stdin.Scan() {
			t := stdin.Text()
			switch t {
			case "s":
				fmt.Println("Skiped")
				break FLABEL
			case "c":
				rows, err := db.Queryx(checkSQL)
				checkError(err)

				printRows(26, rows)
				fmt.Printf("%d: %s\n> ", n, sql)
			default:
				if checkRegexp(`(?i)^SELECT`, sql) {
					go queryTx(n, sql)
				} else {
					go execTx(n, sql)
				}
				time.Sleep(50 * time.Millisecond)
				break FLABEL
			}
		}
	}
	return nil
}

func printRows(width int, rows *sqlx.Rows) {
	cols, err := rows.Columns()
	checkError(err)

	cnt := 1
	colstr := fmt.Sprintf("%%%dv: ", width)
	for rows.Next() {
		fmt.Printf("====================== row: %2d ====================== \n", cnt)
		ret := make(map[string]interface{})
		err = rows.MapScan(ret)
		for _, cname := range cols {
			fmt.Printf(colstr, cname)
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
	fmt.Println()
}

func checkRegexp(reg, str string) bool {
	return regexp.MustCompile(reg).Match([]byte(str))
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
