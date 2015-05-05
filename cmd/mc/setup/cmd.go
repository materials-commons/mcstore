package setup

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/materials-commons/mcstore/cmd/pkg/user"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db/schema"
	"github.com/parnurzeal/gorequest"
	"gnd.la/net/urlutil"
	"golang.org/x/crypto/ssh/terminal"
)

var Command = cli.Command{
	Name:   "setup",
	Usage:  "Set up the configuration",
	Action: setupCLI,
}

type userConfigSetup struct {
	APIKey string `json:"apikey"`
}

type userLogin struct {
	Password string `json:"password"`
}

func setupCLI(c *cli.Context) {
	fmt.Println("Setting up configuration...")
	username, password := getUsernameAndPassword()
	apikey, err := getAPIKey(username, password)
	if err != nil {
		return
	}
	configSetup := userConfigSetup{
		APIKey: apikey,
	}
	writeConfigFile(configSetup)
	fmt.Println("Done.")
}

func getUsernameAndPassword() (username, password string) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Your MaterialsCommons Username: ")
	username, _ = reader.ReadString('\n')
	username = strings.TrimSpace(username)
	fmt.Print("Your MaterialsCommons Password: ")
	pw, _ := terminal.ReadPassword(0)
	fmt.Println("Contacting MaterialsCommons...")
	return username, string(pw)
}

func getAPIKey(username, password string) (string, error) {
	var u schema.User
	l := userLogin{
		Password: password,
	}
	request := gorequest.New().TLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	resp, body, errs := request.Put(urlutil.MustJoin(app.MCApi.MCUrl(), path.Join("api", "user", username, "apikey"))).
		Send(l).
		End()
	if len(errs) != 0 {
		fmt.Printf("Unable to communicate with MaterialsCommons at: %s\n", app.MCApi.MCUrl())
		return "", app.ErrInvalid
	}
	if resp.StatusCode > 299 {
		fmt.Printf("Error communicating with MaterialsCommons: %s\n", resp.Status)
		return "", app.ErrInvalid
	}
	json.Unmarshal([]byte(body), &u)
	return u.APIKey, nil
}

func writeConfigFile(configSetup userConfigSetup) {
	if err := os.MkdirAll(user.ConfigDir(), 0700); err != nil {
		panic(fmt.Sprintf("Couldn't create dir: %s", err))
	}
	b, err := json.Marshal(configSetup)
	if err != nil {
		panic(fmt.Sprintf("Can't marshal: %s", err))
	}
	ioutil.WriteFile(user.ConfigFile(), b, 0700)
}
