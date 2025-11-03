package flashback

import (
	"fmt"
	"github.com/pkg/sftp"
	gossh "golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"net"
	"os"
)

type Client struct {
	Username   string
	Password   string
	Socket     string
	client     *gossh.Client
	session    *gossh.Session
	LastResult string
}

func (c *Client) Connect() (*Client, error) {
	config := &gossh.ClientConfig{}
	config.SetDefaults()
	config.User = c.Username
	config.Auth = []gossh.AuthMethod{gossh.Password(c.Password)}
	config.HostKeyCallback = func(hostname string, remote net.Addr, key gossh.PublicKey) error { return nil }
	client, err := gossh.Dial("tcp", c.Socket, config)
	if nil != err {
		return c, err
	}
	c.client = client
	return c, nil
}

func (c *Client) UploadFile(localFile string, remoteFile string, client *gossh.Client) {
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		log.Fatal(err)
		return
	}
	// 用来测试的本地文件路径 和 远程机器上的文件夹
	srcFile, err := os.Open(localFile)
	if err != nil {
		log.Fatal(err)
	}
	defer srcFile.Close()

	dstFile, err := sftpClient.Create(remoteFile)
	if err != nil {
		log.Fatal(err)
	}
	defer dstFile.Close()

	all, err := ioutil.ReadAll(srcFile)
	if err != nil {
		log.Fatal(err)
	}
	dstFile.Write(all)
	fmt.Println("upload: copy file to remote server finished!")
}

func (c Client) Run(shell string) (string, error) {
	if c.client == nil {
		if _, err := c.Connect(); err != nil {
			return "", err
		}
	}
	session, err := c.client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	buf, err := session.CombinedOutput(shell)

	c.LastResult = string(buf)
	return c.LastResult, err
}

func (c Client) CreateSession() (*gossh.Session, error) {
	if c.client == nil {
		if _, err := c.Connect(); err != nil {
			return nil, err
		}
	}
	session, err := c.client.NewSession()
	if err != nil {
		return nil, err
	}
	return session, err
}

func (c Client) CloseSession(session *gossh.Session) (string, error) {
	err := session.Close()
	return "", err
}

func (c Client) RunSession(session *gossh.Session, shell string) (string, error) {
	buf, err := session.CombinedOutput(shell)
	c.LastResult = string(buf)
	return c.LastResult, err
}
