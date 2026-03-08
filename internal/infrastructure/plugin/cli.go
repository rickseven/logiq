package plugin

import (
	"fmt"
	"os"
	"path/filepath"
)

// ListPlugins reads the ./plugins/ directory and outputs installed plugins
func ListPlugins() {
	fmt.Println("Installed plugins")
	fmt.Println()

	dirs, err := os.ReadDir("plugins")
	if err != nil {
		fmt.Println("No plugins installed.")
		return
	}

	count := 0
	for _, d := range dirs {
		if d.IsDir() {
			fmt.Println(d.Name())
			count++
		}
	}
	if count == 0 {
		fmt.Println("No plugins installed.")
	}
}

// InstallPlugin scaffolds a generic plugin
func InstallPlugin(name string) {
	path := filepath.Join("plugins", name)
	if err := os.MkdirAll(path, 0755); err != nil {
		fmt.Printf("Error creating plugin directory: %v\n", err)
		return
	}
	fmt.Printf("Plugin %s installed successfully.\n", name)
}

// EnablePlugin hypothetically enables a plugin in config
func EnablePlugin(name string) {
	fmt.Printf("Plugin %s statically enabled.\n", name)
}

// DisablePlugin hypothetically disables a plugin config
func DisablePlugin(name string) {
	fmt.Printf("Plugin %s statically disabled.\n", name)
}

// HandleCLI processes plugin command line arguments
func HandleCLI(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: logiq plugins <list|install|enable|disable> [name]")
		os.Exit(1)
	}

	cmd := args[0]
	switch cmd {
	case "list":
		ListPlugins()
	case "install":
		if len(args) < 2 {
			fmt.Println("Usage: logiq plugins install <name>")
			return
		}
		InstallPlugin(args[1])
	case "enable":
		if len(args) < 2 {
			fmt.Println("Usage: logiq plugins enable <name>")
			return
		}
		EnablePlugin(args[1])
	case "disable":
		if len(args) < 2 {
			fmt.Println("Usage: logiq plugins disable <name>")
			return
		}
		DisablePlugin(args[1])
	default:
		fmt.Println("Unknown plugin command. Use list, install, enable, or disable.")
	}
}
