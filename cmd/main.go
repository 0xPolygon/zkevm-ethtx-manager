package main

import (
	"os"

	"github.com/0xPolygonHermez/zkevm-ethtx-manager"
	"github.com/0xPolygonHermez/zkevm-ethtx-manager/config"
	"github.com/0xPolygonHermez/zkevm-ethtx-manager/log"
	"github.com/urfave/cli/v2"
)

const appName = "zkevm-ethtx-manager"

const (
	// ETHTXMANAGER is the service that manages the tx sent to L1
	ETHTXMANAGER = "ethtx-manager"
)

const (
	// NODE_CONFIGFILE name to identify the node config-file
	NODE_CONFIGFILE = "node"
	// NETWORK_CONFIGFILE name to identify the netowk_custom (genesis) config-file
	NETWORK_CONFIGFILE = "custom_network"
)

var (
	configFileFlag = cli.StringFlag{
		Name:     config.FlagCfg,
		Aliases:  []string{"c"},
		Usage:    "Configuration `FILE`",
		Required: true,
	}
	networkFlag = cli.StringFlag{
		Name:     config.FlagNetwork,
		Aliases:  []string{"net"},
		Usage:    "Load default network configuration. Supported values: [`mainnet`, `testnet`, `custom`]",
		Required: true,
	}
	customNetworkFlag = cli.StringFlag{
		Name:     config.FlagCustomNetwork,
		Aliases:  []string{"net-file"},
		Usage:    "Load the network configuration file if --network=custom",
		Required: false,
	}
	yesFlag = cli.BoolFlag{
		Name:     config.FlagYes,
		Aliases:  []string{"y"},
		Usage:    "Automatically accepts any confirmation to execute the command",
		Required: false,
	}
	componentsFlag = cli.StringSliceFlag{
		Name:     config.FlagComponents,
		Aliases:  []string{"co"},
		Usage:    "List of components to run",
		Required: false,
		Value:    cli.NewStringSlice(ETHTXMANAGER),
	}
	migrationsFlag = cli.BoolFlag{
		Name:     config.FlagMigrations,
		Aliases:  []string{"mig"},
		Usage:    "Blocks the migrations in stateDB to not run them",
		Required: false,
	}
)

func main() {
	app := cli.NewApp()
	app.Name = appName
	app.Version = zkevm.Version
	flags := []cli.Flag{
		&configFileFlag,
		&yesFlag,
		&componentsFlag,
	}
	app.Commands = []*cli.Command{
		{
			Name:    "version",
			Aliases: []string{},
			Usage:   "Application version and build",
			Action:  versionCmd,
		},
		{
			Name:    "run",
			Aliases: []string{},
			Usage:   "Run the zkevm-ethtx-manager",
			Action:  start,
			Flags:   append(flags, &networkFlag, &customNetworkFlag, &migrationsFlag),
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
