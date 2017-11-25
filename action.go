package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/urfave/cli"
)

var db *sqlx.DB
var txs []*sqlx.Tx
var checkSQLs []string

func checkRegexp(reg, str string) bool {
	return regexp.MustCompile(reg).Match([]byte(str))
}

func showHelp() {
	fmt.Println(`
=== Help document ===
	(enter): Execute the current SQL command
	s: Skip, skip executing current command and continue next command
	c: Show lock status. You can change the SQL to check the lock status
	   by specifying .sql files with -c option for executing this scripts.
	h: Show this help.
	`)
}

func initDB(opt Options) {
	cmd := exec.Command("mysql", "-u", opt.user, "-h", opt.host, "-P", strconv.Itoa(opt.port), "-p"+opt.passwd, opt.db)

	if opt.initSQLFilePath != "" {
		readFile := opt.initSQLFilePath
		input, err0 := ioutil.ReadFile(readFile)
		if err0 != nil {
			log.Fatalf("Failed to read an init SQL file: %s", err0.Error())
			panic(err0)
		}
		stdin, _ := cmd.StdinPipe()
		stdin.Write(input)
		stdin.Close()
	}

	var stdErr, stdOut bytes.Buffer
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr

	err1 := cmd.Run()
	if err1 != nil {
		log.Fatalf("Failed to exec init SQL: %s.", err1.Error())
		panic(err1)
	}
	fmt.Println("stdout -> " + stdOut.String())
	fmt.Println("stderr -> " + stdErr.String())

	var err2 error
	db, err2 = sqlx.Open("mysql", opt.user+":"+opt.passwd+"@tcp("+opt.host+":"+strconv.Itoa(opt.port)+")/"+opt.db+"?loc=Local&parseTime=true")
	if err2 != nil {
		log.Fatalf("Filed to connect to DB: %s.", err2.Error())
		panic(err2)
	}

	checkSQLs = getLinesFromFile(opt.checkSQLFilePath)
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
	if err != nil {
		log.Fatalf("Failed to query: %s", err.Error())
		os.Exit(1)
	}
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

func getLinesFromFile(path string) []string {
	var ret []string

	fp, err := os.Open(path)
	if err != nil {
		log.Fatalf("Failed to open the plan file: %s", err.Error())
		os.Exit(1)
	}
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
	initDB(Opt)

	stdin := bufio.NewReader(os.Stdin)

	var planPath string
	if c.NArg() > 0 {
		planPath = c.Args().Get(0)
	} else {
		log.Fatalf("Error: You need to specify a plan file")
	}
	lines := getLinesFromFile(planPath)

	for _, line := range lines {
		l := strings.SplitN(line, ",", 2)
		n, err := strconv.Atoi(l[0])
		if err != nil {
			log.Fatalf("Cannot specify not existing check-SQL: %s", err.Error())
			os.Exit(1)
		}
		sql := strings.Trim(l[1], " ")
		fmt.Printf("%d: %s\n> ", n, sql)
	FLABEL:
		for {
			ln, _, err := stdin.ReadLine()
			if err != nil {
				log.Fatalf("Failed to stdin.ReadLine: %s", err.Error())
				os.Exit(1)
			}
			t := strings.Trim(string(ln), " \t")
			if t == "h" || t == "help" {
				showHelp()
				fmt.Printf("%d: %s\n> ", n, sql)
			} else if t == "s" {
				fmt.Println("Skiped")
				break FLABEL
			} else if strings.HasPrefix(t, "c") {
				execCheckSQL(t)
				fmt.Printf("%d: %s\n> ", n, sql)
			} else if t == "" {
				if checkRegexp(`(?i)^SELECT`, sql) {
					go queryTx(n, sql)
				} else {
					go execTx(n, sql)
				}
				time.Sleep(50 * time.Millisecond)
				break FLABEL
			} else {
				fmt.Println("Only s, c, (enter), h(help) is supported.")
			}
		}
	}
	return nil
}

func execCheckSQL(s string) {
	n := 0
	if s != "c" {
		num := strings.SplitN(s, "c", 2)
		var err error
		n, err = strconv.Atoi(num[1])
		if err != nil {
			fmt.Println("You can specify c[n], [n] is only Integer.")
		}
	}

	if n > len(checkSQLs)-1 || n < 0 {
		fmt.Println("You can specify c[n], [n] is only Integer, and the value of [n] is the line of checkSQLFile specified by -c")
	} else {
		rows, err := db.Queryx(checkSQLs[n])
		if err != nil {
			log.Fatalf("Failed to query: %s", err.Error())
			os.Exit(1)
		}
		printRows(26, rows)
	}
}

func printRows(width int, rows *sqlx.Rows) {
	cols, err := rows.Columns()
	if err != nil {
		log.Fatalf("Failed to get columns: %s", err.Error())
		os.Exit(1)
	}

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
