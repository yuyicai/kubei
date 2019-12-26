package text

import "fmt"

func ResetHosts(apiDomainName string) string {
	cmdText := "sed -i '/%s/d' /etc/hosts"
	return fmt.Sprintf(cmdText, apiDomainName)
}

func RemoveKube() string {

	return ""
}
