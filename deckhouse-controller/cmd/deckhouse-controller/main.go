package main

import (
	"fmt"
	_ "net/http/pprof"
	"os"

	ad_app "github.com/flant/addon-operator/pkg/app"
	sh_app "github.com/flant/shell-operator/pkg/app"
	sh_debug "github.com/flant/shell-operator/pkg/debug"
	utils_signal "github.com/flant/shell-operator/pkg/utils/signal"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/deckhouse/deckhouse/deckhouse-controller/pkg/app"
	"github.com/deckhouse/deckhouse/deckhouse-controller/pkg/deckhouse"
	"github.com/deckhouse/deckhouse/deckhouse-controller/pkg/helpers"
)

// Variables with component versions. They set by 'go build' command.
var (
	DeckhouseVersion     = "dev"
	AddonOperatorVersion = "dev"
	ShellOperatorVersion = "dev"
)

func main() {
	sh_app.Version = ShellOperatorVersion
	ad_app.Version = AddonOperatorVersion

	kpApp := kingpin.New(app.AppName, fmt.Sprintf("%s %s: %s", app.AppName, DeckhouseVersion, app.AppDescription))

	// override usage template to reveal additional commands with information about start command
	kpApp.UsageTemplate(sh_app.OperatorUsageTemplate(app.AppName))

	// print version
	kpApp.Command("version", "Show version.").Action(func(c *kingpin.ParseContext) error {
		fmt.Printf("deckhouse %s (addon-operator %s, shell-operator %s)", DeckhouseVersion, AddonOperatorVersion, ShellOperatorVersion)
		return nil
	})

	// start main loop
	startCmd := kpApp.Command("start", "Start deckhouse.").
		Default().
		Action(func(c *kingpin.ParseContext) error {
			sh_app.SetupLogging()
			log.Infof("deckhouse %s (addon-operator %s, shell-operator %s)", DeckhouseVersion, AddonOperatorVersion, ShellOperatorVersion)

			// Set hook metrics listen port if flat is not passed.
			if sh_app.HookMetricsListenPort == "" {
				sh_app.HookMetricsListenPort = app.DeckhouseHookMetricsListenPort
			}

			operator := deckhouse.DefaultDeckhouse()
			err := deckhouse.InitAndStart(operator)
			if err != nil {
				os.Exit(1)
			}

			// Block action by waiting signals from OS.
			utils_signal.WaitForProcessInterruption(func() {
				operator.Shutdown()
				os.Exit(1)
			})

			return nil
		})
	// Set default log type as json
	sh_app.LogType = app.DeckhouseLogTypeDefault
	sh_app.KubeClientQpsDefault = app.DeckshouseKubeClientQPSDefault
	sh_app.KubeClientBurstDefault = app.DeckshouseKubeClientBurstDefault
	app.DefineStartCommandFlags(startCmd)
	ad_app.DefineStartCommandFlags(kpApp, startCmd)

	// Add debug commands from shell-operator and addon-operator
	sh_debug.DefineDebugCommands(kpApp)
	ad_app.DefineDebugCommands(kpApp)

	// deckhouse-controller helper subcommands
	helpers.DefineHelperCommands(kpApp)

	kingpin.MustParse(kpApp.Parse(os.Args[1:]))
}
