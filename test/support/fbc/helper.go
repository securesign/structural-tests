package fbc

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2" //nolint:stylecheck
	. "github.com/onsi/gomega"    //nolint:stylecheck
	"github.com/securesign/structural-tests/test/support"
	"github.com/securesign/structural-tests/test/support/olm"
)

// DescribeFBCImageTests verifies file-based catalog images for the given product.
func DescribeFBCImageTests(product string, defaultsData []byte) bool {
	return Describe("File-based catalog images", Ordered, func() {
		defer GinkgoRecover()

		cfg, err := GetFBCConfig(product, defaultsData)
		Expect(err).NotTo(HaveOccurred(), "failed to load FBC config for product %q", product)

		snapshotData, err := support.ParseSnapshotData()
		Expect(err).NotTo(HaveOccurred())

		var ocps []TableEntry
		for key, snapshotImage := range snapshotData.Images {
			if strings.HasPrefix(key, cfg.ImageKeyPrefix) {
				ocps = append(ocps, Entry(key, key, snapshotImage))
			}
		}
		Expect(ocps).NotTo(BeEmpty(), "no FBC images found with prefix %q in snapshot", cfg.ImageKeyPrefix)

		bundleImageKey := cfg.OLMPackage + "-bundle-image"
		bundleImage := snapshotData.Images[bundleImageKey]

		DescribeTableSubtree("ocp", func(key, fbcImage string) {
			verifyCatalogImage(cfg, key, fbcImage, bundleImage)
		}, ocps)
	})
}

//nolint:funlen
func verifyCatalogImage(cfg FBCConfig, key, fbcImage, bundleImage string) {
	var bundles []olm.Bundle
	var channels []olm.Channel
	var packages []olm.Package
	var deprecation olm.Deprecation

	catalogFileName := filepath.Base(cfg.CatalogPath)

	It("extract "+catalogFileName, func() {
		dir, err := os.MkdirTemp("", key)
		Expect(err).NotTo(HaveOccurred())
		defer os.RemoveAll(dir)

		Expect(support.FileFromImage(context.Background(), fbcImage, cfg.CatalogPath, dir)).To(Succeed())
		file, err := os.Open(dir + "/" + catalogFileName)
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
			Expect(p.Name).To(Equal(cfg.OLMPackage))
			Expect(p.DefaultChannel).To(Equal(cfg.DefaultChannel))
		}
	})

	It("verify channels", func() {
		Expect(channels).To(HaveLen(len(cfg.ExpectedChannels)))

		for _, channel := range channels {
			Expect(channel.Package).To(Equal(cfg.OLMPackage))
			Expect(cfg.ExpectedChannels).To(ContainElement(channel.Name))
		}
	})

	It("contains operator-bundle", func() {
		bundleImageHash := support.ExtractHash(bundleImage)
		exists := false

		for _, bundle := range bundles {
			Expect(bundle.Package).To(Equal(cfg.OLMPackage))
			if bundle.Image == fmt.Sprintf("%s@sha256:%s", cfg.OperatorBundleImage, bundleImageHash) {
				exists = true
			}
		}
		Expect(exists).To(BeTrue(), fmt.Sprintf("olm bundle with %s hash not found", bundleImageHash))
	})

	if len(cfg.ExpectedDeprecations) > 0 {
		It("verify deprecations", func() {
			Expect(deprecation.Entries).To(HaveLen(len(cfg.ExpectedDeprecations)))

			for _, entry := range deprecation.Entries {
				Expect(cfg.ExpectedDeprecations).To(ContainElement(entry.Reference.Name))
			}
		})
	}
}
