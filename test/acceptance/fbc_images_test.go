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

	snapshotImages, err := support.ParseSnapshotImages()
	Expect(err).NotTo(HaveOccurred())
	for key, snapshotImage := range snapshotImages {
		if strings.Index(key, "fbc-") == 0 {
			ocps = append(ocps, Entry(key, key, snapshotImage))
		}
	}
	Expect(ocps).NotTo(BeEmpty())

	bundleImage = snapshotImages[support.OperatorBundleImageKey]

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
				snapshotImages, err := support.ParseSnapshotImages()
				Expect(err).NotTo(HaveOccurred())
				Expect(snapshotImages).NotTo(BeEmpty())

			})

			It("verify package", func() {
				for _, p := range packages {
					Expect(p.Name).To(Equal(olmPackage))
					Expect(p.DefaultChannel).To(Equal("stable"))
				}
			})

			It("verify channels", func() {
				expectedChannels := []string{"stable", "stable-v1.0", "stable-v1.1"}
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

		},
		ocps)
})
