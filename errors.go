package main

import (
	"database/sql"
	"log"

	"github.com/pkg/errors"
)

var DB sql.DB

type Data struct {
	id   int64
	name string
}

func IsExist(id int64) (bool, error) {
	var name string
	err := DB.QueryRow("select name from users where id = ?", 1).Scan(&name)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		} else {
			return false, errors.Wrap(err, "IsExist:")
		}
	}
	return true, nil
}

func GetData(id int64) (string, error) {
	var name string
	err := DB.QueryRow("select name from users where id = ?", 1).Scan(&name)
	if err != nil {
		return "", errors.Wrap(err, "GetData:")
	}
	return name, nil
}

func ConvertName(name string) string {
	return name + "_test"
}
func main() {
	DB, err := sql.Open("mysql",
		"user:password@tcp(127.0.0.1:3306)/hello")
	if err != nil {
		log.Fatal(err)
	}
	defer DB.Close()
	/*
	  关于遇到sql.ErrNoRows如何处理的问题, 具体场景具体分析, 提供两个不同的处理方式的场景
	  1. 如果只是为了检查数据是否存在, 则不需要将sql.ErrNoRows上抛, 按照业务需求正常返回即可, 具体可参考IsExist方法
	  2. 如果是要对查询的数据进行处理, 则需要将sql.ErrNoRows上抛, 由上层进行处理, 具体可参考GetData方法
	*/
	check, err := IsExist(int64(1))
	if err != nil {
		log.Fatalf("err: %v", errors.Cause(err))
		log.Fatalf("stack is %+v", err)
	} else {
		log.Printf("id is %d status is %v", int64(1), check)
	}

	name, err := GetData(int64(1))
	if err != nil {
		log.Fatalf("err: %v", errors.Cause(err))
		log.Fatalf("stack is %+v", err)
	} else {
		newName := ConvertName(name)
		log.Printf("id is %d new name is %s", int64(1), newName)
	}
}
