package service

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Chips-zhang/DBProjectHust/tools"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
)

func checkUserUpdatePermission(commiter tools.UidT, updatedUserPerm []string) bool {
	updatedUserIsCustomer := true
	if tools.ArrayContains(updatedUserPerm, tools.PermAdmin) ||
		tools.ArrayContains(updatedUserPerm, tools.PermCashier) ||
		tools.ArrayContains(updatedUserPerm, tools.PermCustomerServ) {
		updatedUserIsCustomer = false
	}

	if commiter != tools.RootUid {
		// if uid is 1(root), just skip all check. so that system can create uid 1 without permission.
		if !updatedUserIsCustomer {
			// Admin can create non-customer user.
			if tools.CheckPermission(commiter, tools.PermAdmin) == false {
				return false
			}
		} else {
			// costomer_serv can create customer user.
			if tools.CheckPermission(commiter, tools.PermCustomerServ) == false {
				return false
			}
		}
	}
	return true
}

func AddUser(commiter tools.UidT, name, password, permissions, email string) (tools.UidT, error) {
	// password is already salted-hashed in client.
	// TODO: role maybe a comma-seperated string as permission list.
	perm := strings.Split(permissions, ",")
	if checkUserUpdatePermission(commiter, perm) == false {
		return -1, errors.New("Permission denied.")
	}

	newUid := tools.UidT(0)

	err := tools.DB_.RunInTransaction(func(tx *pg.Tx) error {
		u := tools.UserInfo{
			Name:         name,
			Password:     password,
			Permissions:  perm,
			Balance:      0,
			Achievements: 0,
			Email:        email,
		}

		err := tx.Insert(&u)
		if err != nil {
			return err
		}
		newUid = u.Id

		if tools.ArrayContains(perm, tools.PermCustomer) {
			// The CustomerService is introducing new customer. Give him salary!
			commiter := tools.UserInfo{
				Id: commiter,
			}
			err2 := tx.Select(&commiter)
			if err2 != nil {
				return err2
			}

			commiter.Achievements += tools.EarningPerAdduser
			return tx.Update(&commiter)
		} else {
			return nil
		}
	})

	return newUid, err
}

func RemoveUser(commiter tools.UidT, fuckedUsername string) error {
	fuckedUser, err := tools.UsernameToInfo(fuckedUsername)
	if err != nil {
		return err
	}

	if checkUserUpdatePermission(commiter, fuckedUser.Permissions) == false {
		return errors.New("Permission denied.")
	}

	return tools.DB_.Delete(&tools.UserInfo{Id: fuckedUser.Id})
}

func AddPlan(commiter tools.UidT, planName string, planPriceStr string) (tools.PlanidT, error) {
	if tools.CheckPermission(commiter, tools.PermCustomerServ) == false {
		return -1, errors.New("Permission denied")
	}

	planPrice, err0 := tools.StringToMoneyT(planPriceStr)
	if err0 != nil {
		return -1, err0
	}

	p := tools.PlanInfo{
		Name:  planName,
		Price: planPrice,
	}

	err := tools.DB_.Insert(&p)
	if err != nil {
		return -1, err
	}

	return p.Id, nil
}

func RemovePlan(commiter tools.UidT, fuckedPlanname string) error {
	fucked, err := tools.PlannameToInfo(fuckedPlanname)
	if err != nil {
		return err
	}

	if tools.CheckPermission(commiter, tools.PermCustomerServ) == false {
		return errors.New("Permission denied.")
	}

	var userUsingThisPlan []tools.UserInfo
	err = tools.DB_.Model(&userUsingThisPlan).Where("plan = " + strconv.FormatInt(int64(fucked.Id), 10)).Select()
	if err != nil && err.Error() != tools.PgNotFoundErr {
		return err
	}

	if len(userUsingThisPlan) > 0 {
		return errors.New("The plan is still in use by some user: " + userUsingThisPlan[0].Name)
	}

	return tools.DB_.Delete(&tools.PlanInfo{Id: fucked.Id})
}

func UpdateUserPlan(commiter tools.UidT, fuckedUsername string, planName string) error {
	if tools.CheckPermission(commiter, tools.PermCustomerServ) == false {
		return errors.New("Permission denied.")
	}

	u, err := tools.UsernameToInfo(fuckedUsername)
	if err != nil {
		return err
	}

	if tools.CheckPermission(u.Id, tools.PermCustomer) == false {
		return errors.New("Only customer can have a plan.")
	}

	newPlan, err2 := tools.PlannameToInfo(planName)
	if err2 != nil {
		return err2
	}

	u.Plan = newPlan.Id

	return tools.DB_.Update(&u)
}

func UpdateUserBalance(commiter tools.UidT, customerUsername string, balanceChangeStr string) error {
	if tools.CheckPermission(commiter, tools.PermCashier) == false {
		return errors.New("Permission denied.")
	}

	balanceChange, err0 := tools.StringToMoneyT(balanceChangeStr)
	if err0 != nil {
		return err0
	}

	u, err := tools.UsernameToInfo(customerUsername)
	if err != nil {
		return err
	}

	if tools.CheckPermission(u.Id, tools.PermCustomer) == false {
		return errors.New("Only customer can be updated balance.")
	}

	// Pull commiter and customer info.

	cashier := tools.UserInfo{
		Id: commiter,
	}
	err1 := tools.DB_.Select(&cashier)
	if err1 != nil {
		return err1
	}

	// Prepare event
	event := tools.UserBalanceEvent{
		UId: u.Id,
		What: fmt.Sprintf("balance_update %s from %s to %s by %d",
			balanceChange.String(), u.Balance.String(), (u.Balance + balanceChange).String(), commiter),
	}

	u.Balance += balanceChange
	if balanceChange > 0 {
		// cashier receive money and charge user.
		cashier.Achievements += balanceChange
	}

	err2 := tools.DB_.RunInTransaction(func(tx *pg.Tx) error {
		err := tx.Update(&u)
		if err != nil {
			return err
		}

		err2 := tx.Update(&cashier)
		if err2 != nil {
			return err
		}

		return tx.Insert(&event)
	})

	return err2
}

func QueryUserInfo(commiter tools.UidT, usernameToQuery string) (string, error) {
	u, err := tools.UsernameToInfo(usernameToQuery)
	if err != nil {
		return "", err
	}

	if u.Id != commiter {
		if tools.CheckPermission(commiter, tools.PermCustomerServ) == false {
			return "", errors.New("Permission denied.")
		}
	}

	p := tools.PlanInfo{Id: u.Plan}
	err2 := tools.DB_.Select(&p)
	if u.Plan != 0 && err2 != nil {
		return "", err2
	}

	return fmt.Sprintf("name=%s&permission=%s&balance=%s&achi=%s&plan_name=%s&plan_price=%s",
		u.Name, strings.Join(u.Permissions, ","), u.Balance.String(), u.Achievements.String(),
		p.Name, p.Price.String()), nil
}

func QueryBalanceLog(commiter tools.UidT, usernameToQuery string) (string, error) {
	u, err := tools.UsernameToInfo(usernameToQuery)
	if err != nil {
		return "", err
	}

	if u.Id != commiter {
		if tools.CheckPermission(commiter, tools.PermCustomerServ) == false {
			return "", errors.New("Permission denied.")
		}
	}

	var events []tools.UserBalanceEvent
	err2 := tools.DB_.Model(&events).Where(fmt.Sprintf("u_id = %d", u.Id)).Select()
	if err2 != nil {
		if err2.Error() == tools.PgNotFoundErr {
			return "events=", nil
		} else {
			return "", err2
		}
	}

	eventStrs := make([]string, len(events))
	for index, event := range events {
		eventStrs[index] = event.What
	}

	return "events=" + strings.Join(eventStrs, "\n"), nil
}

func ResetDatabase(commiter tools.UidT, newRootPassword string) error {
	if tools.CheckPermission(commiter, tools.PermAdmin) == false {
		return errors.New("Permission denied.")
	}

	err := tools.DB_.RunInTransaction(func(tx *pg.Tx) error {
		for _, model := range []interface{}{&tools.UserInfo{}, &tools.UserBalanceEvent{}, &tools.PlanInfo{}} {
			err := tx.DropTable(model, &orm.DropTableOptions{
				IfExists: true,
				Cascade:  true,
			})
			if err != nil {
				return err
			}
		}

		for _, model := range []interface{}{&tools.UserInfo{}, &tools.UserBalanceEvent{}, &tools.PlanInfo{}} {
			err := tx.CreateTable(model, &orm.CreateTableOptions{
				IfNotExists: true,
			})
			if err != nil {
				return err
			}
		}

		u := tools.UserInfo{
			Name:        "root",
			Permissions: tools.RolesPermission[tools.RoleAdmin],
			Password:    newRootPassword,
		}
		err3 := tx.Insert(&u)
		if err3 != nil {
			return err3
		}
		if u.Id != tools.RootUid {
			return errors.New("ROOT UID is incorrect. Failed to clear db.")
		}
		return nil
	})

	return err
}

func ListAllUserInfo(commiter tools.UidT) (string, error) {
	if tools.CheckPermission(commiter, tools.PermCustomerServ) == false {
		return "", errors.New("Permission denied.")
	}

	var users []tools.UserInfo
	err := tools.DB_.Model(&users).Select()
	if err != nil {
		return "", err
	}

	result := ""

	for _, u := range users {
		p := tools.PlanInfo{Id: u.Plan}
		err2 := tools.DB_.Select(&p)
		if u.Plan != 0 && err2 != nil {
			return "", err2
		}

		result += fmt.Sprintf("name=%s&permission=%s&balance=%s&achi=%s&plan_name=%s&plan_price=%s",
			u.Name, strings.Join(u.Permissions, ","), u.Balance.String(), u.Achievements.String(),
			p.Name, p.Price.String())
		result += "\n"
	}
	return result, nil
}

func ListAllPlanInfo(commiter tools.UidT) (string, error) {
	if tools.CheckPermission(commiter, tools.PermCustomerServ) == false {
		return "", errors.New("Permission denied.")
	}

	var plans []tools.PlanInfo
	err := tools.DB_.Model(&plans).Select()
	if err != nil {
		return "", err
	}

	result := ""

	for _, p := range plans {
		result += fmt.Sprintf("plan_name=%s&plan_price=%s",
			p.Name, p.Price.String())
		result += "\n"
	}
	return result, nil
}
