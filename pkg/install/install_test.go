package install

import (
	"testing"

	"github.com/instrumenta/kubeval/kubeval"
	"github.com/stretchr/testify/assert"
)

func testFillInTemplates(t *testing.T, expectedManifestCount int, params TemplateParameters) map[string][]byte {
	manifests, err := FillInTemplates(params)
	assert.NoError(t, err)
	assert.Len(t, manifests, expectedManifestCount)
	for fileName, contents := range manifests {
		validationResults, err := kubeval.Validate(contents)
		assert.NoError(t, err)
		for _, result := range validationResults {
			if len(result.Errors) > 0 {
				t.Errorf("found problems with manifest %s (Kind %s):\ncontent:\n%s\nerrors: %s",
					fileName,
					result.Kind,
					string(contents),
					result.Errors)
			}
		}
	}
	return manifests
}

func TestFillInTemplatesAllParameters(t *testing.T) {
	testFillInTemplates(t, 6, TemplateParameters{
		GitURL:             "git@github.com:fluxcd/flux-get-started",
		GitBranch:          "branch",
		GitPaths:           []string{"dir1", "dir2"},
		GitLabel:           "label",
		GitUser:            "User",
		GitEmail:           "this.is@anemail.com",
		Namespace:          "flux",
		GitReadOnly:        false,
		ManifestGeneration: true,
		AdditionalFluxArgs: []string{"arg1=foo", "arg2=bar"},
		RegistryScanning:   true,
	})
}

func TestFillInTemplatesMissingValues(t *testing.T) {
	testFillInTemplates(t, 6, TemplateParameters{
		GitURL:           "git@github.com:fluxcd/flux-get-started",
		GitBranch:        "branch",
		GitPaths:         []string{},
		GitLabel:         "label",
		RegistryScanning: true,
	})
}

func TestFillInTemplatesNoMemcached(t *testing.T) {
	testFillInTemplates(t, 4, TemplateParameters{
		GitURL:           "git@github.com:fluxcd/flux-get-started",
		GitBranch:        "branch",
		GitPaths:         []string{},
		GitLabel:         "label",
		RegistryScanning: false,
	})
	testFillInTemplates(t, 4, TemplateParameters{
		GitURL:      "git@github.com:fluxcd/flux-get-started",
		GitBranch:   "branch",
		GitPaths:    []string{},
		GitLabel:    "label",
		GitReadOnly: false,
	})
}

func TestFillInTemplatesConfigFile(t *testing.T) {

	configContent := `config1: configuration1
config2: configuration2
config3: configuration3`

	tests := map[string]struct {
		params              TemplateParameters
		configFileCheck     string
		deploymentFileCheck string
	}{
		"configMap": {
			params: TemplateParameters{
				GitURL:             "git@github.com:fluxcd/flux-get-started",
				GitBranch:          "branch",
				GitPaths:           []string{"dir1", "dir2"},
				GitLabel:           "label",
				GitUser:            "User",
				GitEmail:           "this.is@anemail.com",
				Namespace:          "flux",
				ConfigAsConfigMap:  true,
				AdditionalFluxArgs: []string{"arg1=foo", "arg2=bar"},
			},
			configFileCheck:     "    config2: configuration2",
			deploymentFileCheck: "name: flux-config",
		},
		"secret": {
			params: TemplateParameters{
				GitURL:             "git@github.com:fluxcd/flux-get-started",
				GitBranch:          "branch",
				GitPaths:           []string{"dir1", "dir2"},
				GitLabel:           "label",
				GitUser:            "User",
				GitEmail:           "this.is@anemail.com",
				Namespace:          "flux",
				ConfigAsConfigMap:  false,
				AdditionalFluxArgs: []string{"arg1=foo", "arg2=bar"},
			},
			// the following field value is the base64 encoding of the config file string above
			configFileCheck:     `  flux-config.yaml: "Y29uZmlnMTogY29uZmlndXJhdGlvbjEKY29uZmlnMjogY29uZmlndXJhdGlvbjIKY29uZmlnMzogY29uZmlndXJhdGlvbjM="`,
			deploymentFileCheck: "secretName: flux-config",
		},
	}

	for name, test := range tests {
		t.Run(name, func(*testing.T) {
			test.params.ConfigFileContent = configContent
			manifests := testFillInTemplates(t, 4, test.params)
			for fileName, contents := range manifests {
				if fileName == "flux-config.yaml" {
					assert.Contains(t, string(contents), test.configFileCheck)
				}
				if fileName == "flux-deployment.yaml" {
					assert.Contains(t, string(contents), test.deploymentFileCheck)
				}
			}
		})
	}
}
