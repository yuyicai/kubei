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

func SwapOff() string {
	return dedent.Dedent(`
        swapoff -a && sysctl -w vm.swappiness=0
        sed -i '/swap/ s/^#*/#/' /etc/fstab
	`)
}

func Iptables() string {
	cmd := dedent.Dedent(`
        cat <<EOF | tee /etc/sysctl.d/99-k8s-sysctl.conf 
        net.ipv4.ip_forward=1
        net.bridge.bridge-nf-call-iptables=1
        net.bridge.bridge-nf-call-arptables=1
        net.bridge.bridge-nf-call-ip6tables=1
        EOF
        sysctl --system
	`)
	return cmd
}
