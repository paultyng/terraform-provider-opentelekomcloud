package vpcep

const (
	ErrClientCreate        = "error creating VPC Endpoint v1 client: %w"
	keyClient              = "vpcep-client"
	actionReceive   string = "receive"
	actionReject    string = "reject"
)

var approvalActionStatusMap = map[string]string{
	actionReceive: "accepted",
	actionReject:  "rejected",
}
