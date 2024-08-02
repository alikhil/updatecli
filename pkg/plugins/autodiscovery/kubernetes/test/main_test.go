package fleet

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/autodiscovery/kubernetes"
)

func TestDiscoverManifests(t *testing.T) {
	// Disable condition testing with running short test
	if testing.Short() {
		return
	}

	testdata := []struct {
		name              string
		rootDir           string
		digest            bool
		expectedPipelines []string
	}{
		{
			name:    "Scenario 1",
			rootDir: "testdata/success",
			expectedPipelines: []string{`name: 'deps: bump container image "updatecli"'
sources:
  updatecli:
    name: 'get latest container image tag for "ghcr.io/updatecli/updatecli"'
    kind: 'dockerimage'
    spec:
      image: 'ghcr.io/updatecli/updatecli'
      tagfilter: '^v\d*(\.\d*){2}$'
      versionfilter:
        kind: 'semver'
        pattern: '>=v0.67.0'
targets:
  updatecli:
    name: 'deps: bump container image "ghcr.io/updatecli/updatecli" to {{ source "updatecli" }}'
    kind: 'yaml'
    spec:
      file: 'pod.yaml'
      key: '$.spec.containers[0].image'
    sourceid: 'updatecli'
    transformers:
      - addprefix: 'ghcr.io/updatecli/updatecli:'
`},
		},
		{
			name:    "Scenario 2 - Kustomize",
			rootDir: "testdata/kustomize",
			expectedPipelines: []string{`name: 'deps: bump container image "nginx"'
sources:
  nginx:
    name: 'get latest container image tag for "nginx"'
    kind: 'dockerimage'
    spec:
      image: 'nginx'
      tagfilter: '^\d*(\.\d*){2}$'
      versionfilter:
        kind: 'semver'
        pattern: '>=1.20.0'
targets:
  nginx:
    name: 'deps: bump container image "nginx" to {{ source "nginx" }}'
    kind: 'yaml'
    spec:
      file: 'deployment.yaml'
      key: '$.spec.template.spec.containers[0].image'
    sourceid: 'nginx'
    transformers:
      - addprefix: 'nginx:'
`},
		},
		{
			name:    "Scenario - latest and digest",
			rootDir: "testdata/success",
			digest:  true,
			expectedPipelines: []string{`name: 'deps: bump container image "updatecli"'
sources:
  updatecli:
    name: 'get latest container image tag for "ghcr.io/updatecli/updatecli"'
    kind: 'dockerimage'
    spec:
      image: 'ghcr.io/updatecli/updatecli'
      tagfilter: '^v\d*(\.\d*){2}$'
      versionfilter:
        kind: 'semver'
        pattern: '>=v0.67.0'
  updatecli-digest:
    name: 'get latest container image digest for "ghcr.io/updatecli/updatecli:v0.67.0"'
    kind: 'dockerdigest'
    spec:
      image: 'ghcr.io/updatecli/updatecli'
      tag: '{{ source "updatecli" }}'
    dependson:
      - 'updatecli'
targets:
  updatecli:
    name: 'deps: bump container image digest for "ghcr.io/updatecli/updatecli:v0.67.0"'
    kind: 'yaml'
    spec:
      file: 'pod.yaml'
      key: '$.spec.containers[0].image'
    sourceid: 'updatecli-digest'
    transformers:
      - addprefix: 'ghcr.io/updatecli/updatecli:'
`},
		},
	}

	for _, tt := range testdata {

		t.Run(tt.name, func(t *testing.T) {
			digest := tt.digest
			pod, err := kubernetes.New(
				kubernetes.Spec{
					Digest: &digest,
				}, tt.rootDir, "")

			require.NoError(t, err)

			var pipelines []string
			rawPipelines, err := pod.DiscoverManifests()
			require.NoError(t, err)

			if len(rawPipelines) == 0 {
				t.Errorf("No pipelines found for %s", tt.name)
			}

			for i := range rawPipelines {
				// We expect manifest generated by the autodiscovery to use the yaml syntax
				pipelines = append(pipelines, string(rawPipelines[i]))
				assert.Equal(t, tt.expectedPipelines[i], pipelines[i])
			}
		})
	}

}
