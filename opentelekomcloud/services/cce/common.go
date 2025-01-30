package cce

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/cce/v3/nodes"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
)

const (
	cceClientError   = "error creating Open Telekom Cloud CCEv3 client: %w"
	keyClientV3      = "cce-v3-client"
	keyClientAddonV3 = "cce-addon-v3-client"
)

func buildResourceNodeStorage(d *schema.ResourceData) *nodes.Storage {
	v, ok := d.GetOk("storage")
	if !ok {
		return nil
	}

	var storageSpec nodes.Storage
	storageSpecRaw := v.([]interface{})
	storageSpecRawMap := storageSpecRaw[0].(map[string]interface{})
	storageSelectorSpecRaw := storageSpecRawMap["selectors"].([]interface{})
	storageGroupSpecRaw := storageSpecRawMap["groups"].([]interface{})

	var selectors []nodes.StorageSelector
	for _, s := range storageSelectorSpecRaw {
		sMap := s.(map[string]interface{})
		selector := nodes.StorageSelector{
			Name:        sMap["name"].(string),
			StorageType: sMap["type"].(string),
			MatchLabels: &nodes.MatchLabels{
				Size:              sMap["match_label_size"].(string),
				VolumeType:        sMap["match_label_volume_type"].(string),
				MetadataEncrypted: sMap["match_label_metadata_encrypted"].(string),
				MetadataCmkid:     sMap["match_label_metadata_cmkid"].(string),
				Count:             sMap["match_label_count"].(string),
			},
		}
		selectors = append(selectors, selector)
	}
	storageSpec.StorageSelectors = selectors

	var groups []nodes.StorageGroup
	for _, g := range storageGroupSpecRaw {
		gMap := g.(map[string]interface{})
		group := nodes.StorageGroup{
			Name:          gMap["name"].(string),
			CceManaged:    gMap["cce_managed"].(bool),
			SelectorNames: common.ExpandToStringList(gMap["selector_names"].([]interface{})),
		}

		virtualSpacesRaw := gMap["virtual_spaces"].([]interface{})
		virtualSpaces := make([]nodes.VirtualSpace, 0, len(virtualSpacesRaw))
		for _, v := range virtualSpacesRaw {
			virtualSpaceMap := v.(map[string]interface{})
			virtualSpace := nodes.VirtualSpace{
				Name: virtualSpaceMap["name"].(string),
				Size: virtualSpaceMap["size"].(string),
			}

			if virtualSpaceMap["lvm_lv_type"].(string) != "" {
				lvmConfig := nodes.LvmConfig{
					LvType: virtualSpaceMap["lvm_lv_type"].(string),
					Path:   virtualSpaceMap["lvm_path"].(string),
				}
				virtualSpace.LvmConfig = &lvmConfig
			}

			if virtualSpaceMap["runtime_lv_type"].(string) != "" {
				runtimeConfig := nodes.RuntimeConfig{
					LvType: virtualSpaceMap["runtime_lv_type"].(string),
				}
				virtualSpace.RuntimeConfig = &runtimeConfig
			}

			virtualSpaces = append(virtualSpaces, virtualSpace)
		}
		group.VirtualSpaces = virtualSpaces

		groups = append(groups, group)
	}

	storageSpec.StorageGroups = groups
	return &storageSpec
}

func buildResourceNodeLoginSpec(d *schema.ResourceData) (nodes.LoginSpec, error) {
	var loginSpec nodes.LoginSpec
	if v, ok := d.GetOk("key_pair"); ok {
		loginSpec = nodes.LoginSpec{
			SshKey: v.(string),
		}
	} else {
		password, err := common.TryPasswordEncrypt(d.Get("password").(string))
		if err != nil {
			return loginSpec, err
		}
		loginSpec = nodes.LoginSpec{
			UserPassword: nodes.UserPassword{
				Username: "root",
				Password: password,
			},
		}
	}

	return loginSpec, nil
}
