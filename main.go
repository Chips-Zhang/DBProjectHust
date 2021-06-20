package main

import (
	"flag"
	"fmt"
	
	"github.com/Amber-JY/DBProjectHust/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	log "github.com/sirupsen/logrus"
)

func main(){
	// 命令行输入dbname， password
	dbLoginName := flag.String("loginName", "root", "用户名")
	dbPassword := flag.String("loginPassword", "", "密码")
	flag.Parse()

	if *dbPassword == "" {
		log.Error("the password is empty.")
		return
	}

	log.Info("Connectint to mysql as %s", dbLoginName)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/EBill?charset=utf8mb4&parseTime=True&loc=Local", 
		*dbLoginName, *dbPassword, config.DbIP, config.DbPort)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil{
		log.Error(err)
		return
	}



}