package v1alpha3

import (
	google_protobuf "github.com/gogo/protobuf/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VirtualService is a specification for a VirtualService resource
type VirtualService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec VirtualServiceSpec `json:"spec"`
}

// VirtualServiceSpec is the spec for a VirtualService resource
type VirtualServiceSpec struct {
	Hosts    []string     `json:"hosts,omitempty"`
	Gateways []string     `json:"gateways,omitempty"`
	Http     []*HTTPRoute `json:"http,omitempty"`
	Tcp      []*TCPRoute  `json:"tcp,omitempty"`
}

type HTTPRoute struct {
	Match            []*HTTPMatchRequest  `json:"match,omitempty"`
	Route            []*DestinationWeight `json:"route,omitempty"`
	Redirect         *HTTPRedirect        `json:"redirect,omitempty"`
	Rewrite          *HTTPRewrite         `json:"rewrite,omitempty"`
	WebsocketUpgrade bool                 `json:"websocket_upgrade,omitempty"`
	Timeout          time.Duration        `json:"timeout,omitempty"`
	Retries          *HTTPRetry           `json:"retries,omitempty"`
	Fault            *HTTPFaultInjection  `json:"fault,omitempty"`
	Mirror           *Destination         `json:"mirror,omitempty"`
	CorsPolicy       *CorsPolicy          `json:"cors_policy,omitempty"`
	AppendHeaders    map[string]string    `json:"append_headers,omitempty"`
}

type HTTPMatchRequest struct {
	Uri          *StringMatch           `json:"uri,omitempty"`
	Scheme       *StringMatch           `json:"scheme,omitempty"`
	Method       *StringMatch           `json:"method,omitempty"`
	Authority    *StringMatch           `json:"authority,omitempty"`
	Headers      map[string]StringMatch `json:"headers,omitempty"`
	Port         uint32                 `json:"port,omitempty"`
	SourceLabels map[string]string      `json:"source_labels,omitempty"`
	Gateways     []string               `json:"gateways,omitempty"`
}

type DestinationWeight struct {
	Destination *Destination `json:"destination,omitempty"`
	Weight      int32        `json:"weight,omitempty"`
}

type HTTPRedirect struct {
	Uri       string `json:"uri,omitempty"`
	Authority string `json:"authority,omitempty"`
}

type HTTPRewrite struct {
	Uri       string `json:"uri,omitempty"`
	Authority string `json:"authority,omitempty"`
}

type HTTPRetry struct {
	Attempts      int32         `json:"attempts,omitempty"`
	PerTryTimeout time.Duration `json:"per_try_timeout,omitempty"`
}

type HTTPFaultInjection struct {
	Delay *Delay `json:"delay,omitempty"`
	Abort *Abort `json:"abort,omitempty"`
}

type Delay struct {
	Percent          int32         `json:"percent,omitempty"`
	FixedDelay       time.Duration `json:"fixed_delay,omitempty"`
	ExponentialDelay time.Duration `json:"exponential_delay,omitempty"`
}

type Abort struct {
	Percent    int32  `json:"percent,omitempty"`
	HttpStatus int32  `json:"http_status,omitempty"`
	GrpcStatus string `json:"grpc_status,omitempty"`
	Http2Error string `json:"http2_error,omitempty"`
}

type Destination struct {
	Host   string        `json:"host,omitempty"`
	Subset string        `json:"subset,omitempty"`
	Port   *PortSelector `json:"port,omitempty"`
}

type CorsPolicy struct {
	AllowOrigin      []string      `json:"allow_origin,omitempty"`
	AllowMethods     []string      `json:"allow_methods,omitempty"`
	AllowHeaders     []string      `json:"allow_headers,omitempty"`
	ExposeHeaders    []string      `json:"expose_headers,omitempty"`
	MaxAge           time.Duration `json:"max_age,omitempty"`
	AllowCredentials bool          `json:"allow_credentials,omitempty"`
}

type StringMatch struct {
	Exact  string `json:"exact,omitempty"`
	Prefix string `json:"prefix,omitempty"`
	Regex  string `json:"regex,omitempty"`
}

type PortSelector struct {
	Name   string `json:"name,omitempty"`
	Number uint32 `json:"number,omitempty"`
}

type TCPRoute struct {
	Match []*L4MatchAttributes `json:"match,omitempty"`
	Route []*DestinationWeight `json:"route,omitempty"`
}

type L4MatchAttributes struct {
	DestinationSubnet string            `json:"destination_subnet,omitempty"`
	Port              uint32            `json:"port,omitempty"`
	SourceSubnet      string            `json:"source_subnet,omitempty"`
	SourceLabels      map[string]string `json:"source_labels,omitempty"`
	Gateways          []string          `json:"gateways,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VirtualServiceList is a list of VirtualService resources
type VirtualServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []VirtualService `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DestinationRule is a specification for a DestinationRule resource
type DestinationRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec DestinationRuleSpec `json:"spec"`
}

// DestinationRuleSpec is the spec for a DestinationRule resource
type DestinationRuleSpec struct {
	Host          string         `json:"host,omitempty"`
	TrafficPolicy *TrafficPolicy `json:"trafficPolicy,omitempty"`
	Subsets       []*Subset      `json:"subsets,omitempty"`
}

type TrafficPolicy struct {
	LoadBalancer      *LoadBalancerSettings              `json:"loadBalancer,omitempty"`
	ConnectionPool    *ConnectionPoolSettings            `json:"connectionPool,omitempty"`
	OutlierDetection  *OutlierDetection                  `json:"outlierDetection,omitempty"`
	Tls               *TLSSettings                       `json:"tls,omitempty"`
	PortLevelSettings []*TrafficPolicy_PortTrafficPolicy `json:"portLevelSettings,omitempty"`
}

type TrafficPolicy_PortTrafficPolicy struct {
	Port             *PortSelector           `json:"port,omitempty"`
	LoadBalancer     *LoadBalancerSettings   `json:"loadBalancer,omitempty"`
	ConnectionPool   *ConnectionPoolSettings `json:"connectionPool,omitempty"`
	OutlierDetection *OutlierDetection       `json:"outlierDetection,omitempty"`
	Tls              *TLSSettings            `json:"tls,omitempty"`
}

type Subset struct {
	Name          string            `json:"name,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
	TrafficPolicy *TrafficPolicy    `json:"trafficPolicy,omitempty"`
}

type LoadBalancerSettings struct {
	Simple         string                                 `json:"simple,omitempty"`
	ConsistentHash *LoadBalancerSettings_ConsistentHashLB `json=consistentHash,omitempty"`
}

type LoadBalancerSettings_ConsistentHashLB struct {
	HttpHeader      string `json:"httpHeader,omitempty"`
	MinimumRingSize uint32 `json:"minimumRingSize,omitempty"`
}

type ConnectionPoolSettings struct {
	Tcp  *ConnectionPoolSettings_TCPSettings  `json:"tcp,omitempty"`
	Http *ConnectionPoolSettings_HTTPSettings `json:"http,omitempty"`
}

type ConnectionPoolSettings_TCPSettings struct {
	MaxConnections int32                     `json:"maxConnections,omitempty"`
	ConnectTimeout *google_protobuf.Duration `json:"connectTimeout,omitempty"`
}

type ConnectionPoolSettings_HTTPSettings struct {
	Http1MaxPendingRequests  int32 `json:"http1MaxPendingRequests,omitempty"`
	Http2MaxRequests         int32 `json:"http2MaxRequests,omitempty"`
	MaxRequestsPerConnection int32 `json:"maxRequestsPerConnection,omitempty"`
	MaxRetries               int32 `json:"maxRetries,omitempty"`
}

type OutlierDetection struct {
	Http *OutlierDetection_HTTPSettings `json:"http,omitempty"`
}

type OutlierDetection_HTTPSettings struct {
	ConsecutiveErrors  int32                     `json:"consecutiveErrors,omitempty"`
	Interval           *google_protobuf.Duration `json:"interval,omitempty"`
	BaseEjectionTime   *google_protobuf.Duration `json:"baseEjectionTime,omitempty"`
	MaxEjectionPercent int32                     `json:"maxEjectionPercent,omitempty"`
}

type TLSSettings struct {
	Mode              string   `json:"mode,omitempty"`
	ClientCertificate string   `json:"clientCertificate,omitempty"`
	PrivateKey        string   `json:"privateKey,omitempty"`
	CaCertificates    string   `json:"caCertificates,omitempty"`
	SubjectAltNames   []string `json:"subjectAltNames,omitempty"`
	Sni               string   `json:"sni,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DestinationRuleList is a list of DestinationRule resources
type DestinationRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []DestinationRule `json:"items"`
}
