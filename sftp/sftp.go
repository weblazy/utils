package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"
	"regexp"
	"time"

	"github.com/pkg/sftp"
	"github.com/sunmi-OS/gocore/v2/utils"
	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/ssh"
)

func main() {
	// 打印banner
	utils.PrintBanner("sftp")
	// 配置cli参数
	app := cli.NewApp()
	app.Name = "sftp"
	app.Usage = "sftp"
	app.Version = "v1.0.0"
	// 指定命令运行的函数
	app.Commands = []*cli.Command{
		Sftp,
	}
	// 启动cli
	if err := app.Run(os.Args); err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}
}

// Sftp 创建配置文件
var Sftp = &cli.Command{
	Name: "sftp",
	Subcommands: []*cli.Command{
		{
			Name: "push",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "src",
					Usage: "path",
				},
				&cli.StringFlag{
					Name:  "dst",
					Usage: "user:password@host/path",
				},
			},
			Usage:  "sftp push --src [src path] --dst [dst path]",
			Action: push,
		},
		{
			Name: "pull",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "src",
					Usage: "user:password@host/path",
				},
				&cli.StringFlag{
					Name:  "dst",
					Usage: "path",
				},
			},
			Usage:  "sftp pull --src [src path] --dst [dst path]",
			Action: pull,
		},
	},
}

func connect(user, password, host string, port int) (*sftp.Client, error) {
	var (
		auth         []ssh.AuthMethod
		addr         string
		clientConfig *ssh.ClientConfig
		sshClient    *ssh.Client
		sftpClient   *sftp.Client
		err          error
	)
	// get auth method
	auth = make([]ssh.AuthMethod, 0)
	auth = append(auth, ssh.Password(password))

	clientConfig = &ssh.ClientConfig{
		User:    user,
		Auth:    auth,
		Timeout: 30 * time.Second,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	// connet to ssh
	addr = fmt.Sprintf("%s:%d", host, port)

	if sshClient, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		return nil, err
	}

	// create sftp client
	if sftpClient, err = sftp.NewClient(sshClient); err != nil {
		return nil, err
	}

	return sftpClient, nil
}

// push 文件上传远程服务器
func push(c *cli.Context) error {
	src := c.String("src")
	dst := c.String("dst")
	reg := regexp.MustCompile("^(.+?):(.+)@(.+?)(/.*)")
	res := reg.FindStringSubmatch(dst)
	if len(res) != 5 {
		return fmt.Errorf("参数错误")
	}

	// 这里换成实际的 SSH 连接的 用户名，密码，主机名或IP，SSH端口
	sftpClient, err := connect(res[1], res[2], res[3], 22)
	if err != nil {
		log.Fatal(err)
	}
	defer sftpClient.Close()

	ok, err := IsDir(src)
	if err != nil {
		return err
	}
	if ok {
		_ = sftpClient.Mkdir(res[4])
		uploadDirectory(sftpClient, src, res[4])
	} else {
		uploadFile(sftpClient, src, res[4])
	}
	return nil
}

// 判断所给路径是否为文件夹
func IsDir(path string) (bool, error) {
	s, err := os.Stat(path)
	if err != nil {
		return false, fmt.Errorf("path not found")
	}
	return s.IsDir(), nil
}

// 判断所给路径是否为文件
func IsFile(path string) (bool, error) {
	s, err := os.Stat(path)
	if err != nil {
		return false, fmt.Errorf("path not found")
	}
	return !s.IsDir(), nil
}

func uploadDirectory(sftpClient *sftp.Client, localPath string, remotePath string) {
	localFiles, err := ioutil.ReadDir(localPath)
	if err != nil {
		log.Fatal("read dir list fail ", err)
	}
	for _, backupDir := range localFiles {

		localFilePath := path.Join(localPath, backupDir.Name())
		remoteFilePath := path.Join(remotePath, backupDir.Name())
		if backupDir.IsDir() {
			_ = sftpClient.Mkdir(remoteFilePath)
			uploadDirectory(sftpClient, localFilePath, remoteFilePath)
		} else {
			uploadFile(sftpClient, path.Join(localPath, backupDir.Name()), remotePath)
		}
	}
	fmt.Println(localPath + " copy directory finished!")
}

func uploadFile(sftpClient *sftp.Client, localFilePath string, remotePath string) {
	srcFile, err := os.Open(localFilePath)
	if err != nil {
		fmt.Println("os.Open error : ", localFilePath)
		log.Fatal(err)
	}
	defer srcFile.Close()
	var remoteFileName = path.Base(localFilePath)
	dstFile, err := sftpClient.Create(path.Join(remotePath, remoteFileName))
	if err != nil {
		fmt.Println("sftpClient.Create error : ", path.Join(remotePath, remoteFileName))
		log.Fatal(err)
	}
	defer dstFile.Close()
	ff, err := ioutil.ReadAll(srcFile)
	if err != nil {
		fmt.Println("ReadAll error : ", localFilePath)
		log.Fatal(err)
	}
	dstFile.Write(ff)
	fmt.Println(localFilePath + " copy file finished!")
}

// pull 从传远程服务器拉取文件
func pull(c *cli.Context) error {
	src := c.String("src")
	dst := c.String("dst")
	reg := regexp.MustCompile("^(.+?):(.+)@(.+?)(/.*)")
	res := reg.FindStringSubmatch(src)
	fmt.Printf("%#v\n", res)
	if len(res) != 5 {
		return fmt.Errorf("参数错误")
	}

	// 这里换成实际的 SSH 连接的 用户名，密码，主机名或IP，SSH端口
	sftpClient, err := connect(res[1], res[2], res[3], 22)
	if err != nil {
		log.Fatal(err)
	}
	defer sftpClient.Close()

	ok, err := IsRemoteDir(sftpClient, res[4])
	if err != nil {
		return err
	}
	if ok {
		os.MkdirAll(dst, os.ModePerm)
		pullDirectory(sftpClient, dst, res[4])
	} else {
		pullFile(sftpClient, dst, res[4])
	}
	return nil
}

// 判断所给路径是否为文件夹
func IsRemoteDir(client *sftp.Client, path string) (bool, error) {
	s, err := client.Stat(path)
	if err != nil {
		return false, fmt.Errorf("path not found")
	}
	return s.IsDir(), nil
}

// 判断所给路径是否为文件
func IsRemoteFile(client *sftp.Client, path string) (bool, error) {
	s, err := client.Stat(path)
	if err != nil {
		return false, fmt.Errorf("path not found")
	}
	return !s.IsDir(), nil
}

func pullDirectory(sftpClient *sftp.Client, localPath string, remotePath string) {
	remoteFiles, err := sftpClient.ReadDir(remotePath)
	if err != nil {
		log.Fatal("read dir list fail ", err)
	}
	for _, backupDir := range remoteFiles {
		localFilePath := path.Join(localPath, backupDir.Name())
		remoteFilePath := path.Join(remotePath, backupDir.Name())
		if backupDir.IsDir() {
			_ = os.Mkdir(localFilePath, os.ModePerm)
			pullDirectory(sftpClient, localFilePath, remoteFilePath)
		} else {
			pullFile(sftpClient, localPath, path.Join(remotePath, backupDir.Name()))
		}
	}
	fmt.Println(localPath + " copy directory finished!")
}

func pullFile(sftpClient *sftp.Client, localFilePath string, remotePath string) {
	srcFile, err := sftpClient.Open(remotePath)
	if err != nil {
		fmt.Println("os.Open error : ", remotePath)
		log.Fatal(err)
	}
	defer srcFile.Close()
	var remoteFileName = path.Base(remotePath)
	dstFile, err := os.Create(path.Join(localFilePath, remoteFileName))
	if err != nil {
		fmt.Println("sftpClient.Create error : ", path.Join(localFilePath, remoteFileName))
		log.Fatal(err)
	}
	defer dstFile.Close()
	ff, err := ioutil.ReadAll(srcFile)
	if err != nil {
		fmt.Println("ReadAll error : ", localFilePath)
		log.Fatal(err)
	}
	dstFile.Write(ff)
	fmt.Println(localFilePath + " copy file finished!")
}
