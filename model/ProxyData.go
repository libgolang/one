package model

// NodeInfo is used by nodes to report the status of each node
// to the hosts
type ProxyData struct {
	Node       Node        `json:"node"`
	Containers []Container `json:"containers"`
}

// NodeInfoResponse is the response from the server after a NodeInfo
// is posted
type ProxyDataResponse struct {
	Containers []Container `json:"containers"`
}
