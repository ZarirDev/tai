package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zarirdev/tai/groq"
	"github.com/zarirdev/tai/pkg/config"
	"github.com/zarirdev/tai/pkg/keystore"
)

var RootCmd = rootCmd

func GetRootCommand() *cobra.Command {
	return RootCmd
}

var (
	cfgFile    string
	apiKeyFlag string
)

var rootCmd = &cobra.Command{
	Use:   "tai [flags] <query...>",
	Short: "Ask Groq's AI with built‑in web search",
	Long: `tai is a CLI chatbot powered by Groq's compound models.
It supports web search and can be configured via command‑line flags
or a YAML config file (~/.config/tai/config.yaml).`,
	// Custom Args: only require query if --api is not used
	Args: func(cmd *cobra.Command, args []string) error {
		if apiKeyFlag != "" {
			return nil
		}
		if len(args) < 1 {
			return fmt.Errorf("requires at least 1 argument (the question)")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// If --api is set, handle it and exit
		if apiKeyFlag != "" {
			setAPIKey(cmd, apiKeyFlag)
			return
		}

		query := strings.Join(args, " ")

		v := viper.GetViper()

		// Decrypt API key
		encryptedKey := v.GetString("api_key_encrypted")
		apiKey, err := keystore.Decrypt(encryptedKey)
		if err != nil {
			log.Fatalf("Failed to decrypt API key: %v (run 'tai --api <key>' first)", err)
		}
		if apiKey == "" {
			log.Fatal("No API key configured. Use 'tai --api <key>' to set it.")
		}

		model := v.GetString("model")
		maxTokens := v.GetInt("max_tokens")
		includeDomains := v.GetStringSlice("include_domains")
		excludeDomains := v.GetStringSlice("exclude_domains")
		debug := v.GetBool("debug")

		if debug {
			log.Println("🔍 Debug mode ON")
			log.Printf("Model      : %s", model)
			log.Printf("Max tokens : %d", maxTokens)
			log.Printf("Include    : %v", includeDomains)
			log.Printf("Exclude    : %v", excludeDomains)
			log.Printf("➡️  Query: %s", query)
		}

		answer, err := groq.Ask(query, apiKey, model, maxTokens, includeDomains, excludeDomains)
		if err != nil {
			log.Fatalf("❌ Error: %v", err)
		}

		if debug {
			log.Println("✅ Answer received")
		}
		// Use cmd.Println so output goes to the command's writer (testable)
		cmd.Println(answer)
	},
}

// setAPIKey encrypts the provided key and saves it to the user config.
func setAPIKey(cmd *cobra.Command, key string) {
	if key == "" {
		cmd.Print("Enter Groq API key: ")
		fmt.Scanln(&key)
		if key == "" {
			log.Fatal("No key entered")
		}
	}

	encrypted, err := keystore.Encrypt(key)
	if err != nil {
		log.Fatalf("Encryption failed: %v", err)
	}

	v := viper.GetViper()
	v.Set("api_key_encrypted", encrypted)

	if err := config.SaveUserConfig(v); err != nil {
		log.Fatalf("Failed to save config: %v", err)
	}

	cmd.Println("✅ API key encrypted and saved to ~/.config/tai/config.yaml")
}

func init() {
	// Flags bound to Viper
	rootCmd.Flags().BoolP("debug", "d", false, "Enable debug output")
	rootCmd.Flags().String("model", "", "Groq model name (overrides config)")
	rootCmd.Flags().Int("max-tokens", 0, "Maximum response tokens")
	rootCmd.Flags().StringSlice("include-domain", nil, "Include search domains (can repeat)")
	rootCmd.Flags().StringSlice("exclude-domain", nil, "Exclude search domains (can repeat)")

	// --api flag (does not require a query)
	rootCmd.Flags().StringVar(&apiKeyFlag, "api", "", "Set API key permanently (encrypted)")

	// Bind flags to Viper keys
	viper.BindPFlag("debug", rootCmd.Flags().Lookup("debug"))
	viper.BindPFlag("model", rootCmd.Flags().Lookup("model"))
	viper.BindPFlag("max_tokens", rootCmd.Flags().Lookup("max-tokens"))
	viper.BindPFlag("include_domains", rootCmd.Flags().Lookup("include-domain"))
	viper.BindPFlag("exclude_domains", rootCmd.Flags().Lookup("exclude-domain"))

	cobra.OnInitialize(initConfig)
}

func initConfig() {
	v, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Config error: %v", err)
	}
	// Merge into the global Viper instance
	*viper.GetViper() = *v
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
