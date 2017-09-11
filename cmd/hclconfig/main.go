// Program hclconfig is a CLI that helps with
// encrypting HCL config files.
package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/jjeffery/hclconfig/amzn"
	"github.com/spf13/cobra"
)

var (
	// errUsagePrinted is returned when usage has been printed.
	// Special handling is required because we do not want an error
	// message printed, but we want the process to exit with an error code.
	errUsagePrinted = errors.New("usage printed")
)

func main() {
	cmd := rootCommand()
	if err := cmd.Execute(); err != nil {
		if err == errUsagePrinted {
			os.Exit(1)
		}
		log.Fatalln("error:", err)
	}
}

func rootCommand() *cobra.Command {
	programName := strings.TrimSuffix(strings.ToLower(filepath.Base(os.Args[0])), ".exe")
	cmd := &cobra.Command{
		Short:         "manage secrets in HCL config files",
		Use:           programName,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.HelpFunc()(cmd, args)
			return errUsagePrinted
		},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			log.SetFlags(0)
			return nil
		},
	}
	cmd.AddCommand(encryptCommand())
	cmd.AddCommand(decryptCommand())
	cmd.AddCommand(generateCommand())
	return cmd
}

func decryptCommand() *cobra.Command {
	const long = `
	Reads the file at location and decrypts any secrets in that file.
	
	The location can be a HTTP(S) URL, and S3 URL or a local file.
	
	The decrypted file is written to standard output, unless the --inplace
	flag is specified, in which case it will overwrite the existing file.
	This only works for local files.
	`
	var inplace bool
	cmd := &cobra.Command{
		Short: "decrypt secrets in HCL file",
		Use:   "decrypt <location>",
		Long:  long,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := decryptFile(args[0], inplace); err != nil {
				return err
			}
			return nil
		},
		PreRunE: requireOneFilename,
	}
	cmd.Flags().BoolVar(&inplace, "inplace", inplace, "update file inplace")
	return cmd
}

func encryptCommand() *cobra.Command {
	const long = `
Reads the file at location and encrypts any secrets in that file.
A secret is identified if the configuration key contains any of the
keywords, or the configuration value contains any of the valwords.

The location can be a HTTP(S) URL, and S3 URL or a local file.

The encrypted file is written to standard output, unless the --inplace
flag is specified, in which case it will overwrite the existing file.
This only works for local files.
`
	keywords := []string{
		"password",
		"secret",
		"apikey",
	}
	values := []string{
		"password=",
	}
	var inplace bool
	cmd := &cobra.Command{
		Short: "encrypt secrets in HCL file",
		Use:   "encrypt <location>",
		Long:  long,
		RunE: func(cmd *cobra.Command, args []string) error {

			if err := encryptFile(args[0], inplace, keywords, values); err != nil {
				return err
			}
			return nil
		},
		PreRunE: requireOneFilename,
	}
	cmd.Flags().BoolVar(&inplace, "inplace", inplace, "update file in place")
	cmd.PersistentFlags().StringSliceVar(&keywords, "keywords", keywords, "keywords to encrypt")
	cmd.PersistentFlags().StringSliceVar(&values, "values", values, "values to encrypt")
	return cmd
}

func requireOneFilename(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		fmt.Println("expected file name")
		return errUsagePrinted
	}
	if len(args) > 1 {
		fmt.Println("only one file name may be specified")
		return errUsagePrinted
	}
	return nil
}

func generateCommand() *cobra.Command {
	tmpl := template.Must(template.New("hcl").Parse(`
encryption {
    // {{.KeyARN}}
    {{if .Alias}}// {{.Alias}}{{end}}
    kms = "{{.DataKey}}"
}
`))

	cmd := &cobra.Command{
		Short: "generate data key for use in HCL config file",
		Use:   "generate <kms-key-id>",
		Long:  "generates a data key using the AWS KMS key ID, which can be an ARN or an alias",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				fmt.Println("expected KMS key ID")
				return errUsagePrinted
			}
			keyID := args[0]
			dataKey, keyARN, err := amzn.GenerateDataKey(args[0])
			if err != nil {
				return err
			}

			var data struct {
				KeyARN  string
				Alias   string
				DataKey string
			}

			data.DataKey = dataKey
			data.KeyARN = keyARN
			if strings.HasPrefix(keyID, "alias/") {
				data.Alias = keyID
			}
			if err := tmpl.Execute(os.Stdout, data); err != nil {
				return err
			}
			return nil
		},
	}
	return cmd
}
