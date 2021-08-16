package commands

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"testing"
	"time"

	"github.com/Gravity-Tech/solanoid/commands/executor"
	"github.com/Gravity-Tech/solanoid/commands/ws"
	"github.com/Gravity-Tech/solanoid/models"
	"github.com/Gravity-Tech/solanoid/models/nebula"
	"github.com/gorilla/websocket"

	solclient "github.com/portto/solana-go-sdk/client"
	"github.com/portto/solana-go-sdk/common"
	"github.com/portto/solana-go-sdk/types"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

/*
 * Test logical steps
 *
 * 1. Deploy Nebula
 * 2. Init Nebula
 * 3. Deploy Port
 * 4. Subscribe Port to Nebula
 * 5. Call mocked attach data.
 *
 * Goals:
 * 1. Validate minting flow.
 * 2. Validate oracle multisig. (with various bft*)
 * 3. Validate double spend on attach
 * 4. Validate the atomic call: nebula.send_value_to_subs() -> nebula.attach()
 */
func TestNebulaSendValueToIBPortSubscriber(t *testing.T) {
	var err error

	deployer, err := NewOperatingAddress(t, "../private-keys/_test_deployer-pk-deployer.json", nil)
	ValidateError(t, err)

	gravityProgram, err := NewOperatingAddress(t, "../private-keys/_test_only-gravity-program.json", nil)
	ValidateError(t, err)

	nebulaProgram, err := NewOperatingAddress(t, "../private-keys/_test_only-nebula-program.json", &OperatingAddressBuilderOptions{
		WithPDASeeds: []byte(executor.CommonGravityBumpSeeds),
	})
	ValidateError(t, err)

	ibportProgram, err := NewOperatingAddress(t, "../private-keys/_test_only_ibport-program.json", &OperatingAddressBuilderOptions{
		WithPDASeeds: []byte(executor.CommonGravityBumpSeeds),
	})
	ValidateError(t, err)

	const BFT = 3

	const OmitSendValueFlow = false

	consulsList, err := GenerateConsuls(t, "../private-keys/_test_consul_prefix_", BFT)
	ValidateError(t, err)

	operatingConsul := consulsList.List[0]

	for i, consul := range append(consulsList.List, *deployer) {
		if i == BFT {
			WrappedFaucet(t, deployer.PKPath, "", 10)
		}

		WrappedFaucet(t, deployer.PKPath, consul.PublicKey.ToBase58(), 10)
	}

	waitTransactionConfirmations()

	RPCEndpoint, _ := InferSystemDefinedRPC()

	tokenDeployResult, err := CreateToken(deployer.PKPath)
	ValidateError(t, err)

	tokenProgramAddress := tokenDeployResult.Token.ToBase58()

	fmt.Printf("Token deployed: %v \n", tokenDeployResult.Signature)

	deployerTokenAccount, err := CreateTokenAccount(deployer.PKPath, tokenProgramAddress)
	ValidateError(t, err)

	waitTransactionConfirmations()

	// mint some tokens for deployer
	err = MintToken(deployer.PKPath, tokenProgramAddress, 1_000_000, deployerTokenAccount)
	ValidateError(t, err)
	t.Log("Minted some tokens")

	gravityDataAccount, err := GenerateNewAccount(deployer.PrivateKey, GravityContractAllocation, gravityProgram.PublicKey.ToBase58(), RPCEndpoint)
	ValidateError(t, err)

	gravityMultisigAccount, err := GenerateNewAccount(deployer.PrivateKey, MultisigAllocation, gravityProgram.PublicKey.ToBase58(), RPCEndpoint)
	ValidateError(t, err)

	nebulaDataAccount, err := GenerateNewAccount(deployer.PrivateKey, NebulaAllocation, nebulaProgram.PublicKey.ToBase58(), RPCEndpoint)
	ValidateError(t, err)

	nebulaMultisigAccount, err := GenerateNewAccount(deployer.PrivateKey, MultisigAllocation, nebulaProgram.PublicKey.ToBase58(), RPCEndpoint)
	ValidateError(t, err)

	ibportDataAccount, err := GenerateNewAccount(deployer.PrivateKey, IBPortAllocation, ibportProgram.PublicKey.ToBase58(), RPCEndpoint)
	ValidateError(t, err)

	ParallelExecution(
		[]func(){
			func() {
				_, err = DeploySolanaProgram(t, "ibport", ibportProgram.PKPath, consulsList.List[0].PKPath, "../binaries/ibport.so")
				ValidateError(t, err)
			},
			func() {
				_, err = DeploySolanaProgram(t, "gravity", gravityProgram.PKPath, consulsList.List[1].PKPath, "../binaries/gravity.so")
				ValidateError(t, err)
			},
			func() {
				_, err = DeploySolanaProgram(t, "nebula", nebulaProgram.PKPath, consulsList.List[2].PKPath, "../binaries/nebula.so")
				ValidateError(t, err)
			},
		},
	)

	waitTransactionConfirmations()

	err = AuthorizeToken(t, deployer.PKPath, tokenProgramAddress, "mint", ibportProgram.PDA.ToBase58())
	ValidateError(t, err)
	t.Log("Authorizing ib port to allow minting")

	waitTransactionConfirmations()

	gravityBuilder := executor.GravityInstructionBuilder{}
	gravityExecutor, err := InitGenericExecutor(
		deployer.PrivateKey,
		gravityProgram.PublicKey.ToBase58(),
		gravityDataAccount.Account.PublicKey.ToBase58(),
		gravityMultisigAccount.Account.PublicKey.ToBase58(),
		RPCEndpoint,
		common.PublicKeyFromString(""),
	)

	nebulaBuilder := executor.NebulaInstructionBuilder{}
	nebulaExecutor, err := InitGenericExecutor(
		deployer.PrivateKey,
		nebulaProgram.PublicKey.ToBase58(),
		nebulaDataAccount.Account.PublicKey.ToBase58(),
		nebulaMultisigAccount.Account.PublicKey.ToBase58(),
		RPCEndpoint,
		gravityDataAccount.Account.PublicKey,
	)
	ValidateError(t, err)

	ibportExecutor, err := InitGenericExecutor(
		deployer.PrivateKey,
		ibportProgram.PublicKey.ToBase58(),
		ibportDataAccount.Account.PublicKey.ToBase58(),
		"",
		RPCEndpoint,
		common.PublicKeyFromString(""),
	)
	ValidateError(t, err)

	oracles := consulsList.ConcatConsuls()

	waitTransactionConfirmations()

	ParallelExecution(
		[]func(){
			func() {
				gravityInitResponse, err := gravityExecutor.BuildAndInvoke(
					gravityBuilder.Init(BFT, 1, oracles),
				)
				fmt.Printf("Gravity Init: %v \n", gravityInitResponse.TxSignature)
				ValidateError(t, err)
			},
			func() {
				// (2)
				nebulaInitResponse, err := nebulaExecutor.BuildAndInvoke(
					nebulaBuilder.Init(BFT, nebula.Bytes, gravityDataAccount.Account.PublicKey, oracles),
				)
				ValidateError(t, err)
				fmt.Printf("Nebula Init: %v \n", nebulaInitResponse.TxSignature)
			},
			func() {
				ibportInitResult, err := ibportExecutor.BuildAndInvoke(
					executor.IBPortIXBuilder.InitWithOracles(nebulaProgram.PublicKey, common.TokenProgramID, tokenDeployResult.Token, BFT, consulsList.ConcatConsuls()),
				)

				fmt.Printf("IB Port Init: %v \n", ibportInitResult.TxSignature)
				ValidateError(t, err)
			},
		},
	)

	waitTransactionConfirmations()

	fmt.Println("IB Port Program is being subscribed to Nebula")

	var subID [16]byte
	rand.Read(subID[:])

	fmt.Printf("subID: %v \n", subID)

	// (4)
	nebulaSubscribePortResponse, err := nebulaExecutor.BuildAndInvoke(
		nebulaBuilder.Subscribe(ibportProgram.PDA, 1, 1, subID),
	)
	ValidateError(t, err)

	fmt.Printf("Nebula Subscribe: %v \n", nebulaSubscribePortResponse.TxSignature)
	fmt.Println("Now checking for valid double spend prevent")

	waitTransactionConfirmations()
	// waitTransactionConfirmations()

	_, err = nebulaExecutor.BuildAndInvoke(
		nebulaBuilder.Subscribe(ibportProgram.PDA, 1, 1, subID),
	)
	ValidateErrorExistence(t, err)

	fmt.Printf("Nebula Subscribe with the same subID must have failed: %v \n", err.Error())

	waitTransactionConfirmations()

	i, requestsCount := 0, 1
	pulseID := 0

	if !OmitSendValueFlow {

		fmt.Printf("send %v attach requests with random amount \n", requestsCount)

		for i < requestsCount {
			swapId := make([]byte, 16)
			rand.Read(swapId)

			attachedAmount := float64(uint64(rand.Float64() * 10))

			var rawDataValue [64]byte
			copy(rawDataValue[:], executor.BuildCrossChainMintByteVector(swapId, common.PublicKeyFromString(deployerTokenAccount), attachedAmount))

			var dataHashForAttach [32]byte

			hashingFunction := func(input []byte) []byte {
				digest := sha256.Sum256(input)
				return digest[:]
			}

			// copy(dataHashForAttach[:], ethcrypto.Keccak256(rawDataValue[:]))
			copy(dataHashForAttach[:], hashingFunction(rawDataValue[:]))

			fmt.Printf("Iteration #%v \n", i)
			fmt.Printf("Amount: %v \n", attachedAmount)
			fmt.Printf("Raw Data Value: %v \n", rawDataValue)
			fmt.Printf("Data Value Hash: %v \n", dataHashForAttach)

			nebulaExecutor.EraseAdditionalMeta()
			nebulaExecutor.SetAdditionalSigners(consulsList.ToBftSigners())
			nebulaExecutor.SetDeployerPK(operatingConsul.Account)

			waitTransactionConfirmations()

			nebulaSendHashValueResponse, err := nebulaExecutor.BuildAndInvoke(
				nebulaBuilder.SendHashValue(dataHashForAttach),
			)
			ValidateError(t, err)

			fmt.Printf("#%v Nebula SendHashValue Call: %v \n", i, nebulaSendHashValueResponse.TxSignature)

			nebulaExecutor.EraseAdditionalSigners()
			nebulaExecutor.SetDeployerPK(operatingConsul.Account)

			nebulaExecutor.SetAdditionalMeta([]types.AccountMeta{
				{PubKey: common.TokenProgramID, IsWritable: false, IsSigner: false},
				{PubKey: ibportProgram.PublicKey, IsWritable: false, IsSigner: false},
				{PubKey: ibportDataAccount.Account.PublicKey, IsWritable: true, IsSigner: false},
				{PubKey: common.PublicKeyFromString(tokenProgramAddress), IsWritable: true, IsSigner: false},
				{PubKey: common.PublicKeyFromString(deployerTokenAccount), IsWritable: true, IsSigner: false},
				{PubKey: ibportProgram.PDA, IsWritable: false, IsSigner: false},
			})

			waitTransactionConfirmations()

			nebulaAttachResponse, err := nebulaExecutor.BuildAndInvoke(
				nebulaBuilder.SendValueToSubs(rawDataValue, nebula.Bytes, uint64(pulseID), subID),
			)
			ValidateError(t, err)
			if err != nil {
				continue
			}

			fmt.Printf("#%v Nebula SendValueToSubs Call:  %v \n", i, nebulaAttachResponse.TxSignature)

			waitTransactionConfirmations()

			i++
			pulseID++
		}

		waitTransactionConfirmations()
	}

	const MaxIBPortRequestsLimit = 15
	amountForUnwrap := 2.227

	fmt.Printf("Reaching limit of unprocessed requests on IB Port \n")

	ibportInstructionBuilder := executor.NewIBPortInstructionBuilder()

	nebulaExecutor.EraseAdditionalMeta()
	nebulaExecutor.EraseAdditionalSigners()
	nebulaExecutor.SetDeployerPK(deployer.Account)

	ibportExecutor.SetAdditionalMeta([]types.AccountMeta{
		{PubKey: common.TokenProgramID, IsWritable: false, IsSigner: false},
		{PubKey: common.PublicKeyFromString(tokenProgramAddress), IsWritable: true, IsSigner: false},
		{PubKey: common.PublicKeyFromString(deployerTokenAccount), IsWritable: true, IsSigner: false},
		{PubKey: ibportProgram.PDA, IsWritable: false, IsSigner: false},
	})

	var allTotallySentByteOperations []executor.PortOperation

	sendNumerousBurnRequests := func(n int) (*models.CommandResponse, error) {
		var instructionBatches []interface{}

		err = DelegateSPLTokenAmount(deployer.PKPath, deployerTokenAccount, ibportProgram.PDA.ToBase58(), amountForUnwrap*float64(MaxIBPortRequestsLimit))
		if err != nil {
			return nil, err
		}

		t.Log("Delegated some tokens to ibport from  deployer")
		t.Log("Creating cross chain transfer tx")

		waitTransactionConfirmations()

		i = 0
		for i < n {
			ethReceiverPK, err := ethcrypto.GenerateKey()
			ValidateError(t, err)

			var ethReceiverAddress [32]byte
			copy(ethReceiverAddress[:], ethcrypto.PubkeyToAddress(ethReceiverPK.PublicKey).Bytes())

			fmt.Printf("Iteration #%v \n", i)
			t.Logf("#%v EVM Receiver:  %v \n", i, ethcrypto.PubkeyToAddress(ethReceiverPK.PublicKey).String())
			t.Logf("#%v EVM Receiver (bytes): %v \n", i, ethReceiverAddress[:])

			fmt.Printf("amountForUnwrap: %v \n", amountForUnwrap)

			ix := ibportInstructionBuilder.CreateTransferUnwrapRequest(ethReceiverAddress, amountForUnwrap)

			instructionBatches = append(instructionBatches, ix)

			castedIx := ix.(executor.CreateTransferUnwrapRequestInstruction)
			portOperation, err := executor.UnpackByteArray(castedIx.Pack()[:])

			fmt.Printf("castedIx %+v \n", castedIx)
			fmt.Printf("portOperation %+v \n", portOperation)

			if err != nil {
				return nil, err
			}

			allTotallySentByteOperations = append(allTotallySentByteOperations, *portOperation)

			i++
		}

		multipleBurnsResult, err := ibportExecutor.InvokeInstructionBatches(
			instructionBatches,
		)
		if err != nil {
			return nil, err
		}

		return multipleBurnsResult, nil
	}

	waitTransactionConfirmations()

	// check for the limit

	approvedLimitBurnsResult, err := sendNumerousBurnRequests(5)
	ValidateError(t, err)
	t.Logf("Sent %v times: CreateTransferUnwrapRequest - Tx: %v \n", i, approvedLimitBurnsResult.TxSignature)

	waitTransactionConfirmations()

	approvedLimitBurnsResult, err = sendNumerousBurnRequests(2)
	ValidateError(t, err)
	t.Logf("Sent %v times: CreateTransferUnwrapRequest - Tx: %v \n", i, approvedLimitBurnsResult.TxSignature)

	waitTransactionConfirmations()

	// approvedLimitBurnsResult, err = sendNumerousBurnRequests(1)
	// ValidateErrorExistence(t, err)

	// t.Logf("+1 On limit unwrap must have failed: %v \n", err)

	// waitTransactionConfirmations()

	// t.Logf("Now - process unconfirmed requests on ib port \n")

	// t.Logf("Setting one of the oracles as the invoker")
	// ibportExecutor.SetDeployerPK(operatingConsul.Account)

	// for j, portOperation := range allTotallySentByteOperations[0 : len(allTotallySentByteOperations)-1] {
	// 	byteArr := portOperation.Pack()
	// 	fmt.Printf("byteArr: %v \n", byteArr)
	// 	fmt.Printf("byteArr(len): %v \n", len(byteArr))
	// 	ix := ibportInstructionBuilder.ConfirmProcessedRequest(portOperation.Pack())

	// 	confirmRes, err := ibportExecutor.BuildAndInvoke(
	// 		ix,
	// 	)
	// 	ValidateError(t, err)

	// 	t.Logf("Confirm Swap #%v: Tx: %v \n", j, confirmRes.TxSignature)

	// 	waitTransactionConfirmations()
	// }

	// allTotallySentByteOperations = make([]executor.PortOperation, 10)

	// approvedLimitBurnsResult, err = sendNumerousBurnRequests(5)
	// ValidateError(t, err)
	// t.Logf("Sent %v times: CreateTransferUnwrapRequest - Tx: %v \n", i, approvedLimitBurnsResult.TxSignature)

	// waitTransactionConfirmations()
}

type Logger struct {
	Tag     string
	counter int
}

func (l *Logger) Log(val interface{}) {
	fmt.Printf("#%v - %v: %v \n", l.counter, l.Tag, val)
	l.counter++
}

func BuildLoggerWithTag(tag string) *Logger {
	return &Logger{Tag: tag}
}

func TestDataRealloc(t *testing.T) {
	var err error

	deployer, err := NewOperatingAddress(t, "../private-keys/_test_deployer-pk-deployer.json", nil)
	ValidateError(t, err)

	testProgram, err := NewOperatingAddress(t, "../private-keys/_test_only-gravity-program.json", nil)
	ValidateError(t, err)

	WrappedFaucet(t, deployer.PKPath, "", 10)

	waitTx := func() {
		time.Sleep(time.Second * 15)
	}

	RPCEndpoint, _ := InferSystemDefinedRPC()

	waitTx()

	txLogger := BuildLoggerWithTag("tx")

	const StartAllocation = 500

	dataAcc := types.NewAccount()

	testDataAccount, err := GenerateNewAccountWithSeed(deployer.PrivateKey, dataAcc, StartAllocation, testProgram.PublicKey.ToBase58(), RPCEndpoint)
	ValidateError(t, err)

	fmt.Printf("#1 Test Data Account: %v \n", testDataAccount.Account.PublicKey.ToBase58())
	fmt.Printf("#1 Test Data Account(PK): %v \n", testDataAccount.Account.PrivateKey)

	txLogger.Log(testDataAccount.TxSignature)

	// testDataAccount2, err := AllocateAccount(deployer.PrivateKey, *testDataAccount.Account, StartAllocation * 3, testProgram.PublicKey.ToBase58(), RPCEndpoint)
	testDataAccount2, err := GenerateNewAccountWithSeed(deployer.PrivateKey, dataAcc, StartAllocation*3, testProgram.PublicKey.ToBase58(), RPCEndpoint)
	ValidateError(t, err)

	fmt.Printf("#2 Test Data Account: %v \n", testDataAccount2.Account.PublicKey.ToBase58())
	fmt.Printf("#2 Test Data Account(PK): %v \n", testDataAccount2.Account.PrivateKey)

	txLogger.Log(testDataAccount2.TxSignature)

	fmt.Printf("#1 == #2: %v \n", bytes.Equal(testDataAccount.Account.PublicKey[:], testDataAccount2.Account.PublicKey[:]))
}

func buildAccountSubscribeRequest(watched string) ws.RequestBody {
	watchRequestParams := []interface{}{
		watched,
		ws.Encoding{
			Encoding:   "base64",
			Commitment: "finalized",
		},
	}

	return ws.RequestBody{
		Jsonrpc: "2.0",
		ID:      1,
		Method:  "accountSubscribe",
		Params:  watchRequestParams,
	}
}

func buildLogsSubscribeRequest(watched string) ws.LogsSubscribeBody {
	return ws.LogsSubscribeBody{
		Jsonrpc: "2.0",
		ID:      1,
		Method:  "logsSubscribe",
		Params: []ws.LogsSubscribeParam{
			{
				Mentions: []string{
					watched,
				},
			},
			{
				Commitment: "finalized",
			},
		},
	}
}
func TestMintWatcher(t *testing.T) {
	var err error

	deployer, err := NewOperatingAddress(t, "../private-keys/_test_deployer-pk-deployer.json", nil)
	ValidateError(t, err)

	// ibportProgram, err := NewOperatingAddress(t, "../private-keys/_test_only_ibport-program.json", &OperatingAddressBuilderOptions{
	// 	WithPDASeeds: []byte(executor.IBPortPDABumpSeeds),
	// })
	// ValidateError(t, err)

	RPCEndpoint, _ := InferSystemDefinedRPC()
	WSEndpoint, err := InferSystemDefinedWebSocketURL()
	ValidateError(t, err)

	// ibportDataAccount, err := GenerateNewAccount(deployer.PrivateKey, IBPortAllocation, ibportProgram.PublicKey.ToBase58(), RPCEndpoint)
	// ValidateError(t, err)

	WrappedFaucet(t, deployer.PKPath, "", 10)

	waitTransactionConfirmations()

	tokenDeployResult, err := CreateToken(deployer.PKPath)
	ValidateError(t, err)

	tokenProgramAddress := tokenDeployResult.Token.ToBase58()

	fmt.Printf("Token deployed: %v \n", tokenDeployResult.Signature)

	deployerTokenAccount, err := CreateTokenAccount(deployer.PKPath, tokenProgramAddress)
	ValidateError(t, err)

	waitTransactionConfirmations()

	// ibportExecutor, err := InitGenericExecutor(
	// 	deployer.PrivateKey,
	// 	ibportProgram.PublicKey.ToBase58(),
	// 	ibportDataAccount.Account.PublicKey.ToBase58(),
	// 	"",
	// 	RPCEndpoint,
	// 	common.PublicKeyFromString(""),
	// )
	// ValidateError(t, err)

	// ibportExecutor.BuildAndInvoke(

	// )

	// waitTransactionConfirmations()

	fmt.Printf("WS Endpoint: %v \n", WSEndpoint)
	// u := url.URL{Scheme: "ws", Host: WSEndpoint, Path: "/" }

	log.Printf("connecting to %s", WSEndpoint)

	c, _, err := websocket.DefaultDialer.Dial(WSEndpoint, nil)
	ValidateError(t, err)

	defer c.Close()

	done := make(chan struct{})

	// {
	// 	"jsonrpc": "2.0",
	// 	"id": 1,
	// 	"method": "accountSubscribe",
	// 	"params": [
	// 	  "CM78CPUeXjn8o3yroDHxUtKsZZgoy4GPkPPXfouKNH12",
	// 	  {
	// 		"encoding": "base64",
	// 		"commitment": "finalized"
	// 	  }
	// 	]
	//   }

	// mint some tokens for deployer
	go func() {
		time.Sleep(time.Second * 3)
		err = MintToken(deployer.PKPath, tokenProgramAddress, 1_000_000, deployerTokenAccount)
		ValidateError(t, err)

		t.Log("Minted some tokens")
	}()

	// buildAccountUnsubscribeRequest := func(subID int64) WSRequestBody {
	// 	return WSRequestBody {
	// 		Jsonrpc: "2.0",
	// 		ID: 1,
	// 		Method: "accountSubscribe",
	// 		Params: []interface{} {
	// 			subID,
	// 		},
	// 	}
	// }

	// watchRequest := buildAccountSubscribeRequest(deployerTokenAccount)
	watchRequest := buildLogsSubscribeRequest(deployerTokenAccount)

	watchRequestBytes, err := json.Marshal(&watchRequest)
	ValidateError(t, err)

	err = c.WriteMessage(websocket.TextMessage, watchRequestBytes)
	ValidateError(t, err)

	// go func() {
	// 	defer close(done)

	// 	for {
	// 		_, message, err := c.ReadMessage()
	// 		if err != nil {
	// 			log.Println("read:", err)
	// 			return
	// 		}
	// 		log.Printf("recv: %s", message)
	// 	}
	// }()
	defer close(done)

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}
		log.Printf("recv: %s", message)

		var responseUnpacked ws.LogsSubscribeNotification
		err = json.Unmarshal(message, &responseUnpacked)
		if err != nil {
			fmt.Printf("Error on WSLogsSubscribeNotification unpack: %v \n", err)
			continue
		}

		txID := responseUnpacked.Params.Result.Value.Signature
		if txID == "" {
			continue
		}

		solanaClient := solclient.NewClient(RPCEndpoint)

		ctx := context.Background()

		response, err := solanaClient.GetConfirmedTransaction(ctx, txID)
		ValidateError(t, err)

		fmt.Printf("RESPONSE: %+v \n", response)
		// var responseUnpacked WSAccountNotification
		// err = json.Unmarshal(message, &responseUnpacked)
		// if err != nil {
		// 	fmt.Printf("Error on WSAccountNotification unpack: %v \n", err)
		// 	continue
		// }

		// if len(responseUnpacked.Params.Result.Value.Data) == 0 {
		// 	continue
		// }

		// accountValue := responseUnpacked.Params.Result.Value.Data[0]
		// decodedAccountValue, err := base64.StdEncoding.DecodeString(accountValue)
		// ValidateError(t, err)

		// tokenAccount, err := tokenprog.TokenAccountFromData(decodedAccountValue)
		// ValidateError(t, err)

		// if err != nil {
		// 	// log.Println("read:", err)
		// 	fmt.Printf("Error on TokenAccount unpack: %v \n", err)
		// 	continue
		// }

		// log.Printf("recv: %s", message)

		// fmt.Printf("%+v \n", tokenAccount)
	}

	// ticker := time.NewTicker(time.Second)
	// defer ticker.Stop()

	// for {
	// 	select {
	// 	case <-done:
	// 		return
	// 	case t := <-ticker.C:
	// 		err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
	// 		if err != nil {
	// 			log.Println("write:", err)
	// 			return
	// 		}
	// 	}
	// }
}
