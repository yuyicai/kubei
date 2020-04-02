package ssh

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/sync/errgroup"
	"io"
	"io/ioutil"
	"k8s.io/klog"
	"net"
	"os"
	"strings"
)

type Client struct {
	client   *ssh.Client
	host     string
	password string
	user     string
}

func Connect(host, port, user, password, key string) (*Client, error) {
	config, err := setConf(user, password, key)
	if err != nil {
		return nil, err
	}

	client, err := ssh.Dial("tcp", net.JoinHostPort(host, port), config)
	if err != nil {
		return nil, err
	}
	return &Client{client: client, host: host, password: password, user: user}, nil
}

func ConnectByJumpServer(host, port, user, password, key string, jumpServer *Client) (*Client, error) {
	config, err := setConf(user, password, key)
	if err != nil {
		return nil, err
	}

	conn, err := jumpServer.client.Dial("tcp", net.JoinHostPort(host, port))
	if err != nil {
		return nil, err
	}
	ncc, chans, reqs, err := ssh.NewClientConn(conn, net.JoinHostPort(host, port), config)
	if err != nil {
		return nil, err
	}

	return &Client{client: ssh.NewClient(ncc, chans, reqs), host: host, password: password, user: user}, nil
}

func setConf(user, password, key string) (*ssh.ClientConfig, error) {
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
		config.Auth = []ssh.AuthMethod{ssh.Password(password)}
	}
	return config, nil
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
	cmd = c.cmdPrefix(cmd)

	klog.V(6).Infof("[%s] [commands] Execute commands: \n%s", c.host, cmd)

	session, err := c.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	stdout, err := session.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := session.StderrPipe()
	if err != nil {
		return err
	}
	in, err := session.StdinPipe()
	if err != nil {
		return err
	}

	if err := session.Start(cmd); err != nil {
		return err
	}

	g := errgroup.Group{}
	g.Go(func() error {
		if err := sendSudoPassword(c.password, c.host, in, stderr); err != nil {
			return err
		}
		return nil
	})

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		str := scanner.Text()
		if str != "" {
			klog.V(8).Infof("[%s] [remote-stdout] %s", c.host, str)
		}
	}

	if err := g.Wait(); err != nil {
		return err
	}

	return session.Wait()
}

func (c *Client) RunOut(cmd string) ([]byte, error) {
	//host, _, _ := net.SplitHostPort(c.client.RemoteAddr().String())
	cmd = c.cmdPrefix(cmd)

	klog.V(6).Infof("[%s] [commands] Execute commands: \n%s", c.host, cmd)

	session, err := c.client.NewSession()
	if err != nil {
		return []byte{}, err
	}
	defer session.Close()

	var buf bytes.Buffer

	stdout, err := session.StdoutPipe()
	if err != nil {
		return []byte{}, err
	}
	stderr, err := session.StderrPipe()
	if err != nil {
		return []byte{}, err
	}
	in, err := session.StdinPipe()
	if err != nil {
		return []byte{}, err
	}

	if err := session.Start(cmd); err != nil {
		return []byte{}, err
	}

	g := errgroup.Group{}
	g.Go(func() error {
		if err := sendSudoPassword(c.password, c.host, in, stderr); err != nil {
			return err
		}
		return nil
	})

	tee := io.TeeReader(stdout, &buf)
	scanner := bufio.NewScanner(tee)

	for scanner.Scan() {
		str := scanner.Text()
		if str != "" {
			klog.V(8).Infof("[%s] [remote-stdout] %s", c.host, str)
		}
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return buf.Bytes(), session.Wait()
}

func (c *Client) SendFile(dstFile, srcFile string) error {
	sc, err := sftp.NewClient(c.client)
	if err != nil {
		return fmt.Errorf("unable to start sftp subsytem: %v", err)
	}
	defer c.Close()

	w, err := sc.Create(dstFile)
	if err != nil {
		return err
	}
	defer w.Close()

	f, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(w, f)
	if err != nil {
		return err
	}

	return nil
}

func sendSudoPassword(password, host string, in io.WriteCloser, out io.Reader) error {
	var line string

	r := bufio.NewReader(out)
	for {
		b, err := r.ReadByte()
		if err != nil {
			break
		}
		if b == byte('\n') {
			if line != "" {
				klog.V(7).Infof("[%s] [remote-stderr] %s", host, line)
			}
			line = ""
			continue
		}

		line += string(b)
		if line != "" && strings.HasPrefix(line, "[sudo] password for ") && strings.HasSuffix(line, ": ") {
			_, err = in.Write([]byte(password + "\n"))
			if err != nil {
				return err
			}
			line = ""
			klog.V(6).Infof("[%s] [sudo] send the sudo password to remote host", host)
		}
	}
	return nil
}

func (c *Client) cmdPrefix(cmd string) string {
	r := strings.NewReplacer("$", "\\$", "\"", "\\\"")
	cmd = r.Replace(cmd)

	if c.user == "root" {
		cmd = fmt.Sprintf("bash -c \"\nset -e\n%s\"", cmd)
	} else {
		cmd = fmt.Sprintf("sudo -S bash -c \"\nset -e\n%s\"", cmd)
	}

	return cmd
}
