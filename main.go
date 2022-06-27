package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type user struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

type Arguments map[string]string

var validOperations = map[string]bool{
	"add":      true,
	"remove":   true,
	"list":     true,
	"findById": true,
}

func Perform(args Arguments, writer io.Writer) error {
	if args == nil {
		return fmt.Errorf("error: args is empty: %v", args)
	}

	operation, ok := args["operation"]
	if !ok || len(operation) == 0 {
		return errors.New("-operation flag has to be specified")
	}
	_, ok = validOperations[operation]
	if !ok {
		return fmt.Errorf("Operation %v not allowed!", operation)
	}

	fileName, ok := args["fileName"]
	if !ok || len(fileName) == 0 {
		return errors.New("-fileName flag has to be specified")
	}

	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			fmt.Println("defer err", err)
		}
	}(f)
	allUsersBs, err := ioutil.ReadAll(f)
	if err != nil {
		log.Println("ReadAll", err)
		return err
	}

	var allUsers []user
	if len(allUsersBs) != 0 {
		err = json.Unmarshal(allUsersBs, &allUsers)
		if err != nil {
			return err
		}
	}

	if operation == "list" {
		bsNbr, err := fmt.Fprintf(writer, string(allUsersBs))
		if err != nil {
			return err
		}
		fmt.Printf("%d bytes written writer\n", bsNbr)
	}

	if operation == "add" {
		item, ok := args["item"]
		if !ok || len(item) == 0 {
			return errors.New("-item flag has to be specified")
		}

		var u user
		err := json.Unmarshal([]byte(item), &u)
		if err != nil {
			return err
		}

		var same bool
		for _, user := range allUsers {
			if user.ID == u.ID {
				same = true
				bsNbr, err := fmt.Fprintf(writer, fmt.Errorf("Item with id %s already exists", u.ID).Error())
				if err != nil {
					return err
				}
				fmt.Printf("%d bytes written writer \n", bsNbr)
				return nil
			}
		}
		if same {
			return nil
		}
		allUsers = append(allUsers, u)
		bs, err := json.Marshal(allUsers)
		if err != nil {
			return err
		}

		err = os.Truncate(args["fileName"], 0)
		if err != nil {
			return err
		}

		fmt.Printf("%s truncated successfully", args["fileName"])

		write, err := f.Write(bs)
		if err != nil {
			return err
		}
		fmt.Printf("%d bytes written %s\n", write, f.Name())

	} else if operation == "findById" {
		id, ok := args["id"]
		if !ok || len(id) == 0 {
			return errors.New("-id flag has to be specified")
		}

		for _, user := range allUsers {
			if user.ID == id {
				bs, err := json.Marshal(user)
				if err != nil {
					return err
				}
				bsNbr, err := fmt.Fprintf(writer, string(bs))
				if err != nil {
					return err
				}
				fmt.Printf("%d bytes written writer \n", bsNbr)
				return nil
			}
		}
		bsNbr, err := fmt.Fprintf(writer, "")
		if err != nil {
			return err
		}
		fmt.Printf("%d bytes written writer \n", bsNbr)

		return nil
	} else if operation == "remove" {

		id, ok := args["id"]
		if !ok || len(id) == 0 {
			return errors.New("-id flag has to be specified")
		}

		for i, user := range allUsers {
			if user.ID == id {
				allUsers = append(allUsers[:i], allUsers[i+1:]...)

				bs, err := json.Marshal(allUsers)
				if err != nil {
					return err
				}
				err = os.Truncate(args["fileName"], 0)
				if err != nil {
					return err
				}

				fmt.Printf("%s truncated successfully", args["fileName"])
				bsNbr, err := f.Write(bs)
				if err != nil {
					return err
				}
				fmt.Printf("%d bytes written %s \n", bsNbr, f.Name())
				return nil
			}

		}

		bsNbr, err := fmt.Fprintf(writer, errors.New("Item with id 2 not found").Error())
		if err != nil {
			return err
		}
		fmt.Printf("%d bytes written writer \n", bsNbr)
		return nil
	}

	return nil
}

func parseArgs() Arguments {
	aa := os.Args
	fmt.Println("initial data: ", aa)
	if len(aa) <= 1 {
		return nil
	}

	aa = aa[1:]
	fmt.Printf("aa %#v\n", aa)

	var operation, item, fileName, id string

	flag.StringVar(&operation, "operation", "defaultValue", "")
	flag.StringVar(&item, "item", "defaultValue", "")
	//item = *flag.String("item", "def", "")
	flag.StringVar(&fileName, "fileName", "defaultValue", "")
	flag.StringVar(&id, "id", "defaultValue", "")

	flag.Parse()

	fmt.Println(flag.NFlag())
	fmt.Println(flag.NArg())
	fmt.Println(item)

	mm := make(Arguments)

	operation = strings.Trim(operation, "«»")
	mm["operation"] = operation

	cc0 := 0
	cc1 := 0
	for i, el := range aa {
		if el == "-item" {
			cc0 = i + 1
		}
		if el == "-fileName" {
			cc1 = i
		}
	}

	items := aa[cc0:cc1]

	var itemJS string
	for _, el := range items {
		fmt.Println("el==>", el)
		el = strings.ReplaceAll(el, "«", "\"")
		el = strings.ReplaceAll(el, "»", "\"")
		el = strings.ReplaceAll(el, "’", "")
		itemJS += el
	}
	itemJS = strings.Trim(itemJS, "‘")

	var u user
	err := json.Unmarshal([]byte(itemJS), &u)
	if err != nil {
		return nil
	}

	m := map[string]interface{}{
		"age":   u.Age,
		"email": u.Email,
		"id":    u.ID,
	}

	bs, err := json.Marshal(m)
	if err != nil {
		return nil
	}

	mm["item"] = string(bs)

	fileName = aa[cc1+1]
	fileName = strings.Trim(fileName, "«»")
	mm["fileName"] = fileName

	fmt.Println(mm)
	return mm
}

func main() {

	err := Perform(parseArgs(), os.Stdout)
	if err != nil {
		panic(err)
	}

}
