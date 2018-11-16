package cmd

import (
	"fmt"
	"log"
	"path"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vigliag/isamuni-go/model"
)

func getModel() *model.Model {
	dbname := path.Join(viper.GetString("data"), "database.db")
	return model.New(model.Connect(dbname))
}

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage users",
}

var userVerificationMail = &cobra.Command{
	Use:   "sendverification [email]",
	Short: "Register a users",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		m := getModel()
		defer m.Close()

		email := args[0]
		ctl := GetController()

		var u model.User
		err := m.Db.First(&u, "email=?", email).Error
		if err != nil {
			fmt.Println("Can't find user with that mail")
			return
		}

		err = ctl.SendEmailVerification(&u)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("done")
	},
}

var userRegisterCmd = &cobra.Command{
	Use:   "register [name] [email]",
	Short: "Register a users",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		m := getModel()
		defer m.Close()

		user := args[0]
		email := args[1]

		fmt.Printf("Insert password for user %s with email %s: ", user, email)
		password, err := terminal.ReadPassword(int(syscall.Stdin))
		fmt.Println()

		if !strings.Contains(email, "@") {
			fmt.Println("the second argument must be an email")
			return
		}

		if err != nil || len(password) < 4 {
			fmt.Println("unable to read password, aborting")
			return
		}

		_, err = m.RegisterEmail(user, email, string(password), "user")
		if err != nil {
			fmt.Println("Unable to register user because of error: ", err)
		}

		fmt.Println("done")
	},
}

func init() {
	userCmd.AddCommand(userVerificationMail)
	userCmd.AddCommand(userRegisterCmd)
	rootCmd.AddCommand(userCmd)
}
