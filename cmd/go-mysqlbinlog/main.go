// go-mysqlbinlog: a simple binlog tool to sync remote MySQL binlog.
// go-mysqlbinlog supports semi-sync mode like facebook mysqlbinlog.
// see http://yoshinorimatsunobu.blogspot.com/2014/04/semi-synchronous-replication-at-facebook.html
package main

import (
	"bufio"
	"context"
	"database/sql"
	"flag"
	"fmt"
	"github.com/siddontang/go-log/log"
	"os"

	"github.com/pingcap/errors"

	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-mysql-org/go-mysql/replication"

	_ "github.com/go-sql-driver/mysql"
)

var host = flag.String("host", "127.0.0.1", "MySQL host")

//var host = flag.String("host", "10.2.5.130", "MySQL host")
var port = flag.Int("port", 3306, "MySQL port")
var user = flag.String("user", "root", "MySQL user, must have replication privilege")
var password = flag.String("password", "", "MySQL password")

var flavor = flag.String("flavor", "mysql", "Flavor: mysql or mariadb")

var file = flag.String("file", "", "Binlog filename")
var pos = flag.Int("pos", 4, "Binlog position")
var gtid = flag.String("gtid", "", "Binlog GTID set that this slave has executed")

var semiSync = flag.Bool("semisync", false, "Support semi sync")
var backupPath = flag.String("backup_path", "", "backup path to store binlog files")

var rawMode = flag.Bool("raw", false, "Use raw mode")

func handle() {
	fileName := "output.txt"
	dstFile, err := os.Create(fileName)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer dstFile.Close()

	// open the file
	file, err := os.Open("./binlog.txt")

	// handle errors while opening
	if err != nil {
		log.Fatalf("Error when opening file: %s", err)
	}
	defer file.Close()

	fileScanner := bufio.NewScanner(file)

	// read line by line
	for fileScanner.Scan() {
		//fmt.Println(fileScanner.Text())
		dstFile.WriteString(fileScanner.Text() + "\n")
	}
	// handle first encountered error while reading
	if err := fileScanner.Err(); err != nil {
		log.Fatalf("Error while reading file: %s", err)
		//dstFile.WriteString("Error while reading file")
	}
}

func main() {
	flag.Parse()

	cfg := replication.BinlogSyncerConfig{
		ServerID: 101,
		Flavor:   *flavor,

		Host:            *host,
		Port:            uint16(*port),
		User:            *user,
		Password:        *password,
		RawModeEnabled:  *rawMode,
		SemiSyncEnabled: *semiSync,
		UseDecimal:      true,
	}

	b := replication.NewBinlogSyncer(cfg)

	pos := mysql.Position{Name: *file, Pos: uint32(*pos)}
	if len(*backupPath) > 0 {
		// Backup will always use RawMode.
		err := b.StartBackup(*backupPath, pos, 0)
		if err != nil {
			fmt.Printf("Start backup error: %v\n", errors.ErrorStack(err))
			return
		}
	} else {
		var (
			s   *replication.BinlogStreamer
			err error
		)
		if len(*gtid) > 0 {
			gset, err := mysql.ParseGTIDSet(*flavor, *gtid)
			if err != nil {
				fmt.Printf("Failed to parse gtid %s with flavor %s, error: %v\n",
					*gtid, *flavor, errors.ErrorStack(err))
			}
			s, err = b.StartSyncGTID(gset)
			if err != nil {
				fmt.Printf("Start sync by GTID error: %v\n", errors.ErrorStack(err))
				return
			}
		} else {
			s, err = b.StartSync(pos)
			if err != nil {
				fmt.Printf("Start sync error: %v\n", errors.ErrorStack(err))
				return
			}
		}

		//go handle()
		db, err := sql.Open("mysql", "root:xxxx@tcp(127.0.0.1:3306)/test2")
		if err != nil {
			panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
		}
		defer db.Close()

		// Execute the query
		//rows1, err := db.Query("SELECT * FROM product2")
		//if err != nil {
		//	panic(err.Error()) // proper error handling instead of panic in your app
		//}
		//
		//// Get column names
		//columns, err := rows1.Columns()
		//if err != nil {
		//	panic(err.Error()) // proper error handling instead of panic in your app
		//}

		// Prepare statement for inserting data
		stmtIns, err := db.Prepare("INSERT INTO product2 VALUES( ?, ?, ? )") // ? = placeholder
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}
		defer stmtIns.Close() // Close the statement when we leave main() / the program terminates

		stmtIns2, err := db.Prepare("UPDATE product2 SET product_name = ? , maker = ? , category = ? WHERE product_name = ? AND maker = ? AND category = ?") // ? = placeholder
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}
		defer stmtIns2.Close() // Close the statement when we leave main() / the program terminates

		stmtIns3, err := db.Prepare("DELETE FROM product2 WHERE product_name = ? AND maker = ? AND category = ?") // ? = placeholder
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}
		defer stmtIns3.Close() // Close the statement when we leave main() / the program terminates

		for {
			e, err := s.GetEvent(context.Background())
			if err != nil {
				panic(err.Error()) // proper error handling instead of panic in your app
			}
			//if err != nil {
			// Try to output all left events
			//events := s.DumpEvents()
			//for _, e := range events {
			//e.Dump(os.Stdout)
			switch v := e.Event.(type) {
			case *replication.RowsEvent:
				if e.Header.EventType == replication.WRITE_ROWS_EVENTv2 {
					for _, rows := range v.Rows {
						_, err := stmtIns.Exec(rows[0], rows[1], rows[2])
						if err != nil {
							panic(err.Error()) // proper error handling instead of panic in your app
						}
					}
				}
				if e.Header.EventType == replication.UPDATE_ROWS_EVENTv2 {
					var oldRows [3]interface{}
					for count, rows := range v.Rows {
						if count%2 == 0 {
							oldRows[0] = rows[0]
							oldRows[1] = rows[1]
							oldRows[2] = rows[2]
						} else {
							_, err := stmtIns2.Exec(rows[0], rows[1], rows[2], oldRows[0], oldRows[1], oldRows[2])
							if err != nil {
								panic(err.Error()) // proper error handling instead of panic in your app
							}
						}
					}
				}
				if e.Header.EventType == replication.DELETE_ROWS_EVENTv2 {
					for _, rows := range v.Rows {
						_, err := stmtIns3.Exec(rows[0], rows[1], rows[2])
						if err != nil {
							panic(err.Error()) // proper error handling instead of panic in your app
						}
					}
				}
				fmt.Fprintf(os.Stdout, "=== %s ===\n", e.Header.EventType)
				fmt.Fprintf(os.Stdout, "TableID: %d\n", v.TableID)
				fmt.Fprintf(os.Stdout, "Flags: %d\n", v.Flags)
				fmt.Fprintf(os.Stdout, "Column count: %d\n", v.ColumnCount)

				fmt.Fprintf(os.Stdout, "Values:\n")
				for _, rows := range v.Rows {
					fmt.Fprintf(os.Stdout, "--\n")
					for j, d := range rows {
						if _, ok := d.([]byte); ok {
							fmt.Fprintf(os.Stdout, "%d:%q\n", j, d)
						} else {
							fmt.Fprintf(os.Stdout, "%d:%#v\n", j, d)
						}
					}
				}
				fmt.Fprintln(os.Stdout)
			default:
			}
			//}
			//fmt.Printf("Get event error: %v\n", errors.ErrorStack(err))
			//return
			//}

			//e.Dump(os.Stdout)
		}
	}

}
