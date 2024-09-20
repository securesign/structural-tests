# structural-tests
Securesign project structural and acceptance tests.


## Repository list
Pull list of repositories from Pyxis API

```bash
curl --negotiate -u : -b .cookiejar.txt -c .cookiejar.txt 'https://pyxis.engineering.redhat.com/v1/product-listings/id/6604180e80e2fa3e4947d1d5/repositories?filter=release_categories%3Din%3D%28%22Generally%20Available%22%29&include=data.repository,data._id,data.published' | jq > testdata/repositories.json
```
