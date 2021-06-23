package commands

import (

	// "github.com/Gravity-Tech/solanoid/commands/executor"
	"github.com/Gravity-Tech/solanoid/commands/executor"

	"github.com/portto/solana-go-sdk/common"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// UpdateConsulsPrivateKey string
	// GravityProgramID        string
	// GravityDataAccount      string
	// Round                   uint64
	// alias for show
	NebulaDataAccount     string
	initNebulaContractCmd = &cobra.Command{
		Hidden: false,

		Use:   "init-nebula",
		Short: "Display a file from the hoarder storage",
		Long:  ``,
		Run:   initNebula,
	}
)

// init
func init() {
	initNebulaContractCmd.Flags().StringVarP(&GravityProgramID, "program", "p", "", "Program ID")
	viper.BindPFlag("program", initNebulaContractCmd.Flags().Lookup("program"))
	initNebulaContractCmd.MarkFlagRequired("program")

	initNebulaContractCmd.Flags().StringVarP(&GravityDataAccount, "data-account", "d", "", "Gravity Data Account")
	viper.BindPFlag("data-account", initNebulaContractCmd.Flags().Lookup("data-account"))
	initNebulaContractCmd.MarkFlagRequired("data-account")

	initNebulaContractCmd.Flags().StringVarP(&NebulaDataAccount, "nebula-account", "n", "", "Gravity multisig Account")
	viper.BindPFlag("nebula-account", initNebulaContractCmd.Flags().Lookup("nebula-account"))
	initNebulaContractCmd.MarkFlagRequired("nebula-account")

	initNebulaContractCmd.Flags().StringVarP(&MultisigDataAccount, "multisig-account", "m", "", "Gravity multisig Account")
	viper.BindPFlag("multisig-account", initNebulaContractCmd.Flags().Lookup("multisig-account"))
	initNebulaContractCmd.MarkFlagRequired("multisig-account")

	initNebulaContractCmd.Flags().StringVarP(&UpdateConsulsPrivateKey, "private-key", "k", "", "private key in base58 encoding")
	viper.BindPFlag("private-key", initNebulaContractCmd.Flags().Lookup("private-key"))
	initNebulaContractCmd.MarkFlagRequired("private-key")

	SolanoidCmd.AddCommand(initNebulaContractCmd)
}


func InitGenericExecutor(privateKey, nebulaProgramID, dataAccount, multisigDataAccount, clientEndpoint string, gravityProgramID common.PublicKey) (*executor.GenericExecutor, error) {
	nebulaExec, err := executor.NewNebulaExecutor(privateKey, nebulaProgramID, dataAccount, multisigDataAccount, clientEndpoint, gravityProgramID)
	if err != nil {
		return nil, err
	}

	return nebulaExec, nil
}


func initNebula(ccmd *cobra.Command, args []string) {
	endpoint, _ := InferSystemDefinedRPC()

	_, _ = InitGenericExecutor(UpdateConsulsPrivateKey, GravityDataAccount, NebulaDataAccount, MultisigDataAccount, endpoint, common.PublicKeyFromString(GravityProgramID))
}
