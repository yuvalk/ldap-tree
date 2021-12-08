package main

import (
	"fmt"
	"os"
	"regexp"

	"github.com/go-ldap/ldap/v3"
)

func Get_Manager(conn *ldap.Conn, uid string) string {
	searchRequest := ldap.NewSearchRequest(
		"dc=redhat,dc=com",
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		"(&(uid="+uid+"))",
		[]string{"dn", "cn", "manager"},
		nil,
	)

	sr, err := conn.Search(searchRequest)
	if err != nil {
		fmt.Println("Failed to search: %s", err)
	}

	manager := sr.Entries[0].GetAttributeValue("manager")
	re := regexp.MustCompile("uid=(.*?),")
	matches := re.FindSubmatch([]byte(manager))
	if matches != nil {
		return string(matches[1])
	}
	return ""
}

func main() {
	conn, err := ldap.DialURL("ldap://" + os.Args[1] + ":389")
	if err != nil {
		fmt.Println("Failed to DialURL: %s", err)
	}
	defer conn.Close()

	manager := Get_Manager(conn, os.Args[2])
	for manager != "" {
		fmt.Println(manager)
		manager = Get_Manager(conn, manager)
	}
}
