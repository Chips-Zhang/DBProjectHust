package tools

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-pg/pg"
)

const (
	PermAdmin        = "root"
	PermCashier      = "cashier"
	PermCustomerServ = "customer_service"
	PermCustomer     = "customer"
	PermPublic       = "public"

	RoleAdmin          = "root"
	ROLE_CASHIER       = "cashier"
	ROLE_CUSTOMER      = "customer"
	ROLE_CUSTOMER_SERV = "customer_service"
)

const PasswordSalt = "rsalt"
const RootUid = 1
const PgNotFoundErr = "pg: no rows in result set"
const EarningPerAdduser = 1000 // Earn 10 dollar for every new customer
var RolesPermission map[string][]string

var MoneyStrRegex = regexp.MustCompile(`^[0-9]+\.[0-9][0-9]$`)
var UsernameRegex = regexp.MustCompile(`^[A-Za-z0-9_, @.:;-]*$`)

func InitCommon() {
	RolesPermission = map[string][]string{
		"root":     {PermAdmin, PermCashier, PermCustomerServ, PermCustomer, PermPublic},
		"cash":     {PermCashier, PermPublic},
		"serv":     {PermCustomerServ, PermPublic},
		"customer": {PermCustomer, PermPublic},
	}
}

// common data structure

type UidT int64
type MoneyT int64
type PlanidT int64

func (m MoneyT) String() string {
	symbol := ""
	high := "0"
	var low string
	abs := m
	if abs < 0 {
		abs = -abs
		symbol = "-"
	}
	if abs < 10 {
		low = "0" + strconv.Itoa(int(abs))
	} else if abs < 100 {
		low = strconv.Itoa(int(abs))
	} else {
		low = strconv.Itoa(int(abs % 100))
	}

	high = strconv.FormatInt(int64(abs)/100, 10)

	return fmt.Sprintf("%s%s.%s", symbol, high, low)
}
func StringToMoneyT(s string) (MoneyT, error) {
	dotIndex := strings.Index(s, ".")
	strWithoutDot := ""
	s += "00"
	if dotIndex == -1 {
		strWithoutDot = s
	} else {
		strWithoutDot = s[:dotIndex] + s[dotIndex+1:dotIndex+3]
	}
	m, err := strconv.ParseInt(strWithoutDot, 10, 64)
	if err != nil {
		return 0, err
	} else {
		return MoneyT(m), nil
	}
}

type UserInfo struct {
	Id           UidT   `sql:",pk,unique"`
	Name         string `sql:",unique"`
	Password     string
	Permissions  []string `sql:",array"`
	Balance      MoneyT   // `10.32` is saved as `1032`
	Achievements MoneyT
	Plan         PlanidT `sql:",notnull"`
	Email        string  `sql:",unique"`
}

func (u UserInfo) String() string {
	return fmt.Sprintf("UserInfo<%d %s %s BAL=%d ACHI=%d>", u.Id, u.Name, strings.Join(u.Permissions, ","), u.Balance, u.Achievements)
}

type UserBalanceEvent struct {
	EventId UidT `sql:",pk,unique"`
	UId     UidT
	What    string // should not contain `;`
}

func (u UserBalanceEvent) String() string {
	return fmt.Sprintf("UserBalanceEvent<%d %d %s>", u.EventId, u.UId, u.What)
}

type PlanInfo struct {
	Id    PlanidT `sql:",pk,unique"`
	Name  string  `sql:",unique"`
	Price MoneyT
}

func (p PlanInfo) String() string {
	return fmt.Sprintf("PlanInfo<%d %s %d>", p.Id, p.Name, p.Price)
}

// database

var DB_ *pg.DB

// other common functions
func ArrayContains(arr []string, item string) bool {
	for _, ele := range arr {
		if ele == item {
			return true
		}
	}
	return false
}

func CheckPermission(commiter UidT, perm string) bool {
	commiterInfo := UserInfo{
		Id: commiter,
	}
	err := DB_.Select(&commiterInfo)
	if err != nil {
		return false
	}

	if !ArrayContains(commiterInfo.Permissions, perm) {
		return false
	}
	return true
}

func UsernameToInfo(name string) (UserInfo, error) {
	if !UsernameRegex.MatchString(name) {
		return UserInfo{}, errors.New("Invalid username format.")
	}
	u := UserInfo{}
	err := DB_.Model(&u).Where(fmt.Sprintf("name = '%s'", name)).Select()
	if err != nil && err.Error() == PgNotFoundErr {
		return u, errors.New("Name not found: " + name)
	}
	return u, err
}
func PlannameToInfo(name string) (PlanInfo, error) {
	if !UsernameRegex.MatchString(name) {
		return PlanInfo{}, errors.New("Invalid planname format.")
	}
	u := PlanInfo{}
	err := DB_.Model(&u).Where(fmt.Sprintf("name = '%s'", name)).Select()
	if err != nil && err.Error() == PgNotFoundErr {
		return u, errors.New("Name not found: " + name)
	}
	return u, err
}
