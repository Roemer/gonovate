package datasources

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeDockerIo(t *testing.T) {
	assert := assert.New(t)

	// Normal replacements
	assert.Equal("index.docker.io", normalizeDockerIo("docker.io"))
	assert.Equal("index.docker.io", normalizeDockerIo("registry-1.docker.io"))
	assert.Equal("index.docker.io", normalizeDockerIo("index.docker.io"))
	assert.Equal("index.docker.io/test", normalizeDockerIo("docker.io/test"))
	assert.Equal("index.docker.io/test", normalizeDockerIo("registry-1.docker.io/test"))
	assert.Equal("index.docker.io/test", normalizeDockerIo("index.docker.io/test"))

	// Replacements with Schema
	assert.Equal("https://index.docker.io", normalizeDockerIo("https://docker.io"))
	assert.Equal("https://index.docker.io", normalizeDockerIo("https://registry-1.docker.io"))
	assert.Equal("https://index.docker.io", normalizeDockerIo("https://index.docker.io"))
	assert.Equal("https://index.docker.io/test", normalizeDockerIo("https://docker.io/test"))
	assert.Equal("https://index.docker.io/test", normalizeDockerIo("https://registry-1.docker.io/test"))
	assert.Equal("https://index.docker.io/test", normalizeDockerIo("https://index.docker.io/test"))

	// Invalid replacements
	assert.Equal("docker.io.my.domain", normalizeDockerIo("docker.io.my.domain"))
	assert.Equal("registry-1.docker.io.my.domain", normalizeDockerIo("registry-1.docker.io.my.domain"))
	assert.Equal("index.docker.io.my.domain", normalizeDockerIo("index.docker.io.my.domain"))
	assert.Equal("docker.io.my.domain/test", normalizeDockerIo("docker.io.my.domain/test"))
	assert.Equal("registry-1.docker.io.my.domain/test", normalizeDockerIo("registry-1.docker.io.my.domain/test"))
	assert.Equal("index.docker.io.my.domain/test", normalizeDockerIo("index.docker.io.my.domain/test"))

	// Invalid replacements with Schema
	assert.Equal("https://docker.io.my.domain", normalizeDockerIo("https://docker.io.my.domain"))
	assert.Equal("https://registry-1.docker.io.my.domain", normalizeDockerIo("https://registry-1.docker.io.my.domain"))
	assert.Equal("https://index.docker.io.my.domain", normalizeDockerIo("https://index.docker.io.my.domain"))
	assert.Equal("https://docker.io.my.domain/test", normalizeDockerIo("https://docker.io.my.domain/test"))
	assert.Equal("https://registry-1.docker.io.my.domain/test", normalizeDockerIo("https://registry-1.docker.io.my.domain/test"))
	assert.Equal("https://index.docker.io.my.domain/test", normalizeDockerIo("https://index.docker.io.my.domain/test"))
}

func TestGetRegistryLocal(t *testing.T) {
	assert := assert.New(t)

	// Try all possible variants for scheme, port and suffixes
	for _, scheme := range []string{"", "http://", "https://"} {
		for _, host := range []string{"registry", "registry.local"} {
			for _, port := range []string{"", ":5000"} {
				for _, suffix := range []string{"", "/org"} {
					inputPackageName := fmt.Sprintf("%s%s/org/image", host, port)
					inputRegistryUrl := fmt.Sprintf("%s%s%s%s", scheme, host, port, suffix)
					fmt.Printf("Testing: %s - %s\n", inputPackageName, inputRegistryUrl)
					registryUrl, imagePath, err := getDockerRegistry(inputPackageName, inputRegistryUrl)
					assert.NoError(err)
					if scheme == "http://" {
						assert.Equal(fmt.Sprintf("http://%s%s", host, port), registryUrl, "RegistryUrl is not equal")
					} else {
						assert.Equal(fmt.Sprintf("https://%s%s", host, port), registryUrl, "RegistryUrl is not equal")
					}
					assert.Equal("org/image", imagePath, "ImagePath is not equal")
				}
			}
		}
	}
}

func TestGetRegistryLocalWithMismatchingRegistryUrl(t *testing.T) {
	assert := assert.New(t)

	for _, host := range []string{"registry:5000", "registry.local"} {
		for _, inputRegistryUrl := range []string{"", "https://index.docker.io", "registry"} {
			inputPackageName := fmt.Sprintf("%s/org/package", host)
			fmt.Printf("Testing: %s - %s\n", inputPackageName, inputRegistryUrl)
			registryUrl, imagePath, err := getDockerRegistry(inputPackageName, inputRegistryUrl)
			assert.NoError(err)
			assert.Equal(fmt.Sprintf("https://%s", host), registryUrl, "RegistryUrl is not equal")
			assert.Equal("org/package", imagePath, "ImagePath is not equal")
		}
	}
}

func TestGetRegistryCustom(t *testing.T) {
	assert := assert.New(t)

	host, path, err := getDockerRegistry("my.local.registry/prefix/image", "https://my.local.registry/prefix")
	assert.NoError(err)
	assert.Equal("https://my.local.registry", host)
	assert.Equal("prefix/image", path)
}

func TestGetRegistryLocal2(t *testing.T) {
	assert := assert.New(t)

	host, path, err := getDockerRegistry("registry:5000/org/package", "https://index.docker.io")
	assert.NoError(err)
	assert.Equal("https://registry:5000", host)
	assert.Equal("org/package", path)
}

func TestGetRegistryHttp(t *testing.T) {
	assert := assert.New(t)

	host, path, err := getDockerRegistry("my.local.registry/prefix/image", "http://my.local.registry/prefix")
	assert.NoError(err)
	assert.Equal("http://my.local.registry", host)
	assert.Equal("prefix/image", path)
}

func TestGetRegistryWithoutSchema(t *testing.T) {
	assert := assert.New(t)

	host, path, err := getDockerRegistry("my.local.registry/prefix/image", "my.local.registry/prefix")
	assert.NoError(err)
	assert.Equal("https://my.local.registry", host)
	assert.Equal("prefix/image", path)
}

func TestVaria(t *testing.T) {
	assert := assert.New(t)

	testCases := []testCase{
		{
			inputPackageName: "strimzi-kafka-operator",
			inputRegistryUrl: "https://quay.io/strimzi-helm/",
			expectedPath:     "strimzi-helm/strimzi-kafka-operator",
			expectedHost:     "https://quay.io",
		},
		{
			inputPackageName: "strimzi-kafka-operator",
			inputRegistryUrl: "https://docker.io/strimzi-helm/",
			expectedPath:     "strimzi-helm/strimzi-kafka-operator",
			expectedHost:     "https://index.docker.io",
		},
		{
			inputPackageName: "nginx",
			inputRegistryUrl: "https://docker.io",
			expectedPath:     "library/nginx",
			expectedHost:     "https://index.docker.io",
		},
		{
			inputPackageName: "registry-1.docker.io/bitnamicharts/cert-manager",
			inputRegistryUrl: "https://index.docker.io",
			expectedPath:     "bitnamicharts/cert-manager",
			expectedHost:     "https://index.docker.io",
		},
	}
	runAndValidateTestCases(assert, testCases)
}

func runAndValidateTestCases(assert *assert.Assertions, testCases []testCase) {
	for _, tc := range testCases {
		host, path, err := getDockerRegistry(tc.inputPackageName, tc.inputRegistryUrl)
		assert.NoError(err)
		assert.Equal(tc.expectedHost, host)
		assert.Equal(tc.expectedPath, path)
	}
}

type testCase struct {
	inputPackageName string
	inputRegistryUrl string
	expectedHost     string
	expectedPath     string
}
