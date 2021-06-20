/*
项目中相关表的结构
 */

package model

// UserInfo 用户信息定义
type UserInfo struct{
	UID int64 `gorm:"primary_key;unique"`
	Name string `gorm:"unique"`
	Password string `gorm:"not null"`
	Balance int64 `gorm:"not null"`
	Plan int32 `gorm:"not null"`
	State bool `gorm:"not null"`
}

// ServerInfo 客服信息定义
type ServerInfo struct{

}

// CashierInfo 收款员定义
type CashierInfo struct{

}

// Bill 账单定义
type Bill struct{

}

// Plan 套餐定义
type Plan struct{

}