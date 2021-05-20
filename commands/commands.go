package commands

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	config  string //
	daemon  bool   //
	version bool   //

	// HoarderCmd ...go tidy
	SolanoidCmd = &cobra.Command{
		Use:   "solanoid",
		Short: "",
		Long:  ``,

		// parse the config if one is provided, or use the defaults. Set the backend
		// driver to be used
		PersistentPreRun: func(ccmd *cobra.Command, args []string) {
			l, _ := zap.NewDevelopment()
			zap.ReplaceGlobals(l)
			// if --config is passed, attempt to parse the config file
			if config != "" {

				// get the filepath
				abs, err := filepath.Abs(config)
				if err != nil {
					zap.L().Sugar().Errorf("Error reading filepath: %f", err.Error())
				}

				// get the config name
				base := filepath.Base(abs)

				// get the path
				path := filepath.Dir(abs)

				//
				viper.SetConfigName(strings.Split(base, ".")[0])
				viper.AddConfigPath(path)

				// Find and read the config file; Handle errors reading the config file
				if err := viper.ReadInConfig(); err != nil {
					zap.L().Sugar().Fatal("Failed to read config file: ", err.Error())
					os.Exit(1)
				}
			}
		},

		// either run hoarder as a server, or run it as a CLI depending on what flags
		// are provided
		Run: func(ccmd *cobra.Command, args []string) {
			// fall back on default help if no args/flags are passed
			ccmd.HelpFunc()(ccmd, args)
		},
	}
)

func init() {
	SolanoidCmd.PersistentFlags().String("log-level", "INFO", "Output level of logs (TRACE, DEBUG, INFO, WARN, ERROR, FATAL)")
	viper.BindPFlag("log-level", SolanoidCmd.PersistentFlags().Lookup("log-level"))

}

func waitTx(tx string, host string) {
	u := url.URL{Scheme: "ws", Host: host, Path: "/"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()
	req := `
	{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "signatureSubscribe",
		"params": [
		  "%s",
		  {
			"commitment": "finalized"
		  }
		]
	  }
	`
	unsubscribeRequest := "{\"jsonrpc\":\"2.0\", \"id\":1, \"method\":\"signatureUnsubscribe\", \"params\":[%d]}"
	done := make(chan struct{})
	defer close(done)
	go func() {

		subscription := 0
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
			a := struct {
				Method       string `json:"method"`
				Result       int    `json:"result"`
				Subscription int    `json:"subscription"`
			}{}
			json.Unmarshal(message, &a)
			switch a.Method {
			case "":
				subscription = a.Subscription
			case "signatureNotification":
				c.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(unsubscribeRequest, subscription)))
				//time.Sleep(time.Millisecond * 100)
				c.Close()
				done <- struct{}{}
				return
			}

		}

	}()

	err = c.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(req, tx)))
	<-done
	//time.Sleep(time.Millisecond * 100)
}
