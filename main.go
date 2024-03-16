package main

import (
	"fmt"
	"os"

	"github.com/antham/fnf/component"
	"github.com/antham/fnf/forward"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/viper"
)

const envPrefix = "FNF"

func main() {
	var endpoint, appKey, appSecret, consumerKey, domain, defaultEmail string
	viper.SetEnvPrefix(envPrefix)
	viper.AutomaticEnv()
	for key, target := range map[string]*string{
		"OVH_ENDPOINT":     &endpoint,
		"OVH_APP_KEY":      &appKey,
		"OVH_APP_SECRET":   &appSecret,
		"OVH_CONSUMER_KEY": &consumerKey,
		"OVH_DOMAIN":       &domain,
		"DEFAULT_EMAIL":    &defaultEmail,
	} {
		if !viper.IsSet(key) {
			fmt.Printf("%s_%s environment variable is not defined\n", envPrefix, key)
			os.Exit(1)
		}
		*target = viper.GetString(key)
	}
	forward, err := forward.NewOVHProvider(endpoint, appKey, appSecret, consumerKey, domain, defaultEmail)
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
	m, err := component.NewForwardList(forward)
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
