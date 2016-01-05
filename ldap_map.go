package main

import (
	"errors"
	"fmt"
	"github.com/nmcclain/ldap"
)

// LdapMap provides a means to map email addresses returned from Oauth2 providers to uids provided in Ldap.
// Ldap must contain an attribute where the email address can be matched for a user. If found, the
// headers X-Forwarded-User and X-Forwarded-Email will be substituted upstreams.
func LdapMap(p *OAuthProxy, email string) (string, string, error) {
	luser := ""
	lemail := ""

	// Connect to the LDAP server
	l, err := ldap.Dial("tcp", p.LdapUri)
	// be sure to add error checking!
	defer l.Close()

	// Bind using bind user
	err = l.Bind(p.LdapBindUser, p.LdapBindPass)
	if err != nil {
		return "", "", err
	}

	// Setup filter etc and search
	filter := fmt.Sprintf("(%s=%s)", p.LdapFilterAttribute, email)
	attributes := []string{p.LdapUidAttribute, p.LdapMailAttribute}
	search := ldap.NewSearchRequest(
		p.LdapSearchBase,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		filter,
		attributes,
		nil)
	searchResults, err := l.Search(search)

	// Handle errors, lack of results etc
	if err == nil {
		if len(searchResults.Entries) == 0 {
			return "", "", errors.New("No matching user found")
		}
		luser = searchResults.Entries[0].GetAttributeValue(p.LdapUidAttribute)
		lemail = searchResults.Entries[0].GetAttributeValue(p.LdapMailAttribute)
	} else {
		return "", "", err
	}

	// Return successful match
	return luser, lemail, nil
}
