// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"crypto/tls"
	"encoding/json"
	"fmt"

	"github.com/pingcap-incubator/tiops/pkg/utils"
	pdserverapi "github.com/pingcap/pd/v4/server/api"
)

// PDClient is an HTTP client of the PD server "Host"
type PDClient struct {
	Host       string
	tlsEnabled bool
	httpClient *utils.HTTPClient
}

// NewPDClient returns a new PDClient
func NewPDClient(host string, timeout int, tlsConfig *tls.Config) *PDClient {
	enableTls := false
	if tlsConfig != nil {
		enableTls = true
	}

	return &PDClient{
		Host:       host,
		tlsEnabled: enableTls,
		httpClient: utils.NewHTTPClient(timeout, tlsConfig),
	}
}

// GetURL builds the the client URL of PDClient
func (pc *PDClient) GetURL() string {
	httpPrefix := "http"
	if pc.tlsEnabled {
		httpPrefix = "https"
	}
	return fmt.Sprintf("%s://%s", httpPrefix, pc.Host)
}

var (
	pdHealthURI         = "pd/health"
	pdMembersURI        = "pd/api/v1/members"
	pdStoresURI         = "pd/api/v1/stores"
	pdStoreURI          = "pd/api/v1/store"
	pdConfigURI         = "pd/api/v1/config"
	pdClusterIDURI      = "pd/api/v1/cluster"
	pdSchedulersURI     = "pd/api/v1/schedulers"
	pdLeaderURI         = "pd/api/v1/leader"
	pdLeaderTransferURI = "pd/api/v1/leader/transfer"
)

// PDHealthInfo is the member health info from PD's API
type PDHealthInfo struct {
	Healths []pdserverapi.Health
}

// GetHealth queries the health info from PD server
func (pc *PDClient) GetHealth() (*PDHealthInfo, error) {
	url := fmt.Sprintf("%s/%s", pc.GetURL(), pdHealthURI)
	body, err := pc.httpClient.GetURL(url)
	if err != nil {
		return nil, err
	}

	healths := []pdserverapi.Health{}
	if err := json.Unmarshal(body, &healths); err != nil {
		return nil, err
	}
	return &PDHealthInfo{healths}, nil
}

// GetStores queries the stores info from PD server
func (pc *PDClient) GetStores() (*pdserverapi.StoresInfo, error) {
	url := fmt.Sprintf("%s/%s", pc.GetURL(), pdStoresURI)
	body, err := pc.httpClient.GetURL(url)
	if err != nil {
		return nil, err
	}

	storesInfo := pdserverapi.StoresInfo{}
	if err := json.Unmarshal(body, &storesInfo); err != nil {
		return nil, err
	}
	return &storesInfo, nil
}
