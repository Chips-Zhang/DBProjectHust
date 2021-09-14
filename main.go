package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/Chips-zhang/DBProjectHust/service"
	"github.com/Chips-zhang/DBProjectHust/tools"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
)

func createTables() {

	for _, model := range []interface{}{&tools.UserInfo{}, &tools.UserBalanceEvent{}, &tools.PlanInfo{}} {
		err := tools.DB_.CreateTable(model, &orm.CreateTableOptions{
			IfNotExists: true,
		})

		if err != nil {
			panic("Unable to create table: " + err.Error())
		}
	}
}

func tryCreateRootAccount(password string) {
	u := tools.UserInfo{Id: tools.RootUid}
	err := tools.DB_.Select(&u)

	if err != nil {
		if err.Error() == tools.PgNotFoundErr {
			// root not existing
			// create root account
			u.Password = tools.PasswordSaltedHash(password, tools.PasswordSalt)
			u.Name = "root"
			u.Permissions = tools.RolesPermission[tools.RoleAdmin]
			u.Email = "root@recolic.net"
			err2 := tools.DB_.Insert(&u)
			if err2 != nil {
				panic("Unable to insert root record. " + err2.Error())
			}
		} else {
			panic("Unable to select root record. " + err.Error())
		}
	}
}

func main() {
	tools.InitAuthModule()
	tools.InitCommon()
	dbUsername := flag.String("user", "postgres", "Username for PostgreSQL.")
	dbAddr := flag.String("addr", "127.0.0.1:5432", "Address for PostgreSQL.")
	dbPswd := flag.String("password", "", "Password for PostgreSQL.")
	httpBindAddr := flag.String("listen", ":80", "Listen address for http server.")
	defaultRootPassword := flag.String("root-password", "P@ssw0rd", "For first-time launch, set this parameter to set root password.")

	flag.Parse()

	log.Printf("Connecting PostgreSQL %s as %s...", *dbAddr, *dbUsername)
	tools.DB_ = pg.Connect(&pg.Options{
		User:     *dbUsername,
		Addr:     *dbAddr,
		Password: *dbPswd,
	})
	defer tools.DB_.Close()

	// create table if not exist
	createTables()

	tryCreateRootAccount(*defaultRootPassword)

	log.Printf("HTTP listening %s.", *httpBindAddr)
	http.HandleFunc("/", service.HttpApiFunc)
	err := http.ListenAndServe(*httpBindAddr, nil)

	if err != nil {
		panic(err.Error())
	}
}
