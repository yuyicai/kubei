package tmpl

import "fmt"

func ResetHosts(apiDomainName string) string {
	cmdTmpl := "sed -i '/%s/d' /etc/hosts"
	return fmt.Sprintf(cmdTmpl, apiDomainName)
}
