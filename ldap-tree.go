package main

import (
	"fmt"
	"os"
	"regexp"

	"github.com/go-ldap/ldap/v3"
)

func GetManager(conn *ldap.Conn, uid string) string {
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

func GetHeirarchy(conn *ldap.Conn, uid string) []string {
	managers := []string{}

	manager := GetManager(conn, os.Args[2])
	for manager != "" {
		managers = append(managers, manager)
		manager = GetManager(conn, manager)
	}

	return managers
}

func main() {
	conn, err := ldap.DialURL("ldap://" + os.Args[1] + ":389")
	if err != nil {
		fmt.Println("Failed to DialURL: %s", err)
	}
	defer conn.Close()

	for _, manager := range GetHeirarchy(conn, os.Args[2]) {
		fmt.Println(manager)
	}
}
