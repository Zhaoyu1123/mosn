/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package types

import (
	"context"
	"net"
	"sort"

	metrics "github.com/rcrowley/go-metrics"
	"sofastack.io/sofa-mosn/pkg/api/v2"
)

//   Below is the basic relation between clusterManager, cluster, hostSet, and hosts:
//
//           1              * | 1                1 | 1                *| 1          *
//   clusterManager --------- cluster  --------- prioritySet --------- hostSet------hosts

// ClusterManager manages connection pools and load balancing for upstream clusters.
type ClusterManager interface {
	// Add or update a cluster via API.
	AddOrUpdatePrimaryCluster(cluster v2.Cluster) error

	// Add Cluster health check callbacks
	AddClusterHealthCheckCallbacks(name string, cb HealthCheckCb) error

	// Get, use to get the snapshot of a cluster
	GetClusterSnapshot(context context.Context, cluster string) ClusterSnapshot

	// PutClusterSnapshot release snapshot lock
	PutClusterSnapshot(snapshot ClusterSnapshot)

	// UpdateClusterHosts used to update cluster's hosts
	// temp interface todo: remove it
	UpdateClusterHosts(cluster string, hosts []v2.Host) error

	// AppendClusterHosts used to add cluster's hosts
	AppendClusterHosts(clusterName string, hostConfigs []v2.Host) error

	// Get or Create tcp conn pool for a cluster
	TCPConnForCluster(balancerContext LoadBalancerContext, snapshot ClusterSnapshot) CreateConnectionData

	// ConnPoolForCluster used to get protocol related conn pool
	ConnPoolForCluster(balancerContext LoadBalancerContext, snapshot ClusterSnapshot, protocol Protocol) ConnectionPool

	// RemovePrimaryCluster used to remove cluster from set
	RemovePrimaryCluster(clusters ...string) error

	// ClusterExist, used to check whether 'clusterName' exist or not
	ClusterExist(clusterName string) bool

	// RemoveClusterHosts, remove the host by address string
	RemoveClusterHosts(clusterName string, hosts []string) error

	// Destroy the cluster manager
	Destroy()
}

// ClusterSnapshot is a thread-safe cluster snapshot
type ClusterSnapshot interface {
	HostSet() HostSet

	ClusterInfo() ClusterInfo

	LoadBalancer() LoadBalancer

	IsExistsHosts(metadata MetadataMatchCriteria) bool
}

// Cluster is a group of upstream hosts
type Cluster interface {
	Info() ClusterInfo

	HostSet() HostSet

	UpdateHosts([]Host)

	RemoveHosts([]string)

	LBInstance() LoadBalancer

	// Add health check callbacks in health checker
	AddHealthCheckCallbacks(cb HealthCheckCb)
}

// MemberUpdateCallback is called on create a priority set
type MemberUpdateCallback func(hostsAdded []Host, hostsRemoved []Host)

type HostPredicate func(Host) bool

// HostSet is as set of hosts that contains all of the endpoints for a given
// LocalityLbEndpoints priority level.
type HostSet interface {

	// all hosts that make up the set at the current time.
	Hosts() []Host

	HealthyHosts() []Host

	UpdateHosts(hosts []Host)

	RemoveHosts([]string)

	AdddMemberUpdateCb(cb MemberUpdateCallback)
}

// HealthFlag type
type HealthFlag int

const (
	// The host is currently failing active health checks.
	FAILED_ACTIVE_HC HealthFlag = 0x1
	// The host is currently considered an outlier and has been ejected.
	FAILED_OUTLIER_CHECK HealthFlag = 0x02
)

// Host is an upstream host
type Host interface {
	HostInfo

	// Create a connection for this host.
	CreateConnection(context context.Context) CreateConnectionData

	ClearHealthFlag(flag HealthFlag)

	ContainHealthFlag(flag HealthFlag) bool

	SetHealthFlag(flag HealthFlag)

	HealthFlag() HealthFlag

	Health() bool
}

// HostInfo defines a host's basic information
type HostInfo interface {
	Hostname() string

	Metadata() v2.Metadata

	ClusterInfo() ClusterInfo

	Address() net.Addr

	AddressString() string

	HostStats() HostStats

	Weight() uint32

	TLSDisable() bool

	Config() v2.Host

	// TODO: add deploy locality
}

// HostStats defines a host's statistics information
type HostStats struct {
	UpstreamConnectionTotal                        metrics.Counter
	UpstreamConnectionClose                        metrics.Counter
	UpstreamConnectionActive                       metrics.Counter
	UpstreamConnectionConFail                      metrics.Counter
	UpstreamConnectionLocalClose                   metrics.Counter
	UpstreamConnectionRemoteClose                  metrics.Counter
	UpstreamConnectionLocalCloseWithActiveRequest  metrics.Counter
	UpstreamConnectionRemoteCloseWithActiveRequest metrics.Counter
	UpstreamConnectionCloseNotify                  metrics.Counter
	UpstreamRequestTotal                           metrics.Counter
	UpstreamRequestActive                          metrics.Counter
	UpstreamRequestLocalReset                      metrics.Counter
	UpstreamRequestRemoteReset                     metrics.Counter
	UpstreamRequestTimeout                         metrics.Counter
	UpstreamRequestFailureEject                    metrics.Counter
	UpstreamRequestPendingOverflow                 metrics.Counter
	UpstreamRequestDuration                        metrics.Histogram
	UpstreamRequestDurationTotal                   metrics.Counter
	UpstreamResponseSuccess                        metrics.Counter
	UpstreamResponseFailed                         metrics.Counter
}

// ClusterInfo defines a cluster's information
type ClusterInfo interface {
	Name() string

	LbType() LoadBalancerType

	ConnectTimeout() int

	ConnBufferLimitBytes() uint32

	MaxRequestsPerConn() uint32

	Stats() ClusterStats

	ResourceManager() ResourceManager

	TLSMng() TLSContextManager

	LbSubsetInfo() LBSubsetInfo
}

// ResourceManager manages different types of Resource
type ResourceManager interface {
	// Connections resource to count connections in pool. Only used by protocol which has a connection pool which has multiple connections.
	Connections() Resource

	// Pending request resource to count pending requests. Only used by protocol which has a connection pool and pending requests to assign to connections.
	PendingRequests() Resource

	// Request resource to count requests
	Requests() Resource

	// Retries resource to count retries
	Retries() Resource
}

// Resource is a interface to statistics information
type Resource interface {
	CanCreate() bool
	Increase()
	Decrease()
	Max() uint64
}

// ClusterStats defines a cluster's statistics information
type ClusterStats struct {
	UpstreamConnectionTotal                        metrics.Counter
	UpstreamConnectionClose                        metrics.Counter
	UpstreamConnectionActive                       metrics.Counter
	UpstreamConnectionConFail                      metrics.Counter
	UpstreamConnectionRetry                        metrics.Counter
	UpstreamConnectionLocalClose                   metrics.Counter
	UpstreamConnectionRemoteClose                  metrics.Counter
	UpstreamConnectionLocalCloseWithActiveRequest  metrics.Counter
	UpstreamConnectionRemoteCloseWithActiveRequest metrics.Counter
	UpstreamConnectionCloseNotify                  metrics.Counter
	UpstreamBytesReadTotal                         metrics.Counter
	UpstreamBytesWriteTotal                        metrics.Counter
	UpstreamRequestTotal                           metrics.Counter
	UpstreamRequestActive                          metrics.Counter
	UpstreamRequestLocalReset                      metrics.Counter
	UpstreamRequestRemoteReset                     metrics.Counter
	UpstreamRequestRetry                           metrics.Counter
	UpstreamRequestRetryOverflow                   metrics.Counter
	UpstreamRequestTimeout                         metrics.Counter
	UpstreamRequestFailureEject                    metrics.Counter
	UpstreamRequestPendingOverflow                 metrics.Counter
	UpstreamRequestDuration                        metrics.Histogram
	UpstreamRequestDurationTotal                   metrics.Counter
	UpstreamResponseSuccess                        metrics.Counter
	UpstreamResponseFailed                         metrics.Counter
	LBSubSetsFallBack                              metrics.Counter
	LBSubSetsActive                                metrics.Counter
	LBSubsetsCreated                               metrics.Counter
	LBSubsetsRemoved                               metrics.Counter
}

type CreateConnectionData struct {
	Connection ClientConnection
	HostInfo   HostInfo
}

// SimpleCluster is a simple cluster in memory
type SimpleCluster interface {
	UpdateHosts(newHosts []Host)
}

// ClusterConfigFactoryCb is a callback interface
type ClusterConfigFactoryCb interface {
	UpdateClusterConfig(configs []v2.Cluster) error
}

// ClusterHostFactoryCb is a callback interface
type ClusterHostFactoryCb interface {
	UpdateClusterHost(cluster string, priority uint32, hosts []v2.Host) error
}

type ClusterManagerFilter interface {
	OnCreated(cccb ClusterConfigFactoryCb, chcb ClusterHostFactoryCb)
}

// RegisterUpstreamUpdateMethodCb is a callback interface
type RegisterUpstreamUpdateMethodCb interface {
	TriggerClusterUpdate(clusterName string, hosts []v2.Host)
	GetClusterNameByServiceName(serviceName string) string
}

type LBSubsetInfo interface {
	IsEnabled() bool

	FallbackPolicy() FallBackPolicy

	DefaultSubset() SubsetMetadata

	SubsetKeys() []SortedStringSetType
}

// SortedHosts is an implementation of sort.Interface
// a slice of host can be sorted as address string
type SortedHosts []Host

func (s SortedHosts) Len() int {
	return len(s)
}

func (s SortedHosts) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s SortedHosts) Less(i, j int) bool {
	return s[i].AddressString() < s[j].AddressString()
}

// SortedStringSetType is a sorted key collection with no duplicate
type SortedStringSetType struct {
	keys []string
}

// InitSet returns a SortedStringSetType
// The input key will be sorted and deduplicated
func InitSet(input []string) SortedStringSetType {
	var ssst SortedStringSetType
	var keys []string

	for _, keyInput := range input {
		exist := false

		for _, keyIn := range keys {
			if keyIn == keyInput {
				exist = true
				break
			}
		}

		if !exist {
			keys = append(keys, keyInput)
		}
	}
	ssst.keys = keys
	sort.Sort(&ssst)

	return ssst
}

// Keys is the keys in the collection
func (ss *SortedStringSetType) Keys() []string {
	return ss.keys
}

// Len is the number of elements in the collection.
func (ss *SortedStringSetType) Len() int {
	return len(ss.keys)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (ss *SortedStringSetType) Less(i, j int) bool {
	return ss.keys[i] < ss.keys[j]
}

// Swap swaps the elements with indexes i and j.
func (ss *SortedStringSetType) Swap(i, j int) {
	ss.keys[i], ss.keys[j] = ss.keys[j], ss.keys[i]
}

func init() {
	ConnPoolFactories = make(map[Protocol]bool)
}

var ConnPoolFactories map[Protocol]bool

func RegisterConnPoolFactory(protocol Protocol, registered bool) {
	//other
	ConnPoolFactories[protocol] = registered
}
