package cmd

import (
	"errors"
	"fmt"
	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"
	"main/src/internal/data"
	"main/src/internal/repository"
	"main/src/service"
	"os"
	"os/exec"
	"sort"
	"time"
)

func Menu() {
	go func() {
		for {
			balanceWithProfit := repository.Repository.Balance + repository.Repository.GeneralPnl
			if balanceWithProfit <= 500 && balanceWithProfit >= 0 && len(repository.Repository.Positions) > 0 {
				for _, v := range repository.Repository.Positions {
					if _, err := service.ClosePosition(v); err != nil {
						log.Error(err)
					}
				}
			}
		}
	}()
	for {
		//ClearWindow()
		fmt.Println("===========MENU===========")
		fmt.Printf("1 -> get price list\n")
		fmt.Printf("2 -> open positon \n")
		fmt.Printf("3 -> close position (pos-close <uuid-symbol>)\n")
		fmt.Printf("4 -> logout\n")
		fmt.Printf("5 -> (donate <integer>)\n")
		fmt.Printf("========BALANCE: %.2f$ [%-+f]==========\n", repository.Repository.Balance, repository.Repository.GeneralPnl)
		//consoleReader := bufio.NewReader(os.Stdin)
		//
		//inp, err := consoleReader.ReadString('\n')
		//if err != nil {
		//	log.Fatal(err)
		//}
		//inp = strings.TrimSuffix(inp, "\n")
		//inpArgs := strings.Split(inp, " ")
		var inp int

		if _, err := fmt.Scanf("%d", &inp); err != nil {
			log.Error(err)
			continue
		}
		switch inp {
		case 1:
			ClearWindow()
			go func() {
				for {
					if repository.Repository.IsUpdated {

						ClearWindow()
						fmt.Println("=========== UPDATING ===========")
						ClearWindow()
						time.Sleep(time.Second * 1)

						renderPricesTable()

						renderPositionsProfitTable()
						repository.Repository.IsUpdated = false
					}
				}
			}()

		case 2:
			fmt.Println("Input symbol name to open position")
			var symbolNameInp string
			if _, err := fmt.Scanf("%s", &symbolNameInp); err != nil {
				log.Error(err)
			}
			if _, ok := repository.Repository.Data[symbolNameInp]; !ok {
				fmt.Println("No such symbol")
				continue
			}
			fmt.Println("Choose option to open position")
			fmt.Println("1 - buy")
			fmt.Println("2 - sell")
			if _, err := fmt.Scanf("%d", &inp); err != nil {
				log.Error()
				continue
			}
			switch inp {
			case 1:
				if _, err := service.OpenPosition(repository.Repository.Data[symbolNameInp], true); err != nil {
					log.Error(err)
					continue
				}
				fmt.Printf("Failed to open position for: %v", symbolNameInp)
			case 2:
				if _, err := service.OpenPosition(repository.Repository.Data[symbolNameInp], false); err != nil {
					log.Error(err)
				}
				continue
			default:
				fmt.Printf("incorrect <buy/sell> value")
				continue

			}
		case 3:
			fmt.Println("Input position number to close")
			var posNumInp int
			if _, err := fmt.Scanf("%d", &posNumInp); err != nil {
				log.Error(err)
				continue
			}
			if posNumInp > 0 && posNumInp > len(repository.Repository.SortedPositionsKey) {
				log.Error(errors.New("Invalid position number "))
				continue
			}

			posKey := repository.Repository.SortedPositionsKey[posNumInp]

			if pos, ok := repository.Repository.SortedPositionsByPnl[posKey]; ok {
				if _, err := service.ClosePosition(pos); err != nil {
					log.Error(err)
				}
				continue
			}
			fmt.Printf("Failed to close position for: %v", repository.Repository.SortedPositionsByPnl[posKey])
			continue
		case 4:
			if err := service.Logout(); err != nil {
				fmt.Printf("Failed to log out")
				time.Sleep(time.Millisecond * 300)
				continue
			}
			return
		case 5:
			fmt.Println("Input value which you want donate")
			var donateValInp float32
			_, err := fmt.Scanf("%f", &donateValInp)
			if err != nil {
				log.Error(err)
				continue
			}
			if err = service.Donate(donateValInp); err != nil {
				log.Error(err)
				continue
			}
			log.Infof("Successful donate: %f$", donateValInp)

		default:
			continue
		}
	}
}

func renderPricesTable() {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Symbol", "Upd Time", "Ask", "Bid"})
	for _, v := range repository.Repository.Data {
		table.Append([]string{v.Symbol, fmt.Sprint(time.Unix(v.Uuid, 0).Format("15:04:05 02/01/2006")), fmt.Sprintf("%.2f$", v.Ask), fmt.Sprintf("%.2f$", v.Bid)})
	}
	table.Render()
}

func renderPositionsProfitTable() {
	fmt.Printf("Ur profits\n")

	if len(repository.Repository.Positions) == 0 {
		return
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"№", "Symbol", "Open Price", "Profit"})

	repository.Repository.SortedPositionsByPnl = make(map[float64]data.Position)
	repository.Repository.SortedPositionsKey = make([]float64, 0, len(repository.Repository.Positions))

	for _, val := range repository.Repository.Positions {
		pnl := val.PNL(repository.Repository.Data[val.Symbol])
		repository.Repository.SortedPositionsByPnl[float64(pnl)] = val
		repository.Repository.SortedPositionsKey = append(repository.Repository.SortedPositionsKey, float64(pnl))
	}

	var pnl float32
	for _, v := range repository.Repository.Positions {
		pnl += v.PNL(repository.Repository.Data[v.Symbol])
	}

	repository.Repository.GeneralPnl = pnl
	fmt.Printf("========BALANCE: %.2f$ [%-+f]==========\n", repository.Repository.Balance, repository.Repository.GeneralPnl)
	sort.Float64s(repository.Repository.SortedPositionsKey)

	for k, v := range repository.Repository.SortedPositionsKey {
		pos := repository.Repository.SortedPositionsByPnl[v]
		profit := pos.PNL(repository.Repository.Data[pos.Symbol])
		table.Append([]string{
			fmt.Sprintf("%d", k),
			fmt.Sprintf("%v-%v", pos.Symbol, pos.UUID),
			fmt.Sprintf("%.2f$", pos.Open),
			fmt.Sprintf("%.2f$", profit),
		})

	}
	table.Render()
}

func ClearWindow() {
	cmd := exec.Command("clear")

	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		log.Error(err)
	}
}
