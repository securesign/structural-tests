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
	olmPackage          = "policy-controller-operator"
	operatorBundleImage = "registry.redhat.io/rhtas/policy-controller-operator-bundle"
	fbcCatalogFileName  = "catalog.json"
	fbcCatalogPath      = "/configs/policy-controller-operator/catalog.json"
	defaultChannel      = "tech-preview"
)

var expectedChannels = []string{"tech-preview"}

var _ = Describe("File-based catalog images", Ordered, func() {

	defer GinkgoRecover()
	var ocps []TableEntry
	var bundleImage string

	snapshotData, err := support.ParseSnapshotData()
	Expect(err).NotTo(HaveOccurred())
	for key, snapshotImage := range snapshotData.Images {
		if strings.Index(key, "pco-fbc-") == 0 {
			ocps = append(ocps, Entry(key, key, snapshotImage))
		}
	}
	Expect(ocps).NotTo(BeEmpty())

	bundleImage = snapshotData.Images[support.PolicyControllerOperatorBundleImageKey]

	DescribeTableSubtree("ocp",
		func(key, fbcImage string) {

			var bundles []olm.Bundle
			var channels []olm.Channel
			var packages []olm.Package

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
					Expect(p.DefaultChannel).To(Equal(defaultChannel))
				}
			})

			It("verify channels", func() {
				Expect(channels).To(HaveLen(len(expectedChannels)))

				for _, channel := range channels {
					Expect(channel.Package).To(Equal(olmPackage))
					Expect(expectedChannels).To(ContainElement(channel.Name))
				}
			})

			It("contains operator-bundle", func() {
				bundleImageHash := support.ExtractHash(bundleImage)
				exists := false

				for _, bundle := range bundles {
					Expect(bundle.Package).To(Equal(olmPackage))
					if bundle.Image == fmt.Sprintf("%s@sha256:%s", operatorBundleImage, bundleImageHash) {
						exists = true
					}
				}
				Expect(exists).To(BeTrue(), fmt.Sprintf("olm bundle with %s hash not found", bundleImageHash))
			})
		}, ocps)
})
