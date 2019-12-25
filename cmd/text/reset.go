package text

import "fmt"

func ResetHosts(apiDomainName string) string {
	cmdText := "sed -i '/%s/d' /etc/hosts"
	return fmt.Sprintf(cmdText, apiDomainName)
}

func ResetKubeadm() string {

	return ""
}

func RemoveKube() string {

	return ""
}
