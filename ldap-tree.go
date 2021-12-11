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

func GetHierarchy(conn *ldap.Conn, uid string) []string {
	managers := []string{uid}

	manager := GetManager(conn, uid)
	for manager != "" {
		managers = append(managers, manager)
		manager = GetManager(conn, manager)
	}

	return managers
}

func PrintDot(managers []string) {
	fmt.Println("digraph regexp {")
	for i, manager := range managers {
		fmt.Printf("n%d [label=\"%s\"];\n", i, manager)
		if i == 0 {
			continue
		}
		fmt.Printf("n%d -> n%d;\n", i-1, i)
	}
	fmt.Println("}")
}

func PrintDot2(managers1 []string, managers2 []string) {
	fmt.Println("digraph regexp {")

	common := 0
	for i := 0; i < len(managers1); i++ {
        m1 := len(managers1) - i - 1
        m2 := len(managers2) - 1
		if managers1[m1] == managers2[m2] {
	        fmt.Printf("n%d [label=\"%s\" color=red];\n", i, managers1[m1])
			managers2 = managers2[:len(managers2)-1]
		} else {
	        fmt.Printf("n%d [label=\"%s\"];\n", i, managers1[m1])
			if common == 0 {
				common = i - 1
            }
		}

		if i == len(managers1)-1 {
			continue
		}
		fmt.Printf("n%d -> n%d;\n", i, i+1)
	}

	last := 0
	for i, manager := range managers2 {
		fmt.Printf("n%d [label=\"%s\"];\n", len(managers1)+i, manager)
		if i == 0 {
			continue
		}
		fmt.Printf("n%d -> n%d;\n", len(managers1)+i, len(managers1)+i-1)
		last = len(managers1) + i
	}
	fmt.Printf("n%d -> n%d;\n", common, last)

	fmt.Println("}")
}

func main() {
	conn, err := ldap.DialURL("ldap://" + os.Args[1] + ":389")
	if err != nil {
		fmt.Println("Failed to DialURL: %s", err)
	}
	defer conn.Close()

	hei1 := GetHierarchy(conn, os.Args[2])
	hei2 := GetHierarchy(conn, os.Args[3])

	PrintDot2(hei1, hei2)
}
