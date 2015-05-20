package setup

import (
	"bufio"
	"bytes"
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

// Command contains the options to configure the setup command.
var Command = cli.Command{
	Name:   "setup",
	Usage:  "Set up the configuration",
	Action: setupCLI,
}

// userConfigSetup contains all the configuration entries needed for the
// mc command.
type userConfigSetup struct {
	APIKey string `json:"apikey"`
}

// userLogin contains the user password used to retrieve the users apikey.
type userLogin struct {
	Password string `json:"password"`
}

// setupCLI implements the setup cli command. Setup will initialize a users account on the
// local system so that they can use the mc cli.
func setupCLI(c *cli.Context) {
	fmt.Println("Setting up mc configuration...")
	username, password := getUsernameAndPassword()
	apikey, err := getAPIKey(username, password)
	if err != nil {
		return
	}
	configSetup := userConfigSetup{
		APIKey: apikey,
	}
	writeConfigFile(configSetup)
	fmt.Println("\nYou have successfully completed the setup.")
}

// getUsernameAndPassword prompts for the current users materials commons
// username and password.
func getUsernameAndPassword() (username, password string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("  Please enter your MaterialsCommons username: ")
	username, _ = reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Print("  Please enter your MaterialsCommons password: ")
	pw, _ := terminal.ReadPassword(0)

	return username, string(pw)
}

// getAPIKey communicates with the materials commons api to retrieve
// the users application apikey.
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

// writeConfigFile writes the created config.json file. It also creates
// the $HOME/.materialscommons directory.
func writeConfigFile(configSetup userConfigSetup) {
	if err := os.MkdirAll(user.ConfigDir(), 0700); err != nil {
		panic(fmt.Sprintf("Couldn't create dir: %s", err))
	}
	b, err := json.Marshal(configSetup)
	if err != nil {
		panic(fmt.Sprintf("Can't marshal: %s", err))
	}
	var out bytes.Buffer
	json.Indent(&out, b, "", "  ")
	ioutil.WriteFile(user.ConfigFile(), out.Bytes(), 0700)
}
