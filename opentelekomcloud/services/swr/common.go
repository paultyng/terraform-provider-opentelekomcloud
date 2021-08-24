package swr

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

const ClientError = "error creating SWR V2 client: %w"

func organization(d *schema.ResourceData) string {
	return d.Get("organization").(string)
}
