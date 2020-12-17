package cmd

import (
	"encoding/json"
	"os"
	"strings"

	chimeralib "github.com/chimera-kube/chimera-admission-library/pkg/chimera"
	"github.com/chimera-kube/chimera-admission/internal/pkg/chimera"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	admissionv1 "k8s.io/api/admission/v1"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
)

func startServer(c *cli.Context) error {
	if wasmUri == "" {
		return errors.New("Please, provide a Wasm URI to load")
	}

	var wasmModulePath string

	moduleSource, modulePath, err := chimera.WasmModuleSource(wasmUri)
	if err != nil {
		return err
	}

	switch moduleSource {
	case chimera.FileSource:
		wasmModulePath = modulePath
	case chimera.HTTPSource, chimera.RegistrySource:
		var err error
		wasmModulePath, err = chimera.FetchRemoteWasmModule(
			moduleSource,
			modulePath,
			wasmRemoteInsecure,
			wasmRemoteNonTLS,
			wasmRemoteCA)
		if err != nil {
			return errors.Wrap(err, "Cannot download remote Wasm module from OCI registry")
		}
		defer os.Remove(wasmModulePath)
	}

	wasmEnvKeys, wasmEnvValues := computeWasmEnv()
	wasmWorker, err = chimera.NewWasmWorker(wasmModulePath, wasmEnvKeys, wasmEnvValues)
	if err != nil {
		return err
	}

	operationTypes := []admissionregistrationv1.OperationType{}
	for _, operation := range operations.Value() {
		switch strings.ToUpper(operation) {
		case "*":
			operationTypes = append(
				operationTypes,
				admissionregistrationv1.OperationAll,
			)
		case "CREATE":
			operationTypes = append(
				operationTypes,
				admissionregistrationv1.Create,
			)
		case "UPDATE":
			operationTypes = append(
				operationTypes,
				admissionregistrationv1.Update,
			)
		case "DELETE":
			operationTypes = append(
				operationTypes,
				admissionregistrationv1.Delete,
			)
		case "CONNECT":
			operationTypes = append(
				operationTypes,
				admissionregistrationv1.Connect,
			)
		default:
			continue
		}
	}

	config := chimeralib.AdmissionConfig{
		Name:          admissionName,
		CallbackHost:  admissionHost,
		CallbackPort:  admissionPort,
		KubeNamespace: kubeNamespace,
		KubeService:   kubeService,
		Webhooks: []chimeralib.Webhook{
			{
				Rules: []admissionregistrationv1.RuleWithOperations{
					{
						Operations: operationTypes,
						Rule: admissionregistrationv1.Rule{
							APIGroups:   apiGroups.Value(),
							APIVersions: apiVersions.Value(),
							Resources:   resources.Value(),
						},
					},
				},
				Callback:      processRequest,
				Path:          validatePath,
				FailurePolicy: admissionregistrationv1.Ignore,
			},
		},
		TLSExtraSANs:              tlsExtraSANs.Value(),
		CertFile:                  certFile,
		KeyFile:                   keyFile,
		CaFile:                    caFile,
		SkipAdmissionRegistration: skipAdmissionRegistration,
	}

	return chimeralib.StartTLSServer(config)
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
