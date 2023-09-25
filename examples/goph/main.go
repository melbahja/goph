package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	osuser "os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"github.com/tareksalem/goph"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

//
// Run command and auth via password:
// > go run main.go --ip 192.168.122.102 --pass --cmd ls
//
// Run command and auth via private key:
// > go run main.go --ip 192.168.122.102 --cmd ls
// Or:
// > go run main.go --ip 192.168.122.102 --key /path/to/private_key --cmd ls
//
// Run command and auth with private key and passphrase:
// > go run main.go --ip 192.168.122.102 --passphrase --cmd ls
//
// Run a command and interrupt it after 1 second:
// > go run main.go --ip 192.168.122.102 --cmd "sleep 10" --timeout=1s
//
// You can test with the interactive mode without passing --cmd flag.
//

var (
	err        error
	auth       goph.Auth
	client     *goph.Client
	addr       string
	user       string
	port       uint
	key        string
	cmd        string
	pass       bool
	passphrase bool
	timeout    time.Duration
	agent      bool
	sftpc      *sftp.Client
)

func init() {
	usr, err := osuser.Current()
	if err != nil {
		fmt.Println("couldn't determine current user.  defaulting to 'root'")
		usr.Username = "root"
	}

	flag.StringVar(&addr, "ip", "127.0.0.1", "machine ip address.")
	flag.StringVar(&user, "user", usr.Username, "ssh user.")
	flag.UintVar(&port, "port", 22, "ssh port number.")
	flag.StringVar(&key, "key", filepath.Join(os.Getenv("HOME"), ".ssh", "id_rsa"), "private key path.")
	flag.StringVar(&cmd, "cmd", "", "command to run.")
	flag.BoolVar(&pass, "pass", false, "ask for ssh password instead of private key.")
	flag.BoolVar(&agent, "agent", false, "use ssh agent for authentication (unix systems only).")
	flag.BoolVar(&passphrase, "passphrase", false, "ask for private key passphrase.")
	flag.DurationVar(&timeout, "timeout", 0, "interrupt a command with SIGINT after a given timeout (0 means no timeout)")
}

func VerifyHost(host string, remote net.Addr, key ssh.PublicKey) error {

	//
	// If you want to connect to new hosts.
	// here your should check new connections public keys
	// if the key not trusted you shuld return an error
	//

	// hostFound: is host in known hosts file.
	// err: error if key not in known hosts file OR host in known hosts file but key changed!
	hostFound, err := goph.CheckKnownHost(host, remote, key, "")

	// Host in known hosts but key mismatch!
	// Maybe because of MAN IN THE MIDDLE ATTACK!
	if hostFound && err != nil {

		return err
	}

	// handshake because public key already exists.
	if hostFound && err == nil {

		return nil
	}

	// Ask user to check if he trust the host public key.
	if askIsHostTrusted(host, key) == false {

		// Make sure to return error on non trusted keys.
		return errors.New("you typed no, aborted!")
	}

	// Add the new host to known hosts file.
	return goph.AddKnownHost(host, remote, key, "")
}

func main() {

	flag.Parse()

	var err error

	if agent || goph.HasAgent() {

		auth, err = goph.UseAgent()

	} else if pass {

		auth = goph.Password(askPass("Enter SSH Password: "))

	} else {

		auth, err = goph.Key(key, getPassphrase(passphrase))
	}

	if err != nil {
		panic(err)
	}

	client, err = goph.NewConn(&goph.Config{
		User:     user,
		Addr:     addr,
		Port:     port,
		Auth:     auth,
		Callback: VerifyHost,
	})

	if err != nil {
		panic(err)
	}

	// Close client net connection
	defer client.Close()

	// If the cmd flag exists
	if cmd != "" {
		ctx := context.Background()
		// create a context with timeout, if supplied in the argumetns
		if timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}

		out, err := client.RunContext(ctx, cmd)

		fmt.Println(string(out), err)
		return
	}

	// else open interactive mode.
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

func getPassphrase(ask bool) string {

	if ask {

		return askPass("Enter Private Key Passphrase: ")
	}

	return ""
}

func askIsHostTrusted(host string, key ssh.PublicKey) bool {

	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("Unknown Host: %s \nFingerprint: %s \n", host, ssh.FingerprintSHA256(key))
	fmt.Print("Would you like to add it? type yes or no: ")

	a, err := reader.ReadString('\n')

	if err != nil {
		log.Fatal(err)
	}

	return strings.ToLower(strings.TrimSpace(a)) == "yes"
}

func getSftp(client *goph.Client) *sftp.Client {

	var err error

	if sftpc == nil {

		sftpc, err = client.NewSftp()

		if err != nil {
			panic(err)
		}
	}

	return sftpc
}

func playWithSSHJustForTestingThisProgram(client *goph.Client) {

	fmt.Println("Welcome To Goph :D")
	fmt.Printf("Connected to %s\n", client.Config.Addr)
	fmt.Println("Type your shell command and enter.")
	fmt.Println("To download file from remote type: download remote/path local/path")
	fmt.Println("To upload file to remote type: upload local/path remote/path")
	fmt.Println("To create a remote dir type: mkdirall /path/to/remote/newdir")
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
				fmt.Println("please type valid upload command!")
				continue loop
			}

			err = client.Upload(parts[1], parts[2])

			fmt.Println("upload err: ", err)
			break

		case "mkdirall":

			if len(parts) != 2 {
				fmt.Println("please type valid mkdirall command!")
				continue loop
			}

			ftp := getSftp(client)

			err = ftp.MkdirAll(parts[1])
			fmt.Printf("mkdirall err(%v) you can check via: stat %s\n", err, parts[1])

		default:

			command, err := client.Command(parts[0], parts[1:]...)
			if err != nil {
				panic(err)
			}
			out, err = command.CombinedOutput()
			fmt.Println(string(out), err)
		}

		fmt.Print("> ")
	}
}
