package model

// NodeInfo is used by nodes to report the status of each node
// to the hosts
type NodeInfo struct {
	Node       Node        `json:"node"`
	Containers []Container `json:"containers"`
}

// NodeInfoResponse is the response from the server after a NodeInfo
// is posted
type NodeInfoResponse struct {
	Containers []Container `json:"containers"`
}
