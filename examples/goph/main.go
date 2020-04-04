package main

import (
	"os"
	"fmt"
	"flag"
	"bufio"
	"strings"
	"github.com/melbahja/goph"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	err     error
	auth    goph.Auth
	client  *goph.Client
	addr    string
	user    string
	port    int
	key     string
	cmd     string
	pass    bool
	keypass bool
)

func init() {

	flag.StringVar(&addr, "ip", "127.0.0.1", "machine ip address.")
	flag.StringVar(&user, "user", "root", "ssh user.")
	flag.IntVar(&port, "port", 22, "ssh port number.")
	flag.StringVar(&key, "key", strings.Join([]string{os.Getenv("HOME"), ".ssh", "id_rsa"}, "/"), "private key path.")
	flag.StringVar(&cmd, "cmd", "ls", "command to run.")
	flag.BoolVar(&pass, "pass", false, "ask for ssh password instead of private key.")
	flag.BoolVar(&keypass, "keypass", false, "ask for private key passphrase.")
}

func main() {

	flag.Parse()

	if pass == true {

		auth = goph.Password(askPass("Enter SSH Password: "))

	} else {

		auth = goph.Key(key, passphrase(keypass))
	}

	client, err = goph.NewUnknown(user, addr, auth)

	if err != nil {
		panic(err)
	}

	playWithSSHJustForTestingThisProgram(client)
}

func askPass(msg string) string {

	fmt.Print(msg)

	pass, err := terminal.ReadPassword(0)

	if err != nil {
		panic(err)
	}

	fmt.Println("")

	return strings.TrimSpace(string(pass))
}

func passphrase(ask bool) string {

	if ask {

		return askPass("Enter Private Key Passphrase: ")
	}

	return ""
}

func playWithSSHJustForTestingThisProgram(client *goph.Client) {

	fmt.Println("Welcome To Goph :D")
	fmt.Printf("Connected to %s\n", client.Addr)
	fmt.Println("Type your shell command and enter.")
	fmt.Println("To download file from remote type: download remote/path local/path")
	fmt.Println("To upload file to remote type: upload local/path remote/path")
	fmt.Println("To exit type: exit")

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("> ")

	var (
		out   []byte
		err   error
		cmd   string
		parts []string
	)

loop:
	for scanner.Scan() {

		err = nil
		cmd = scanner.Text()
		parts = strings.Split(cmd, " ")

		if len(parts) < 1 {
			continue
		}

		switch parts[0] {

		case "exit":
			fmt.Println("goph bye!")
			break loop

		case "download":

			if len(parts) != 3 {
				fmt.Println("please type valid download command!")
				continue loop
			}

			err = client.Download(parts[1], parts[2])

			fmt.Println("download err: ", err)
			break

		case "upload":

			if len(parts) != 3 {
				fmt.Println("please type valid download command!")
				continue loop
			}

			err = client.Upload(parts[1], parts[2])

			fmt.Println("upload err: ", err)
			break

		default:

			out, err = client.Run(cmd)
			fmt.Println(string(out), err)
		}

		fmt.Print("> ")
	}
}
