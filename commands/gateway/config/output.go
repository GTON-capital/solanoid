package config

type Account struct {
	Address, PrivKey string
}

type CrossChainDeploymentOutput struct {
	Gravity, Nebula, Port Account
	Token                 string
}

type Output struct {
	Bft                 int
	Origin, Destination CrossChainDeploymentOutput
}
