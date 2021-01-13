package cmd

import (
	"github.com/chimera-kube/chimera-admission/internal/pkg/chimera"

	"github.com/urfave/cli/v2"
)

const (
	exportedEnvVarPrefix = "CHIMERA_EXPORT_"
)

var (
	admissionName             string
	skipAdmissionRegistration bool
	admissionHost             string
	admissionPort             int
	kubeNamespace             string
	kubeService               string
	apiGroups                 cli.StringSlice
	apiVersions               cli.StringSlice
	resources                 cli.StringSlice
	operations                cli.StringSlice
	validatePath              string
	wasmUri                   string
	wasmRemoteCA              string
	wasmRemoteInsecure        bool
	wasmRemoteNonTLS          bool
	wasmEnvVars               cli.StringSlice
	tlsExtraSANs              cli.StringSlice
	certFile                  string
	keyFile                   string
	caFile                    string
	insecureServer            bool

	wasmWorker *chimera.WasmWorker
)

func NewApp() *cli.App {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "insecure-server",
				Value:       false,
				Usage:       "Start insecure HTTP server",
				EnvVars:     []string{"CHIMERA_INSECURE_SERVER"},
				Destination: &insecureServer,
			},
			&cli.StringFlag{
				Name:        "admission-name",
				Value:       "chimera.admission.rule",
				Usage:       "Name used to register the admission controller against Kubernetes",
				EnvVars:     []string{"CHIMERA_ADMISSION_NAME"},
				Destination: &admissionName,
			},
			&cli.BoolFlag{
				Name:        "skip-admission-registration",
				Value:       false,
				Usage:       "Skips the admission registration on the Kubernetes API",
				EnvVars:     []string{"CHIMERA_SKIP_ADMISSION_REGISTRATION"},
				Destination: &skipAdmissionRegistration,
			},
			&cli.StringFlag{
				Name:        "callback-host",
				Usage:       "FQDN of the admission controller - must be reachable by the Kubernetes API server",
				EnvVars:     []string{"CHIMERA_CALLBACK_HOST"},
				Destination: &admissionHost,
			},
			&cli.IntFlag{
				Name:        "callback-port",
				Usage:       "Listening port",
				Value:       8443,
				EnvVars:     []string{"CHIMERA_CALLBACK_PORT"},
				Destination: &admissionPort,
			},
			&cli.StringFlag{
				Name:        "kube-namespace",
				Usage:       "The namespace that contains the chimera-admission service",
				EnvVars:     []string{"CHIMERA_KUBE_NAMESPACE"},
				Destination: &kubeNamespace,
			},
			&cli.StringFlag{
				Name:        "kube-service",
				Usage:       "The name of the kubernetes service exposing chimera-admission",
				EnvVars:     []string{"CHIMERA_KUBE_SERVICE"},
				Destination: &kubeService,
			},
			&cli.StringFlag{
				Name:        "cert-file",
				Usage:       "TLS certificate to use",
				EnvVars:     []string{"CHIMERA_CERT_FILE"},
				Destination: &certFile,
			},
			&cli.StringFlag{
				Name:        "cert-key",
				Usage:       "TLS key to use",
				EnvVars:     []string{"CHIMERA_KEY_FILE"},
				Destination: &keyFile,
			},
			&cli.StringFlag{
				Name:        "ca-bundle",
				Usage:       "CA Bundle",
				EnvVars:     []string{"CHIMERA_CA_BUNDLE"},
				Destination: &caFile,
			},
			&cli.StringSliceFlag{
				Name:        "tls-extra-sans",
				Usage:       "Extra TLS SANs to use when generating certificate. Can be repeated several times",
				Destination: &tlsExtraSANs,
			},
			&cli.StringSliceFlag{
				Name:        "api-groups",
				Value:       cli.NewStringSlice("*"),
				Usage:       "Admission Rule - APIGroups",
				EnvVars:     []string{"CHIMERA_API_GROUPS"},
				Destination: &apiGroups,
			},
			&cli.StringSliceFlag{
				Name:        "api-versions",
				Value:       cli.NewStringSlice("v1"),
				Usage:       "Admission Rule - APIVersions",
				EnvVars:     []string{"CHIMERA_API_VERSIONS"},
				Destination: &apiVersions,
			},
			&cli.StringSliceFlag{
				Name:        "resources",
				Value:       cli.NewStringSlice("*"),
				Usage:       "Admission Rule - Resources",
				EnvVars:     []string{"CHIMERA_RESOURCES"},
				Destination: &resources,
			},
			&cli.StringSliceFlag{
				Name:        "operations",
				Value:       cli.NewStringSlice("*"),
				Usage:       "Admission Rule - Operations",
				EnvVars:     []string{"CHIMERA_OPERATIONS"},
				Destination: &operations,
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
				Usage:       "Wasm URI (file:///some/local/program.wasm, https://some-host.com/some/remote/program.wasm, registry://localhost:5000/project/artifact:some-version)",
				EnvVars:     []string{"CHIMERA_WASM_URI"},
				Destination: &wasmUri,
			},
			&cli.StringFlag{
				Name:        "wasm-remote-ca",
				Usage:       "CA used by the remote location hosting the Wasm module",
				EnvVars:     []string{"CHIMERA_WASM_REMOTE_CA"},
				Destination: &wasmRemoteCA,
			},
			&cli.BoolFlag{
				Name:        "wasm-remote-non-tls",
				Usage:       "Wasm remote endpoint is not using TLS. False by default",
				EnvVars:     []string{"CHIMERA_WASM_REMOTE_NON_TLS"},
				Value:       false,
				Destination: &wasmRemoteNonTLS,
			},
			&cli.BoolFlag{
				Name:        "wasm-remote-insecure",
				Usage:       "Do not verify remote TLS certificate. False by default",
				EnvVars:     []string{"CHIMERA_WASM_REMOTE_INSECURE"},
				Value:       false,
				Destination: &wasmRemoteInsecure,
			},
			&cli.StringSliceFlag{
				Name:        "env",
				Usage:       "Admission Rule - Export environment variable on the guest Wasm module (VAR=value), can be repeated several times",
				Destination: &wasmEnvVars,
			},
			&cli.BoolFlag{
				Name:  "debug",
				Usage: "Enable debug output",
				Value: false,
			},
		},
		Action: startServer,
	}

	return app
}
