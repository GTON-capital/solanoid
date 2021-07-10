package mvp

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"

	"github.com/urfave/cli/v2"
)

type MVPInputConfig struct {
	SolanaRecipient string
}

func RunMVP() {
	mvpCfg := &MVPInputConfig{
		SolanaRecipient: "98tnzivwLxfb7ThDegaZciHF6Dzk99q8Fr9F5ZksVcgN",
	}

	app := &cli.App{
		Name:  "Polygon -> Solana MVP",
		Usage: "This App shows how fast cross-chain swaps occur (SuSy Wrapped $GTON)",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "sol-recipient",
				Value:       "98tnzivwLxfb7ThDegaZciHF6Dzk99q8Fr9F5ZksVcgN",
				Usage:       "Solana $GTON Recipient",
				Destination: &mvpCfg.SolanaRecipient,
			},
		},
		Action: func(c *cli.Context) error {
			err := ProcessMVP_PolygonSolana()
			fmt.Printf("Error occured during MVP: %v \n", err)
			debug.PrintStack()
			return err
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
