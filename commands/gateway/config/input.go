package config

import "fmt"

type AbstractValidatable interface {
	Validate() error
}

type CrossChainTokenConfig struct {
	AssetID        string
	NodeURL        string
	GravityAddress string
	ChainType      string
	ConsulsList  []string
	NebulaScriptPath string
	SubscriberScriptPath string
}

func (tokenCfg CrossChainTokenConfig) Validate() error {
	if tokenCfg.GravityAddress == "" {
		return fmt.Errorf("empty gravity address")
	}
	if tokenCfg.NodeURL == "" {
		return fmt.Errorf("node url is empty")
	}
	if tokenCfg.GravityAddress == "" {
		return fmt.Errorf("empty gravity address")
	}
	if tokenCfg.ChainType == "" {
		return fmt.Errorf("chain type not provided")
	}
	if len(tokenCfg.ConsulsList) == 0 {
		return fmt.Errorf("consuls list is empty")
	}

	return nil
}

type CommonInputConfig struct {
	Bft int
}

type DeployInputConfig struct {
	CommonInputConfig
	OriginToken CrossChainTokenConfig
	DestToken   CrossChainTokenConfig
}

func (deployConfig DeployInputConfig) Validate() error {
	var err error
	err = deployConfig.OriginToken.Validate()
	if err != nil {
		return err
	}

	err = deployConfig.DestToken.Validate()
	if err != nil {
		return err
	}

	if deployConfig.Bft <= 0 {
		return fmt.Errorf("bft value is less than or equal to zero")
	}

	return nil
}
