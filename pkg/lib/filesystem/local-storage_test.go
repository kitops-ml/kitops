package filesystem

import (
	"testing"

	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/constants/mediatype"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateManifestAddsCreatedAnnotationForModelPack(t *testing.T) {
	configDesc := ocispec.Descriptor{
		MediaType: mediatype.ModelPackConfigMediaType.String(),
		Digest:    digest.FromString("config"),
		Size:      42,
	}
	layerDescs := []ocispec.Descriptor{{
		MediaType: mediatype.New(mediatype.ModelPackFormat, mediatype.ModelBaseType, mediatype.TarFormat, mediatype.GzipCompression).String(),
		Digest:    digest.FromString("layer"),
		Size:      128,
	}}

	manifest, err := createManifest(
		configDesc,
		layerDescs,
		mediatype.ModelPackFormat,
		"2026-04-16T12:34:56Z",
	)
	require.NoError(t, err)

	assert.Equal(t, constants.Version, manifest.Annotations[constants.CliVersionAnnotation])
	assert.Equal(t, "2026-04-16T12:34:56Z", manifest.Annotations[constants.OciImageCreatedAnnotation])
}

func TestCreateManifestSkipsCreatedAnnotationForKitFormat(t *testing.T) {
	configDesc := ocispec.Descriptor{
		MediaType: mediatype.KitConfigMediaType.String(),
		Digest:    digest.FromString("config"),
		Size:      42,
	}
	layerDescs := []ocispec.Descriptor{{
		MediaType: mediatype.New(mediatype.KitFormat, mediatype.ModelBaseType, mediatype.TarFormat, mediatype.GzipCompression).String(),
		Digest:    digest.FromString("layer"),
		Size:      128,
	}}

	manifest, err := createManifest(
		configDesc,
		layerDescs,
		mediatype.KitFormat,
		"2026-04-16T12:34:56Z",
	)
	require.NoError(t, err)

	assert.Equal(t, constants.Version, manifest.Annotations[constants.CliVersionAnnotation])
	_, exists := manifest.Annotations[constants.OciImageCreatedAnnotation]
	assert.False(t, exists)
}
