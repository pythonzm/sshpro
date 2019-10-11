package root

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"net"
	"regexp"
	"strings"
	"sync"
	"time"
)

type Connection struct {
	Host     string
	Port     string
	Username string
	Password string
	Key      string
	//cipherList                    []string
}

func (c Connection) SSHConnect() (*ssh.Session, error) {
	sshClient, err := c.connect()
	if err != nil {
		return nil, err
	}

	session, err := sshClient.NewSession()
	if err != nil {
		return nil, err
	}

	// 执行sudo命令会用到虚拟终端
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	if err := session.RequestPty("xterm", 80, 160, modes); err != nil {
		return nil, err
	}

	return session, nil
}

func (c Connection) SFTPConnect() (*sftp.Client, error) {
	sshClient, err := c.connect()
	if err != nil {
		return nil, err
	}

	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		return nil, err
	}
	return sftpClient, nil
}

func (c Connection) OutPut(command string, wg *sync.WaitGroup, ch chan bool) (output string, err error) {
	r := Result{Connection: &c}
	defer wg.Done()
	defer func() { <-ch }()
	ch <- true
	session, e := c.SSHConnect()

	if e != nil {
		r = r.GenResult(e)
		err = errors.New(r.ColorResult())
		return
	}
	defer func() {
		if e := session.Close(); e != nil {
			return
		}
	}()

	var stdout bytes.Buffer
	session.Stdout = &stdout

	in, _ := session.StdinPipe()

	go func(in io.Writer, output *bytes.Buffer) {
		defer output.Reset()

		for {
			if strings.Contains(string(output.Bytes()), "[sudo] ") {
				_, err := in.Write([]byte(c.Password + "\n"))
				if err != nil {
					break
				}
				return
			}
		}
	}(in, &stdout)

	if e := session.Run(command); e != nil {
		r.Status = r.ResultStatus(1)
		if stdout.String() != "" {
			r.Msg = stdout.String()
		} else {
			r.Msg = e.Error()
		}
		output = r.ColorResult()
		return
	}

	r.Status = r.ResultStatus(0)
	r.Msg = removeSudoString(stdout.String())
	output = r.ColorResult()
	return
}

func (c Connection) connect() (sshClient *ssh.Client, err error) {
	var (
		auth         []ssh.AuthMethod
		addr         string
		clientConfig *ssh.ClientConfig
		config       ssh.Config
	)
	auth = make([]ssh.AuthMethod, 0)
	if c.Key == "" {
		auth = append(auth, ssh.Password(c.Password))
	} else {
		pemBytes, e := ioutil.ReadFile(c.Key)
		if e != nil {
			return nil, e
		}
		var signer ssh.Signer
		if c.Password == "" {
			signer, err = ssh.ParsePrivateKey(pemBytes)
		} else {
			signer, err = ssh.ParsePrivateKeyWithPassphrase(pemBytes, []byte(c.Password))
		}
		if err != nil {
			return
		}
		auth = append(auth, ssh.PublicKeys(signer))
	}

	//if len(c.cipherList) == 0 {
	//	config = ssh.Config{
	//		Ciphers: []string{"aes128-ctr", "aes192-ctr", "aes256-ctr", "aes128-gcm@openssh.com", "arcfour256", "arcfour128", "aes128-cbc", "3des-cbc", "aes192-cbc", "aes256-cbc"},
	//	}
	//} else {
	//	config = ssh.Config{Ciphers: c.cipherList}
	//}

	clientConfig = &ssh.ClientConfig{
		User:    c.Username,
		Auth:    auth,
		Timeout: 5 * time.Second,
		Config:  config,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	addr = fmt.Sprintf("%s:%s", c.Host, c.Port)

	if sshClient, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		return
	}
	return sshClient, nil
}

func removeSudoString(s string) string {
	compile := regexp.MustCompile(`^\[sudo(.*)`)
	findString := compile.FindString(s)
	a := strings.Replace(s, findString, "", 1)
	return a
}
