package ssh

import (
	"bufio"
	"bytes"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"k8s.io/klog"
	"net"
)

type Client struct {
	client *ssh.Client
	host   string
}

func Connect(host, port, user, passwd, key string) (*Client, error) {
	config := &ssh.ClientConfig{
		User:            user,
		HostKeyCallback: ssh.HostKeyCallback(func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil }),
	}

	if key != "" {
		signer, err := makePrivateKeySignerFromFile(key)
		if err != nil {
			return nil, err
		}
		config.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	} else {
		config.Auth = []ssh.AuthMethod{ssh.Password(passwd)}
	}

	client, err := ssh.Dial("tcp", net.JoinHostPort(host, port), config)
	if err != nil {
		return nil, err
	}
	return &Client{client: client, host: host}, nil
}

func ConnectByJumpServer(host, port, user, passwd, key string, jumpServer *Client) (*Client, error) {
	config := &ssh.ClientConfig{
		User:            user,
		HostKeyCallback: ssh.HostKeyCallback(func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil }),
	}

	if key != "" {
		signer, err := makePrivateKeySignerFromFile(key)
		if err != nil {
			return nil, err
		}
		config.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	} else {
		config.Auth = []ssh.AuthMethod{ssh.Password(passwd)}
	}

	conn, err := jumpServer.client.Dial("tcp", net.JoinHostPort(host, port))
	if err != nil {
		return nil, err
	}
	ncc, chans, reqs, err := ssh.NewClientConn(conn, net.JoinHostPort(host, port), config)
	if err != nil {
		return nil, err
	}

	return &Client{client: ssh.NewClient(ncc, chans, reqs), host: host}, nil
}

func makePrivateKeySignerFromFile(key string) (ssh.Signer, error) {

	buffer, err := ioutil.ReadFile(key)
	if err != nil {
		return nil, fmt.Errorf("unable to read private key %s: %v", key, err)
	}

	signer, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil, fmt.Errorf("unable to parse private key: %v", err)
	}

	return signer, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}

func (c *Client) Run(cmd string) error {
	//host, _, _ := net.SplitHostPort(c.client.RemoteAddr().String())
	klog.V(7).Infof("[%s] [commands] Execute commands: \n%s", c.host, cmd)
	session, err := c.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	var stdout io.Reader
	stdout, err = session.StdoutPipe()
	if err != nil {
		return err
	}

	if err := session.Start(cmd); err != nil {
		return err
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		str := scanner.Text()
		if str != "" {
			klog.V(8).Infof("[%s] [remote-output] %s", c.host, str)
		}
	}

	return session.Wait()
}

func (c *Client) RunOut(cmd string) ([]byte, error) {
	//host, _, _ := net.SplitHostPort(c.client.RemoteAddr().String())
	klog.V(7).Infof("[%s] [commands] Execute commands: \n%s", c.host, cmd)
	session, err := c.client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	var buf bytes.Buffer
	var stdout io.Reader
	stdout, err = session.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := session.Start(cmd); err != nil {
		return nil, err
	}

	tee := io.TeeReader(stdout, &buf)
	scanner := bufio.NewScanner(tee)

	for scanner.Scan() {
		str := scanner.Text()
		if str != "" {
			klog.V(8).Infof("[%s] [remote-output] %s", c.host, str)
		}
	}

	return buf.Bytes(), session.Wait()
}
