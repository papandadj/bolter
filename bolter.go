package main

import (
	"embed"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v2"
)

var conf Config
var confPath string
var remoteHost string

func parseFlag() {
	if len(os.Args) == 2 {
		remoteHost = os.Args[1]
		if strings.HasPrefix(remoteHost, "-") {
			helper()
			os.Exit(0)
		}
	} else {
		helper()
		os.Exit(0)
	}
}

func main() {
	processConfigPath()
	println("config path is ", confPath)
	parseFlag()
	println("select host is ", remoteHost)

	parseConfig(confPath)

	var remoteConf Remote
	for _, c := range conf.Remote {
		if c.Host == remoteHost {
			remoteConf = c
		}
	}

	if remoteConf.Host == "" {
		printErr(errors.New("not found host"))
	}

	client, err := newSessionWithPassword(remoteConf.User, remoteConf.Address, remoteConf.Password)
	if err != nil {
		printErr(err)
	}
	defer client.Close()

	err = scp(client, remoteConf.SystemInfo, remoteConf.AgentName)
	if err != nil {
		printErr(err)
	}

	err = callAgent(client, remoteConf.AgentName, remoteConf.FilePath)
	if err != nil {
		printErr(err)
	}

}

func processConfigPath() {
	confPath = os.Getenv("bolter_config")
	var dirname string
	var err error
	if confPath != "" {
		goto passConfig
	}
	if _, err := os.Stat("./config.yaml"); err == nil {
		confPath = "./config.yaml"
		goto passConfig
	}

	dirname, err = os.UserHomeDir()
	if err != nil {
		printErr(err)
		return
	}

	confPath = fmt.Sprintf("%s/.config/bolter.yaml", dirname)
	if _, err := os.Stat(confPath); err == nil {
		goto passConfig
	}

passConfig:
	if confPath == "" {
		panic("config path not found ")
	}
}

func parseConfig(fpath string) {
	bs, err := ioutil.ReadFile(fpath)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(bs, &conf)
	if err != nil {
		panic(err)
	}
}

//Config .
type Config struct {
	Remote []Remote `yaml:"remote"`
}

type Remote struct {
	Host       string `yaml:"host"`
	Address    string `yaml:"address"`
	User       string `yaml:"user"`
	Password   string `yaml:"password"`
	FilePath   string `yaml:"filePath"`
	SystemInfo string `yaml:"systemInfo"`
	AgentName  string `yaml:"agentName"`
}

func printErr(a ...interface{}) {
	fmt.Print("err: ")
	fmt.Println(a...)
	os.Exit(0)
}

func helper() {
	println("please visit https://github.com/papandadj/bolter/tree/master")
}

//go:embed build
var f embed.FS

const defaultRemoteAgentPath = "/tmp/"

func newSessionWithPassword(user, host, pass string) (*ssh.Client, error) {
	if pass == "" {
		fmt.Print("Password: ")
		fmt.Scanf("%s\n", &pass)
	}

	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.Password(pass)},
	}
	sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()

	client, err := ssh.Dial("tcp", host, sshConfig)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func scp(client *ssh.Client, localfile, remoteFile string) error {
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	// Once a Session is created, you can execute a single command on
	// the remote side using the Run method.

	file, err := f.Open(fmt.Sprintf("build/%s", localfile))
	if err != nil {
		return err
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		return err
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		hostIn, _ := session.StdinPipe()
		defer hostIn.Close()
		fmt.Fprintf(hostIn, "C0775 %d %s\n", stat.Size(), remoteFile)
		io.Copy(hostIn, file)
		fmt.Fprint(hostIn, "\x00")
		wg.Done()
	}()
	cmd := fmt.Sprintf("scp -tr %s", defaultRemoteAgentPath)
	err = session.Run(cmd)
	wg.Wait()
	return err
}

func callAgent(client *ssh.Client, agentName, filePath string) error {
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin
	err = session.Run(fmt.Sprintf("%s %s", fmt.Sprintf("%s%s", defaultRemoteAgentPath, agentName), filePath))
	if err != nil {
		return err
	}
	return nil
}
