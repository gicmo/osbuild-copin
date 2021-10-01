package main

import (
	"context"
	"fmt"
	"os"

	"github.com/containers/image/v5/image"
	"github.com/containers/image/v5/manifest"
	"github.com/containers/image/v5/transports/alltransports"
	"github.com/containers/image/v5/types"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

var defaultUserAgent = "osbuild-depsolve/1.0"

func main() {

	if len(os.Args) < 2 {
		fmt.Printf("usage: %s DOCKER-REFERENCE\n", os.Args[0])
		os.Exit(1)
	}

	dockerRef := fmt.Sprintf("docker://%s", os.Args[1])

	ctx := context.Background()

	sys := &types.SystemContext{
		RegistriesDirPath:        "",
		ArchitectureChoice:       "amd64",
		OSChoice:                 "linux",
		VariantChoice:            "",
		SystemRegistriesConfPath: "",
		BigFilesTemporaryDir:     "/var/tmp",
		DockerRegistryUserAgent:  defaultUserAgent,
	}

	ref, err := alltransports.ParseImageName(dockerRef)
	if err != nil {
		panic(err)
	}

	src, err := ref.NewImageSource(ctx, sys)
	if err != nil {
		panic(err)
	}

	rawManifest, mt, err := src.GetManifest(ctx, nil)
	if err != nil {
		panic(err)
	}

	ml, err := manifest.Schema2ListFromManifest(rawManifest)
	if err != nil {
		panic(err)
	}

	digest, err := manifest.Digest(rawManifest)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Manifest: %#v (%s)\n", digest, mt)

	digest = ""

	for _, manifest := range ml.Manifests {

		if manifest.Platform.Architecture != sys.ArchitectureChoice {
			continue
		}

		if manifest.Platform.OS != sys.OSChoice {
			continue
		}

		digest = manifest.Digest
		mt = manifest.MediaType
		break
	}

	if digest == "" {
		panic("Could not find matching manifest")
	}

	fmt.Printf("Manifest: %#v (%s)\n", digest, mt)

	img, err := image.FromUnparsedImage(ctx, sys, image.UnparsedInstance(src, nil))
	if err != nil {
		panic(err)
	}

	uo := types.ManifestUpdateOptions{
		ManifestMIMEType: v1.MediaTypeImageManifest,
	}

	updated, err := img.UpdatedImage(ctx, uo)
	if err != nil {
		panic(err)
	}

	//fmt.Printf("%#v\n", updated)

	raw, mt, err := updated.Manifest(ctx)
	if err != nil {
		panic(err)
	}

	//fmt.Printf("%#v\n", raw)
	//fmt.Printf("%#v\n", m)

	digest, err = manifest.Digest(raw)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Manifest: %#v (%s)\n", digest, mt)
}
