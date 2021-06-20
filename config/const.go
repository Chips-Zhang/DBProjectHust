package config

// 权限和角色的字符定义
const (
	PermPublic = "public"
	PermAdmin = "root"
	PermCashier = "cashier"
	PermCustomer = "customer"
	PermServer = "server"

	RoleAdmin = "root"
	RoleCashier = "cashier"
	RoleCustomer = "customer"
	RoleServer = "server"
)

const(
	DbIP = "localhost"
	DbPort = "3306"
)

const PasswordSalt = "Salt"