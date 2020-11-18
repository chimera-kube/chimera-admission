package cmd

import (
	"encoding/json"
	"os"

	"github.com/chimera-kube/chimera-admission/internal/pkg/chimera"
	chimeralib "github.com/chimera-kube/chimera/pkg/chimera"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	admissionv1 "k8s.io/api/admission/v1"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
)

func startServer(c *cli.Context) error {
	if wasmUri == "" {
		return errors.New("Please, provide a WASM URI to load")
	}

	var wasmModulePath string

	moduleSource, modulePath, err := chimera.WASMModuleSource(wasmUri)
	if err != nil {
		return err
	}

	switch moduleSource {
	case chimera.FileSource:
		wasmModulePath = modulePath
	case chimera.HTTPSource, chimera.RegistrySource:
		var err error
		wasmModulePath, err = chimera.FetchRemoteWASMModule(moduleSource, modulePath)
		if err != nil {
			return errors.Wrap(err, "Cannot download remote WASM module from OCI registry")
		}
		defer os.Remove(wasmModulePath)
	}

	wasmEnvKeys, wasmEnvValues := computeWasmEnv()
	wasmWorker, err = chimera.NewWasmWorker(wasmModulePath, wasmEnvKeys, wasmEnvValues)
	if err != nil {
		return err
	}

	return chimeralib.StartServer(
		admissionName,
		admissionHost,
		admissionPort,
		[]chimeralib.Webhook{
			{
				Rules: []admissionregistrationv1.RuleWithOperations{
					{
						Operations: []admissionregistrationv1.OperationType{"*"},
						Rule: admissionregistrationv1.Rule{
							APIGroups:   []string{apiGroups},
							APIVersions: []string{apiVersions},
							Resources:   []string{resources},
						},
					},
				},
				Callback: processRequest,
				Path:     validatePath,
			},
		},
	)
}

func processRequest(admissionReviewRequest *admissionv1.AdmissionRequest) (chimeralib.WebhookResponse, error) {
	admissionReviewRequestBytes, err := json.Marshal(admissionReviewRequest)
	if err != nil {
		return chimeralib.WebhookResponse{}, err
	}

	validationResponse, err := wasmWorker.ProcessRequest(admissionReviewRequestBytes)
	if err != nil {
		return chimeralib.WebhookResponse{}, err
	}

	if !validationResponse.Accepted {
		return chimeralib.NewRejectRequest().WithMessage(validationResponse.Message), nil
	}

	return chimeralib.NewAllowRequest(), nil
}
