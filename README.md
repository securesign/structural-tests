# Structural tests
Securesign project structural and acceptance tests. Base on
* Securesign releases: https://github.com/securesign/releases
* Securesign operator: https://github.com/securesign/secure-sign-operator
* Securesign Ansible collection: https://github.com/securesign/artifact-signer-ansible

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
* ``TEST_GITHUB_TOKEN`` - token used to access  ``releases`` project on github.
* ``ANSIBLE`` - ansible collection zip file used instead of the one defined in ``snapshot.json`` file. Can also be local.
* ``REPOSITORIES`` - file with images published in ``registry.redhat.io``, default ``testdata/repositories.json``. For how to get or update this file, 
  check [Repository List](#repository-list) chapter.

### Examples
Run tests based on a github file:

    SNAPSHOT=https://raw.githubusercontent.com/securesign/releases/refs/heads/feat/release-1.1.1/1.1.1/stable/snapshot.json \
    TEST_GITHUB_TOKEN=ghp_Ae \
    go test -v ./test/... --ginkgo.v

Run the same tests on a local (cloned) file:

    SNAPSHOT=../releases/1.1.1/stable/snapshot.json \
    go test -v ./test/... --ginkgo.v

Force different ansible collection instead of the one defined in ``snapshot.json`` file. This may be useful, when checking ansible collection not yet published:

    SNAPSHOT=../releases/1.1.1/stable/snapshot.json \
    ANSIBLE=https://api.github.com/repos/securesign/artifact-signer-ansible/actions/artifacts/2442056100/zip \
    go test -v ./test/... --ginkgo.v

To run just individual test use ``--ginkgo.fokus-file`` parameter:

    SNAPSHOT=../releases/1.1.1/stable/snapshot.json \
    go test -v ./test/... --ginkgo.v --ginkgo.focus-file "ansible"

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