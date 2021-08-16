package ws

type LogsSubscribeNotification struct {
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  struct {
		Result struct {
			Context struct {
				Slot int `json:"slot"`
			} `json:"context"`
			Value struct {
				Signature string      `json:"signature"`
				Err       interface{} `json:"err"`
				Logs      []string    `json:"logs"`
			} `json:"value"`
		} `json:"result"`
		Subscription int `json:"subscription"`
	} `json:"params"`
}

type LogsSubscribeParam struct {
	Mentions   []string `json:"mentions,omitempty"`
	Commitment string   `json:"commitment,omitempty"`
}
type LogsSubscribeBody struct {
	Jsonrpc string               `json:"jsonrpc"`
	ID      int                  `json:"id"`
	Method  string               `json:"method"`
	Params  []LogsSubscribeParam `json:"params"`
}
type AccountNotification struct {
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  struct {
		Result struct {
			Context struct {
				Slot int `json:"slot"`
			} `json:"context"`
			Value struct {
				Data       []string `json:"data"`
				Executable bool     `json:"executable"`
				Lamports   int      `json:"lamports"`
				Owner      string   `json:"owner"`
				RentEpoch  int      `json:"rentEpoch"`
			} `json:"value"`
		} `json:"result"`
		Subscription int `json:"subscription"`
	} `json:"params"`
}

type RequestBody struct {
	Jsonrpc string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}
type Encoding struct {
	Encoding   string `json:"encoding"`
	Commitment string `json:"commitment"`
}
