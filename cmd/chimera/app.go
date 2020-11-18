package cmd

import (
	"os"

	"github.com/chimera-kube/chimera-admission/internal/pkg/chimera"

	"github.com/urfave/cli/v2"
)

const (
	admissionPort        = 8080
	exportedEnvVarPrefix = "AW_EXPORT_"
)

var (
	admissionName = "wasm.admission.rule"
	admissionHost = os.Getenv("AW_CALLBACK_HOST")
	apiGroups     string
	apiVersions   string
	resources     string
	validatePath  string
	wasmUri       string
	wasmEnvVars   cli.StringSlice

	wasmWorker *chimera.WasmWorker
)

func NewApp() *cli.App {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "api-groups",
				Value:       "*",
				Usage:       "Admission Rule - APIGroups",
				EnvVars:     []string{"AW_API_GROUPS"},
				Destination: &apiGroups,
			},
			&cli.StringFlag{
				Name:        "api-versions",
				Value:       "v1",
				Usage:       "Admission Rule - APIVersions",
				EnvVars:     []string{"AW_API_VERSIONS"},
				Destination: &apiVersions,
			},
			&cli.StringFlag{
				Name:        "resources",
				Value:       "*",
				Usage:       "Admission Rule - Resources",
				EnvVars:     []string{"AW_RESOURCES"},
				Destination: &resources,
			},
			&cli.StringFlag{
				Name:        "validate-path",
				Value:       "/validate",
				Usage:       "Admission Rule - Validate path",
				EnvVars:     []string{"AW_VALIDATE_PATH"},
				Destination: &validatePath,
			},
			&cli.StringFlag{
				Name:        "wasm-uri",
				Usage:       "WASM URI (file:///, localhost:5000/project/artifact:version)",
				EnvVars:     []string{"AW_WASM_URI"},
				Destination: &wasmUri,
			},
			&cli.StringSliceFlag{
				Name:        "env",
				Usage:       "Admission Rule - Export environment variable on the guest WASM module (VAR=value), can be repeated several times",
				Destination: &wasmEnvVars,
			},
		},
		Action: startServer,
	}

	return app
}
