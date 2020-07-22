package cert

import (
	"github.com/yuyicai/kubei/config/rundata"
	"k8s.io/apimachinery/pkg/runtime"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	clientcmdlatest "k8s.io/client-go/tools/clientcmd/api/latest"
)

func SendCert(c *rundata.Cluster) error {

	return nil
}

// EncodeKubeConfig serializes the config to yaml.
// Encapsulates serialization without assuming the destination is a file.
func EncodeKubeConfig(config clientcmdapi.Config) ([]byte, error) {
	return runtime.Encode(clientcmdlatest.Codec, &config)
}
