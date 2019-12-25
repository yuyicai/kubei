package text

import (
	"fmt"
	"github.com/lithammer/dedent"
)

func Restart(name string) string {
	cmdText := dedent.Dedent(`
        systemctl daemon-reload && systemctl enable %s && systemctl restart %s`)
	return fmt.Sprintf(cmdText, name, name)
}

func SetHosts(ip, apiDomainName string) string {
	cmdText := dedent.Dedent(`
        sed -i '/%s/d' /etc/hosts
        cat <<EOF | tee -a /etc/hosts
        %s %s
        EOF`)
	return fmt.Sprintf(cmdText, apiDomainName, ip, apiDomainName)
}

func ChangeHosts(ip, apiDomainName string) string {
	cmdText := "sed -i '/%s/c %s %s' /etc/hosts"
	return fmt.Sprintf(cmdText, apiDomainName, ip, apiDomainName)
}
