package acceptance

import (
	"context"
	"fmt"
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/securesign/structural-tests/test/support"
	"github.com/securesign/structural-tests/test/support/olm"
)

const (
	olmPackage          = "rhtas-operator"
	operatorBundleImage = "registry.redhat.io/rhtas/rhtas-operator-bundle"
	fbcCatalogFileName  = "catalog.json"
	fbcCatalogPath      = "/configs/rhtas-operator/" + fbcCatalogFileName
)

var _ = Describe("File-based catalog images", Ordered, func() {

	defer GinkgoRecover()
	var ocps []TableEntry
	var bundleImage string

	snapshotData, err := support.ParseSnapshotData()
	Expect(err).NotTo(HaveOccurred())
	for key, snapshotImage := range snapshotData.Images {
		if strings.Index(key, "fbc-") == 0 {
			ocps = append(ocps, Entry(key, key, snapshotImage))
		}
	}
	Expect(ocps).NotTo(BeEmpty())

	bundleImage = snapshotData.Images[support.OperatorBundleImageKey]

	DescribeTableSubtree("ocp",
		func(key, fbcImage string) {

			var bundles []olm.Bundle
			var channels []olm.Channel
			var packages []olm.Package
			var deprecation olm.Deprecation

			It("extract catalog.json", func() {
				dir, err := os.MkdirTemp("", key)
				Expect(err).NotTo(HaveOccurred())
				defer os.RemoveAll(dir)

				Expect(support.FileFromImage(context.Background(), fbcImage, fbcCatalogPath, dir)).To(Succeed())
				file, err := os.Open(dir + "/" + fbcCatalogFileName)
				Expect(err).NotTo(HaveOccurred())
				defer file.Close()

				catalog, err := olm.ParseCatalogJSON(file)
				Expect(err).NotTo(HaveOccurred())
				Expect(catalog).NotTo(BeNil())

				for _, obj := range catalog {
					switch typedObj := obj.(type) {
					case olm.Bundle:
						bundles = append(bundles, typedObj)
					case olm.Channel:
						channels = append(channels, typedObj)
					case olm.Package:
						packages = append(packages, typedObj)
					case olm.Deprecation:
						deprecation = typedObj
					}
				}

				Expect(bundles).ToNot(BeEmpty())
				Expect(channels).ToNot(BeEmpty())
				Expect(packages).ToNot(BeEmpty())
			})

			It("extract bundle-image from snapshot.json", func() {
				snapshotData, err := support.ParseSnapshotData()
				Expect(err).NotTo(HaveOccurred())
				Expect(snapshotData.Images).NotTo(BeEmpty())

			})

			It("verify package", func() {
				for _, p := range packages {
					Expect(p.Name).To(Equal(olmPackage))
					Expect(p.DefaultChannel).To(Equal("stable"))
				}
			})

			It("verify channels", func() {
				expectedChannels := []string{"stable", "stable-v1.1", "stable-v1.2", "stable-v1.3"}
				Expect(channels).To(HaveLen(len(expectedChannels)))

				for _, channel := range channels {
					Expect(channel.Package).To(Equal(olmPackage))
					Expect(expectedChannels).To(ContainElement(channel.Name))
				}
			})

			It("contains operator-bundle", func() {
				bundleImageHash := support.ExtractHash(bundleImage)
				Expect(bundleImageHash).NotTo(BeEmpty(), "rhtas-operator-bundle-image in snapshot missing or does not have a hash")
				exists := false

				for _, bundle := range bundles {
					Expect(bundle.Package).To(Equal(olmPackage))
					if bundle.Image == fmt.Sprintf("%s@sha256:%s", operatorBundleImage, bundleImageHash) {
						exists = true
					}
				}
				Expect(exists).To(BeTrue(), fmt.Sprintf("olm bundle with %s hash not found", bundleImageHash))
			})

			It("verify deprecations", func() {
				expectedDeprecations := []string{"stable-v1.1", "rhtas-operator.v1.1.0", "rhtas-operator.v1.1.1", "rhtas-operator.v1.1.2"}
				Expect(deprecation.Entries).To(HaveLen(len(expectedDeprecations)))

				for _, entry := range deprecation.Entries {
					Expect(expectedDeprecations).To(ContainElement(entry.Reference.Name))
				}
			})
		},
		ocps)
})
