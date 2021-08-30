package main

import (
	"bufio"
	"fmt"
	log "github.com/sirupsen/logrus"
	"main/src/cmd"
	"main/src/internal/data"
	"main/src/internal/repository"
	"main/src/service"
	"os"
	"strings"
)

func main() {
	var inp string
	repository.Repository.Positions = make(map[string]data.Position)

	go func() {
		if err := service.ListenPriceUpdates(); err != nil {
			log.Fatal(err)
		}
	}()
	for {
		cmd.ClearWindow()
		fmt.Printf("You started forex client app...\n")
		fmt.Printf("1 - login\n" +
			"other - exit\n")
		if _, err := fmt.Scanf("%s", &inp); err != nil {
			continue
		}
		switch inp {
		case "1":
			fmt.Printf("Input <username> <password>")
			consoleReader := bufio.NewReader(os.Stdin)
			inpLogin, err := consoleReader.ReadString('\n')
			if err != nil {
				log.Fatal(err)
			}
			inpLogin = strings.TrimSuffix(inpLogin, "\n")
			inpArgs := strings.Split(inpLogin, " ")
			token, err := service.Login(inpArgs[0], inpArgs[1])
			if err != nil {
				log.Error(err)
				continue
			}
			repository.Repository.UserToken = token
			if err = service.GetUserData(); err != nil {
				log.Error(err)
			}
			cmd.ClearWindow()
			cmd.Menu()
		default:
			os.Exit(1)
		}
	}
}
