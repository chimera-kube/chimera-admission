package cmd

import (
	"os"

	"github.com/chimera-kube/chimera-admission/internal/pkg/chimera"

	"github.com/urfave/cli/v2"
)

const (
	admissionPort        = 8080
	exportedEnvVarPrefix = "CHIMERA_EXPORT_"
)

var (
	admissionName = "wasm.admission.rule"
	admissionHost = os.Getenv("CHIMERA_CALLBACK_HOST")
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
				EnvVars:     []string{"CHIMERA_API_GROUPS"},
				Destination: &apiGroups,
			},
			&cli.StringFlag{
				Name:        "api-versions",
				Value:       "v1",
				Usage:       "Admission Rule - APIVersions",
				EnvVars:     []string{"CHIMERA_API_VERSIONS"},
				Destination: &apiVersions,
			},
			&cli.StringFlag{
				Name:        "resources",
				Value:       "*",
				Usage:       "Admission Rule - Resources",
				EnvVars:     []string{"CHIMERA_RESOURCES"},
				Destination: &resources,
			},
			&cli.StringFlag{
				Name:        "validate-path",
				Value:       "/validate",
				Usage:       "Admission Rule - Validate path",
				EnvVars:     []string{"CHIMERA_VALIDATE_PATH"},
				Destination: &validatePath,
			},
			&cli.StringFlag{
				Name:        "wasm-uri",
				Usage:       "WASM URI (file:///some/local/program.wasm, https://some-host.com/some/remote/program.wasm, registry://localhost:5000/project/artifact:some-version)",
				EnvVars:     []string{"CHIMERA_WASM_URI"},
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
