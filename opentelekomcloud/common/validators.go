package common

import (
	"fmt"
	"net"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

var (
	tagsPattern    = regexp.MustCompile(`^[0-9a-zA-Z-_@]+$`)
	k8sTagsPattern = regexp.MustCompile(`^[./\-_A-Za-z0-9]+$`)
)

// ValidateStringList
// Deprecated
// Use validation.StringInSlice instead
func ValidateStringList(v interface{}, k string, l []string) (ws []string, errors []error) {
	value := v.(string)
	for i := range l {
		if value == l[i] {
			return
		}
	}
	errors = append(errors, fmt.Errorf("%q must be one of %v", k, l))
	return
}

// ValidateIntRange
// Deprecated
// Use validation.IntBetween instead
func ValidateIntRange(v interface{}, k string, l int, h int) (ws []string, errors []error) {
	i, ok := v.(int)
	if !ok {
		errors = append(errors, fmt.Errorf("%q must be an integer", k))
		return
	}
	if i < l || i > h {
		errors = append(errors, fmt.Errorf("%q must be between %d and %d", k, l, h))
		return
	}
	return
}

func ValidateTrueOnly(v interface{}, k string) (ws []string, errors []error) {
	if b, ok := v.(bool); ok && b {
		return
	}
	if v, ok := v.(string); ok && v == "true" {
		return
	}
	errors = append(errors, fmt.Errorf("%q must be true", k))
	return
}

func ValidateJsonString(v interface{}, k string) (ws []string, errors []error) {
	if _, err := NormalizeJsonString(v); err != nil {
		errors = append(errors, fmt.Errorf("%q contains an invalid JSON: %s", k, err))
	}
	return
}

func ValidateName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if len(value) > 64 || len(value) < 1 {
		errors = append(errors, fmt.Errorf("%q must contain more than 1 and less than 64 characters", k))
	}

	pattern := `^[\.\-_A-Za-z0-9]+$`
	if !regexp.MustCompile(pattern).MatchString(value) {
		errors = append(errors, fmt.Errorf("only alphanumeric characters, hyphens, and underscores allowed in %q", k))
	}

	return
}

func ValidateCTSEventName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if len(value) > 64 || len(value) < 1 {
		errors = append(errors, fmt.Errorf("%q must contain more than 1 and less than 64 characters", k))
	}

	pattern := `^[_A-Za-z0-9]+$`
	if !regexp.MustCompile(pattern).MatchString(value) {
		errors = append(errors, fmt.Errorf("only alphanumeric characters and underscores allowed in %q", k))
	}

	return
}

func ValidateStackTemplate(v interface{}, k string) (ws []string, errors []error) {
	if LooksLikeJsonString(v) {
		if _, err := NormalizeJsonString(v); err != nil {
			errors = append(errors, fmt.Errorf("%q contains an invalid JSON: %s", k, err))
		}
	} else {
		if _, err := CheckYamlString(v); err != nil {
			errors = append(errors, fmt.Errorf("%q contains an invalid YAML: %s", k, err))
		}
	}
	return
}

func ValidateIP(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	ipAddr := net.ParseIP(value)

	if ipAddr == nil || value != ipAddr.String() {
		errors = append(errors, fmt.Errorf("%q must contain a valid network IP address, got %q", k, value))
	}

	return
}

func ValidateCIDR(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	_, ipnet, err := net.ParseCIDR(value)
	if err != nil {
		errors = append(errors, fmt.Errorf(
			"%q must contain a valid CIDR, got error parsing: %s", k, err))
		return
	}

	if ipnet == nil || value != ipnet.String() {
		errors = append(errors, fmt.Errorf(
			"%q must contain a valid network CIDR, got %q", k, value))
	}

	return
}

func ValidateVBSPolicyName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if strings.HasPrefix(strings.ToLower(value), "default") {
		errors = append(errors, fmt.Errorf(
			"%q cannot start with default: %q", k, value))
	}

	if len(value) > 64 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be longer than 64 characters: %q", k, value))
	}
	pattern := `^[\.\-_A-Za-z0-9]+$`
	if !regexp.MustCompile(pattern).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q doesn't comply with restrictions (%q): %q",
			k, pattern, value))
	}
	return
}

func ValidateVBSPolicyFrequency(v interface{}, k string) (ws []string, errors []error) {
	value := v.(int)
	if value < 1 || value > 14 {
		errors = append(errors, fmt.Errorf(
			"%q should be in the range of 1-14: %d", k, value))
	}
	return
}

func ValidateVBSPolicyStatus(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if value != "ON" && value != "OFF" {
		errors = append(errors, fmt.Errorf(
			"%q should be either ON or OFF: %q", k, value))
	}
	return
}

func ValidateVBSPolicyRetentionNum(v interface{}, k string) (ws []string, errors []error) {
	value := v.(int)
	if value < 2 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be less than 2: %d", k, value))
	}
	return
}

func ValidateVBSPolicyRetainBackup(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if value != "Y" && value != "N" {
		errors = append(errors, fmt.Errorf(
			"%q should be either N or Y: %q", k, value))
	}
	return
}

func ValidateVBSTagKey(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	if len(value) > 36 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be longer than 36 characters: %q", k, value))
	}
	pattern := `^[\.\-_A-Za-z0-9]+$`
	if !regexp.MustCompile(pattern).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q doesn't comply with restrictions (%q): %q",
			k, pattern, value))
	}
	return
}

func ValidateVBSTagValue(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	if len(value) > 43 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be longer than 43 characters: %q", k, value))
	}
	pattern := `^[\.\-_A-Za-z0-9]+$`
	if !regexp.MustCompile(pattern).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q doesn't comply with restrictions (%q): %q",
			k, pattern, value))
	}
	return
}

func ValidateVBSBackupName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if strings.HasPrefix(strings.ToLower(value), "autobk") {
		errors = append(errors, fmt.Errorf(
			"%q cannot start with autobk: %q", k, value))
	}

	if len(value) > 64 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be longer than 64 characters: %q", k, value))
	}
	pattern := `^[\.\-_A-Za-z0-9]+$`
	if !regexp.MustCompile(pattern).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q doesn't comply with restrictions (%q): %q",
			k, pattern, value))
	}
	return
}

func ValidateAntiDdosTrafficPosID(v interface{}, k string) (ws []string, errors []error) {
	value := v.(int)
	if value < 1 || value > 9 {
		errors = append(errors, fmt.Errorf(
			"%q should be in the range of 1-9: %d", k, value))
	}
	return
}

func ValidateAntiDdosHttpRequestPosID(v interface{}, k string) (ws []string, errors []error) {
	value := v.(int)
	if value < 1 || value > 15 {
		errors = append(errors, fmt.Errorf(
			"%q should be in the range of 1-15: %d", k, value))
	}
	return
}

func ValidateAntiDdosCleaningAccessPosID(v interface{}, k string) (ws []string, errors []error) {
	value := v.(int)
	if value < 1 || value > 8 {
		errors = append(errors, fmt.Errorf(
			"%q should be in the range of 1-8: %d", k, value))
	}
	return
}

func ValidateAntiDdosAppTypeID(v interface{}, k string) (ws []string, errors []error) {
	value := v.(int)
	if value < 0 || value > 1 {
		errors = append(errors, fmt.Errorf(
			"%q should be 0 or 1: %d", k, value))
	}
	return
}

func ValidateTags(v interface{}, k string) (ws []string, errors []error) {
	tagMap := v.(map[string]interface{})

	for key, value := range tagMap {
		if !tagsPattern.MatchString(key) {
			errors = append(errors, fmt.Errorf("key %q doesn't comply with restrictions (%q): %q", k, tagsPattern, key))
		}
		if len(key) < 1 {
			errors = append(errors, fmt.Errorf("key %q cannot be shorter than 1 characters: %q", k, key))
		}
		if len(key) > 36 {
			errors = append(errors, fmt.Errorf("key %q cannot be longer than 36 characters: %q", k, key))
		}

		valueString := value.(string)
		if len(valueString) > 43 {
			errors = append(errors, fmt.Errorf("value %q cannot be longer than 43 characters: %q", k, value))
		}
	}

	return
}

func ValidateK8sTagsMap(v interface{}, k string) (ws []string, errors []error) {
	values := v.(map[string]interface{})

	for key, value := range values {
		valueString := value.(string)
		if len(key) < 1 {
			errors = append(errors, fmt.Errorf("key %q cannot be shorter than 1 characters: %q", k, key))
		}

		if len(valueString) < 1 {
			errors = append(errors, fmt.Errorf("value %q cannot be shorter than 1 characters: %q", k, value))
		}

		if len(key) > 253 {
			errors = append(errors, fmt.Errorf("key %q cannot be longer than 253 characters: %q", k, key))
		}

		if len(valueString) > 63 {
			errors = append(errors, fmt.Errorf("value %q cannot be longer than 63 characters: %q", k, value))
		}

		if !k8sTagsPattern.MatchString(key) {
			errors = append(errors, fmt.Errorf("key %q doesn't comply with restrictions (%q): %q", k, k8sTagsPattern, key))
		}

		if !k8sTagsPattern.MatchString(valueString) {
			errors = append(errors, fmt.Errorf("value %q doesn't comply with restrictions (%q): %q", k, k8sTagsPattern, valueString))
		}
	}

	return ws, errors
}

func ValidateDDSStartTime(v interface{}, _ string) (ws []string, errors []error) {
	value := v.(string)
	re := regexp.MustCompile(`(?P<hh>\d{2}):(?P<mm>\d{2})-(?P<HH>\d{2}):(?P<MM>\d{2})`)
	timeRange := re.FindStringSubmatch(value)

	paramsMap := make(map[string]string)
	for i, name := range re.SubexpNames() {
		if i > 0 && i <= len(timeRange) {
			paramsMap[name] = timeRange[i]
		}
	}

	startHour, err := strconv.Atoi(paramsMap["hh"])
	if err != nil {
		errors = append(errors, fmt.Errorf("%q must be int convertable: %s", paramsMap["hh"], err))
	}
	endHour, err := strconv.Atoi(paramsMap["HH"])
	if err != nil {
		errors = append(errors, fmt.Errorf("%q must be int convertable: %s", paramsMap["HH"], err))
	}
	startMinutes, err := strconv.Atoi(paramsMap["mm"])
	if err != nil {
		errors = append(errors, fmt.Errorf("%q must be int convertable: %s", paramsMap["mm"], err))
	}
	endMinutes, err := strconv.Atoi(paramsMap["MM"])
	if err != nil {
		errors = append(errors, fmt.Errorf("%q must be int convertable: %s", paramsMap["MM"], err))
	}
	if len(errors) != 0 {
		return ws, nil
	}
	if startHour+1 != endHour {
		errors = append(errors, fmt.Errorf("the `HH` value must be 1 greater than the `hh` value: %s", v))
	}
	if startMinutes != endMinutes {
		errors = append(errors, fmt.Errorf("the values from `mm` and `MM` must be the same: %s", v))
	}
	if startMinutes%15 != 0 {
		errors = append(errors, fmt.Errorf("the values from `mm` and `MM` must be set to any of the 00, 15, 30, or 45: %s", v))
	}

	return ws, errors
}

func ValidateASGroupListenerID(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	listenerIds := strings.Split(value, ",")
	if len(listenerIds) <= 3 {
		return
	}
	errors = append(errors, fmt.Errorf("%q supports binding up to 3 ELB listeners which are separated by a comma", k))
	return
}

func ValidateEmail(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	_, err := mail.ParseAddress(value)
	if err != nil {
		errors = append(errors, fmt.Errorf("%q doesn't comply with email standards: %s", k, value))
	}
	return
}

func ValidateIPRange(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	ipAddresses := strings.Split(value, "-")
	if len(ipAddresses) != 2 {
		errors = append(errors, fmt.Errorf(
			"%q must be a valid network IP address range, such as 0.0.0.0-255.255.255.0, but got %q", k, value))
		return
	}
	for _, address := range ipAddresses {
		ipnet := net.ParseIP(address)
		if ipnet == nil || address != ipnet.String() {
			errors = append(errors, fmt.Errorf("%q must contains valid network IP address, got %q", k, address))
		}
	}
	if len(errors) == 0 {
		if ipAddresses[0] == ipAddresses[1] {
			errors = append(errors, fmt.Errorf("Two network IP address of %q cannot equal, got %q", k, value))
		}
		startAddress := strings.Split(ipAddresses[0], ".")
		endAddress := strings.Split(ipAddresses[1], ".")
		for i := 0; i < len(startAddress); i++ {
			if startAddress[i] > endAddress[i] {
				errors = append(errors, fmt.Errorf(
					"%q starting IP address cannot be greater than the ending IP address, got %q", k, value))
				return
			} else if startAddress[i] < endAddress[i] {
				return
			}
		}
	}

	return
}

func ValidateDDMName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	// Check length between 4 and 64
	if len(value) > 64 || len(value) < 4 {
		errors = append(errors, fmt.Errorf("%q must contain more than 4 and less than 64 characters", k))
	}

	// Check if contains invalid character
	pattern := `^[\-A-Za-z0-9]+$`
	if !regexp.MustCompile(pattern).MatchString(value) {
		errors = append(errors, fmt.Errorf("only alphanumeric characters, and hyphens allowed in %q", k))
	}

	// Check if it doesn't start with a letter
	if !unicode.IsLetter(rune(value[0])) {
		errors = append(errors, fmt.Errorf("%q must start with a letter", k))
	}

	return
}

func ValidateUTCOffset(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	// Regular expression pattern for matching UTC offsets from +12:00 to -12:00
	pattern := `^(UTC([+-](0[1-9]|1[0-2]):00)|UTC)$`

	// Compile the regular expression
	re := regexp.MustCompile(pattern)
	if !re.MatchString(value) {
		errors = append(errors, fmt.Errorf("only valid utc offsets allowed in %q", k))
	}

	return
}

func ValidateDDMUsername(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	// Check length between 1 and 32
	if len(value) > 32 || len(value) < 1 {
		errors = append(errors, fmt.Errorf("%q must contain more than 1 and less than 32 characters", k))
	}

	// Check if contains invalid character
	pattern := `^[\_A-Za-z0-9]+$`
	if !regexp.MustCompile(pattern).MatchString(value) {
		errors = append(errors, fmt.Errorf("only alphanumeric characters, and underscores allowed in %q", k))
	}

	// Check if it doesn't start with a letter
	if !unicode.IsLetter(rune(value[0])) {
		errors = append(errors, fmt.Errorf("%q must start with a letter", k))
	}

	return
}
