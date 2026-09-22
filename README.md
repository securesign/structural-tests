# Structural tests
Securesign project structural and acceptance tests. Based on
* Securesign releases: https://github.com/securesign/releases
* Securesign operator: https://github.com/securesign/secure-sign-operator
* Securesign Ansible collection: https://github.com/securesign/artifact-signer-ansible
* Policy Controller operator: https://github.com/securesign/policy-controller-operator
* Model Validation operator: https://github.com/securesign/model-validation-operator

## Automation
Current automation is done via Github actions here: https://github.com/securesign/releases/actions/workflows/structural.yml

## Manual testing
It is necessary to point to the release [snapshot](https://github.com/securesign/releases/blob/main/1.1.0/stable/snapshot.json) file. All other components
for the tests are taken from that file, as shown below:

    "operator": {
        "snapshot_name": "operator-v1-1-4x2vj",
        "rhtas-operator-image": "quay.io/securesign/rhtas-operator-v1-1@sha256:3a61aca9fa8ed6580a367bc08a45cc27fc7f50ff24e786ffde9ec3d9c549b00b",
        "rhtas-operator-bundle-image": "quay.io/securesign/rhtas-operator-bundle-v1-1@sha256:6db817ed76948417f358d402e737df7b320f82462ad164b002ded15e560a0fdf"
    },

    "artifact-signer-ansible": {
        "collection": {
            "url": "https://github.com/securesign/artifact-signer-ansible/actions/runs/11705765669/artifacts/2152648141",
            "sha256": "4da3d330f9e82a65d93b242e0cc14b5912d4bf65d0eac31fe1d226e4c6ae11f5"
        }
    }

### Parameters
* ``SNAPSHOT`` - points to the ``snapshot.json`` file, can be local or on a server (github).
* ``VERSION`` - version of realease in semver format. Example ``1.2.0``
* ``TEST_GITHUB_TOKEN`` - token used to access  ``releases`` project on github.
* ``REPOSITORIES`` - file with images published in ``registry.redhat.io``, default ``testdata/repositories.json``. For how to get or update this file, 
  check [Repository List](#repository-list) chapter.

### Examples
Run tests based on a github file:

    SNAPSHOT=https://raw.githubusercontent.com/securesign/releases/refs/heads/feat/release-1.1.1/1.1.1/stable/snapshot.json \
    TEST_GITHUB_TOKEN=ghp_Ae \
    go test -v ./test/acceptance/rhtas/... --ginkgo.v

Run the same tests on a local (cloned) file:

    SNAPSHOT=../releases/1.1.1/stable/snapshot.json \
    go test -v ./test/acceptance/rhtas/... --ginkgo.v

To run just individual test use ``--ginkgo.fokus-file`` parameter:

    SNAPSHOT=../releases/1.1.1/stable/snapshot.json \
    go test -v ./test/acceptance/rhtas/... --ginkgo.v --ginkgo.focus-file "ansible"

To run policy controller operator tests use:
```
go test -v ./test/acceptance/policy_controller/... --ginkgo.v
```

To run model validation operator tests use:
```
go test -v ./test/acceptance/model_transparency/... --ginkgo.v
```

## CLI stack configuration

CLI stack checks run inside the `rhtas`, `model_transparency`, and
`policy_controller` acceptance suites. Each product owns its expected inventory:

```yaml
rhtas:
  cliStack:
    images:
      - imageKey: tufcli-cli-stack-image
        binaries:
          - {path: /binaries/tufcli_linux_amd64.tar.gz, os: linux, arch: amd64}
```

Set `TEST_CONFIG` to a local or remote YAML file, as for operator and FBC tests.
The example above replaces the complete default inventory with one archive;
list every image and archive the release must publish. Missing configured images
or archives fail the tests. Invalid CLI configuration also fails. Archive paths
must be under `/binaries/`; both `.tar.gz` and `.exe.gz` packages are read as tar
archives. Executable format and architecture are checked without running binaries.

Omitting `cliStack` or its `images` field inherits the product defaults.
`cliStack: {images: []}` explicitly disables CLI stack checks. An unwrapped
`cliStack` section applies only to RHTAS. Product sections remain independent.

RHTAS defaults describe 1.5.x (including tufcli). The 1.4.x override retains tuftool
and model-transparency archives. Older example configurations disable CLI stack
checks. Model Transparency and Policy Controller default to empty inventories;
their release YAML can supply their own lists.

```sh
VERSION=1.5.0 SNAPSHOT=/path/to/1.5.0/snapshot.json \
  go test -v ./test/acceptance/rhtas --ginkgo.focus='CLI Stack Images'

VERSION=1.4.3 SNAPSHOT=/path/to/1.4.3/snapshot.json \
  TEST_CONFIG=testdata/testconfig-1.4.yaml \
  go test -v ./test/acceptance/rhtas --ginkgo.focus='CLI Stack Images'
```

Relative `TEST_CONFIG` paths resolve from this repository's root. The 1.4.x
example overrides only CLI checks; combine its `cliStack`
section with the release's operator, Ansible, and FBC overrides for a full run.
Client-server checks retain the legacy tuftool comparisons and skip RHTAS 1.5+.
Set `VERSION` when the snapshot path does not contain a release version; existing
version resolution treats an unspecified version as the latest release.

Release CI migration (in `securesign/releases`, separately): remove the standalone
`cli_stack` package selection/`CLI_STACK` flag, run the appropriate product suite,
and supply that release's inventory in `testingResources/structural-tests.yaml`.
Use explicit empty inventories for releases without CLI stacks. Existing FBC suite
registration reads the snapshot even when CLI checks are focused, so these examples
use complete product snapshots.

## Repository List
The [repositories.json](testdata/repositories.json) file is used to check of all images are published correctly. To pull the list of repositories from Pyxis API:

```bash
curl --negotiate -u : -b .cookiejar.txt -c .cookiejar.txt 'https://pyxis.engineering.redhat.com/v1/product-listings/id/6604180e80e2fa3e4947d1d5/repositories?filter=release_categories%3Din%3D%28%22Generally%20Available%22%29&include=data.repository,data._id,data.published' | jq > testdata/repositories.json
```

## Ansible Artifacts
Published Ansible collections are also stored as an zip [artifacts](https://github.com/securesign/artifact-signer-ansible/actions/workflows/collection-build.yaml).
To download list of available artifacts:

    curl -L \
    -H "Accept: application/vnd.github+json" \
    -H "Authorization: Bearer ghp_Ae" \
    -H "X-GitHub-Api-Version: 2022-11-28" \
    https://api.github.com/repos/securesign/artifact-signer-ansible/actions/artifacts

Downloading one artifact:

    curl -L -O \
    -H "Accept: application/vnd.github+json" \
    -H "Authorization: Bearer ghp_Ae" \
    -H "X-GitHub-Api-Version: 2022-11-28" \
    https://api.github.com/repos/securesign/artifact-signer-ansible/actions/artifacts/2442056100/zip
