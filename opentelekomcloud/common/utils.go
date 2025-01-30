package common

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/GehirnInc/crypt"

	_ "github.com/GehirnInc/crypt/sha512_crypt"

	ver "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmespath/go-jmespath"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

var letters = []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func LooksLikeJsonString(s interface{}) bool {
	return regexp.MustCompile(`^\s*{`).MatchString(s.(string))
}

func Base64IfNot(src string) string {
	_, err := base64.StdEncoding.DecodeString(src)
	if err == nil {
		return src
	}
	return base64.StdEncoding.EncodeToString([]byte(src))
}

type versionSlice []*ver.Version

func (v versionSlice) Len() int {
	return len(v)
}

func (v versionSlice) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

func (v versionSlice) Less(i, j int) bool {
	return v[i].LessThan(v[j])
}

func (v versionSlice) ToStringSlice() []string {
	res := make([]string, len(v))
	for i, version := range v {
		res[i] = version.Original()
	}
	return res
}

func sortAsStringSlice(src []string) []string {
	res := make([]string, len(src))
	copy(res, src)
	sort.Sort(sort.Reverse(sort.StringSlice(res)))
	return res
}

// SortVersions sorts versions from newer to older.
// If non-version-like string will be found in the slice,
// slice will be sorted as string slice in reversed order (z-a)
func SortVersions(src []string) []string {
	verSlice := make(versionSlice, len(src))
	for i, v := range src {
		val, err := ver.NewVersion(v)
		if err != nil {
			return sortAsStringSlice(src) // in case it's not version-like
		}
		verSlice[i] = val
	}
	sort.Sort(sort.Reverse(verSlice))
	return verSlice.ToStringSlice()
}

// BuildRequest takes an opts struct and builds a request body for
// Gophercloud to execute
func BuildRequest(opts interface{}, parent string) (map[string]interface{}, error) {
	b, err := golangsdk.BuildRequestBody(opts, "")
	if err != nil {
		return nil, err
	}
	b = AddValueSpecs(b)
	return map[string]interface{}{parent: b}, nil
}

// CheckDeleted checks the error to see if it's a 404 (Not Found) and, if so,
// sets the resource ID to the empty string instead of throwing an error.
func CheckDeleted(d *schema.ResourceData, err error, msg string) error {
	_, ok := err.(golangsdk.ErrDefault404)
	if ok {
		d.SetId("")
		return nil
	}

	return fmt.Errorf("%s: %s", msg, err)
}

// CheckDeletedDiag checks the error to see if it's a 404 (Not Found) and, if so,
// sets the resource ID to the empty string instead of throwing an error.
func CheckDeletedDiag(d *schema.ResourceData, err error, msg string) diag.Diagnostics {
	if _, ok := err.(golangsdk.ErrDefault404); ok {
		d.SetId("")
		return nil
	}

	return fmterr.Errorf("%s: %s", msg, err)
}

// AddValueSpecs expands the 'value_specs' object and removes 'value_specs'
// from the request body.
func AddValueSpecs(body map[string]interface{}) map[string]interface{} {
	if body["value_specs"] != nil {
		for k, v := range body["value_specs"].(map[string]interface{}) {
			body[k] = v
		}
		delete(body, "value_specs")
	}

	return body
}

// MapValueSpecs converts ResourceData into a map
func MapValueSpecs(d cfg.SchemaOrDiff) map[string]string {
	m := make(map[string]string)
	for key, val := range d.Get("value_specs").(map[string]interface{}) {
		m[key] = val.(string)
	}
	return m
}

// MapResourceProp converts ResourceData property into a map
func MapResourceProp(d *schema.ResourceData, prop string) map[string]interface{} {
	m := make(map[string]interface{})
	for key, val := range d.Get(prop).(map[string]interface{}) {
		m[key] = val.(string)
	}
	return m
}

func CheckForRetryableError(err error) *resource.RetryError {
	switch err.(type) {
	case golangsdk.ErrDefault409, golangsdk.ErrDefault500, golangsdk.ErrDefault503:
		return resource.RetryableError(err)
	default:
		return resource.NonRetryableError(err)
	}
}

func IsResourceNotFound(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(golangsdk.ErrDefault404)
	return ok
}

func ExpandToStringSlice(v []interface{}) []string {
	s := make([]string, 0, len(v))
	for _, val := range v {
		if strVal, ok := val.(string); ok && strVal != "" {
			s = append(s, strVal)
		}
	}

	return s
}

// StrSliceContains checks if a given string is contained in a slice
// When anybody asks why Go needs generics, here you go.
func StrSliceContains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}

func GetAllAvailableZones(d *schema.ResourceData) []string {
	rawZones := d.Get("available_zones").([]interface{})
	zones := make([]string, len(rawZones))
	for i, raw := range rawZones {
		zones[i] = raw.(string)
	}
	log.Printf("[DEBUG] getAvailableZones: %#v", zones)

	return zones
}

func StringInSlice(str string, slice []string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}

func BuildComponentID(parts ...string) string {
	return strings.Join(parts, "/")
}

// StrSlice is used to wrap single string element in slice
func StrSlice(v interface{}) []string {
	if v == "" {
		return nil
	}
	return []string{v.(string)}
}

// IntSlice is used to wrap single integer element in slice
func IntSlice(v interface{}) []int {
	if v == 0 {
		return nil
	}
	return []int{v.(int)}
}

var (
	DataSourceTooFewDiag  = diag.Errorf("your query returned no results. Please change your search criteria and try again.")
	DataSourceTooManyDiag = diag.Errorf("your query returned more than one result. Please change your search criteria and try again.")
)

// GetSetChanges returns a pair of sets describing removed and added items
func GetSetChanges(d *schema.ResourceData, key string) (removed, added *schema.Set) {
	oldOne, newOne := d.GetChange(key)
	oldSet := oldOne.(*schema.Set)
	newSet := newOne.(*schema.Set)
	return oldSet.Difference(newSet), newSet.Difference(oldSet)
}

// CheckNull returns true if schema parameter is empty
func CheckNull(element string, d *schema.ResourceData) bool {
	return d.GetRawConfig().GetAttr(element).IsNull()
}

func CompareJsonTemplateAreEquivalent(tem1, tem2 string) (bool, error) {
	var obj1 interface{}
	err := json.Unmarshal([]byte(tem1), &obj1)
	if err != nil {
		return false, err
	}

	canonicalJson1, _ := json.Marshal(obj1)

	var obj2 interface{}
	err = json.Unmarshal([]byte(tem2), &obj2)
	if err != nil {
		return false, err
	}

	canonicalJson2, _ := json.Marshal(obj2)

	equal := bytes.Equal(canonicalJson1, canonicalJson2)
	if !equal {
		log.Printf("[DEBUG] Canonical template are not equal.\nFirst: %s\nSecond: %s\n",
			canonicalJson1, canonicalJson2)
	}
	return equal, nil
}

func ValidateRFC3339Timestamp(v interface{}, _ string) (ws []string, errors []error) {
	value := v.(string)
	_, err := time.Parse(time.RFC3339, fmt.Sprintf("%sT00:00:00Z", value))
	if err != nil {
		errors = append(errors, fmt.Errorf(
			"%q cannot be parsed as RFC3339 Timestamp Format", value))
	}

	return
}

// FilterSliceWithField can filter the slice all through a map filter.
// If the field is a nested value, using dot(.) to split them, e.g. "SubBlock.SubField".
// If value in the map is zero, it will be ignored.
func FilterSliceWithField(all interface{}, filter map[string]interface{}) ([]interface{}, error) {
	return filterSliceWithFieldRaw(all, filter, true)
}

func filterSliceWithFieldRaw(all interface{}, filter map[string]interface{}, ignoreZero bool) ([]interface{}, error) {
	var result []interface{}
	var matched bool

	allValue := reflect.ValueOf(all)
	if allValue.Kind() != reflect.Slice {
		return nil, fmt.Errorf("options type is not a slice")
	}

	newFilter := filter
	if ignoreZero {
		for key, val := range filter {
			keyValue := reflect.ValueOf(val)
			if keyValue.IsZero() {
				log.Printf("[DEBUG] ignore zero field %s", key)
				delete(newFilter, key)
			}
		}
	}

	for i := 0; i < allValue.Len(); i++ {
		refValue := allValue.Index(i)
		if refValue.Kind() == reflect.Ptr {
			refValue = refValue.Elem()
		}
		if refValue.Kind() != reflect.Struct {
			return nil, fmt.Errorf("object in slice is not a struct")
		}

		matched = true
		for key, val := range newFilter {
			actual, err := getStructField(refValue, key)
			if err != nil {
				return nil, fmt.Errorf("get slice field %s failed: %s", key, err)
			}

			actualVal := reflect.ValueOf(actual)
			if actualVal.Kind() == reflect.Ptr {
				actualVal = actualVal.Elem()
			}

			if actualVal.Interface() != val {
				log.Printf("[DEBUG] can not match slice[%d] field %s: expect %v, but got %v", i, key, val, actualVal)
				matched = false
				break
			}
		}

		if matched {
			result = append(result, refValue.Interface())
		}
	}
	return result, nil
}

func getStructField(v reflect.Value, field string) (interface{}, error) {
	var subField interface{}
	var err error
	structValue := v

	parts := strings.Split(field, ".")
	for _, key := range parts {
		subField, err = getStructFieldRaw(structValue, key)
		if err != nil {
			return nil, err
		}
		structValue = reflect.ValueOf(subField)
	}
	return subField, nil
}

func getStructFieldRaw(v reflect.Value, field string) (interface{}, error) {
	if v.Kind() == reflect.Struct {
		value := reflect.Indirect(v).FieldByName(field)
		if value.IsValid() {
			return value.Interface(), nil
		}

		return nil, fmt.Errorf("reflect: can not find the field %s", field)
	}
	return nil, fmt.Errorf("reflect: Value is not a struct")
}

// StringToInt convert the string to int, and return the pointer of int value
func StringToInt(i *string) *int {
	if i == nil || len(*i) == 0 {
		return nil
	}

	r, err := strconv.Atoi(*i)
	if err != nil {
		log.Printf("[ERROR] convert the string %q to int failed.", *i)
	}
	return &r
}

// ExpandToStringListBySet takes the result for a set of strings and returns a []string
func ExpandToStringListBySet(v *schema.Set) []string {
	s := make([]string, 0, v.Len())
	for _, val := range v.List() {
		if strVal, ok := val.(string); ok && strVal != "" {
			s = append(s, strVal)
		}
	}

	return s
}

// SliceUnion returns a new slice containing the union of elements from both slices,
// without any duplicates.
func SliceUnion(a, b []string) []string {
	var res []string
	for _, i := range a {
		if !StrSliceContains(res, i) {
			res = append(res, i)
		}
	}
	for _, k := range b {
		if !StrSliceContains(res, k) {
			res = append(res, k)
		}
	}
	return res
}

func RemoveNil(data map[string]interface{}) map[string]interface{} {
	withoutNil := make(map[string]interface{})

	for k, v := range data {
		if v == nil {
			continue
		}

		switch v := v.(type) {
		case map[string]interface{}:
			if len(v) > 0 {
				withoutNil[k] = RemoveNil(v)
			}
		case []map[string]interface{}:
			rv := make([]map[string]interface{}, 0, len(v))
			for _, vv := range v {
				rst := RemoveNil(vv)
				if len(rst) > 0 {
					rv = append(rv, rst)
				}
			}
			if len(rv) > 0 {
				withoutNil[k] = rv
			}
		default:
			withoutNil[k] = v
		}
	}

	return withoutNil
}

// IsSliceContainsAnyAnotherSliceElement is a method that used to determine whether a list contains any element of
// another list (including its fragments belonging to the current string), returns true if it contains.
// sl: The slice body used to determine the inclusion relationship.
// another: The included slice object used to determine the inclusion relationship.
// ignoreCase: Whether to ignore case.
// isExcat: Whether the inclusion relationship of string objects applies exact matching rules.
func IsSliceContainsAnyAnotherSliceElement(sl, another []string, ignoreCase, isExcat bool) bool {
	for _, elem := range sl {
		if IsStrContainsSliceElement(elem, another, ignoreCase, isExcat) {
			return true
		}
	}
	return false
}

// IsStrContainsSliceElement returns true if the string exists in given slice or contains in one of slice elements when
// open exact flag. Also, you can ignore case for this check.
func IsStrContainsSliceElement(str string, sl []string, ignoreCase, isExcat bool) bool {
	if ignoreCase {
		str = strings.ToLower(str)
	}
	for _, s := range sl {
		if ignoreCase {
			s = strings.ToLower(s)
		}
		if isExcat && s == str {
			return true
		}
		if !isExcat && strings.Contains(str, s) {
			return true
		}
	}
	return false
}

// ExpandToStringList takes the result for an array of strings and returns a []string
func ExpandToStringList(v []interface{}) []string {
	s := make([]string, 0, len(v))
	for _, val := range v {
		if strVal, ok := val.(string); ok && strVal != "" {
			s = append(s, strVal)
		}
	}

	return s
}

func ExpandToIntList(v []interface{}) []int {
	s := make([]int, 0, len(v))
	for _, val := range v {
		if intVal, ok := val.(int); ok {
			s = append(s, intVal)
		}
	}
	return s
}

// ValueIgnoreEmpty returns to the string value. if v is empty, return nil
func ValueIgnoreEmpty(v interface{}) interface{} {
	vl := reflect.ValueOf(v)
	if (vl.Kind() != reflect.Bool) && vl.IsZero() {
		return nil
	}

	if (vl.Kind() == reflect.Array || vl.Kind() == reflect.Slice) && vl.Len() == 0 {
		return nil
	}

	return v
}

// SetIfNotEmpty to set key if value is not empty
func SetIfNotEmpty(target map[string]interface{}, key string, value interface{}) {
	if value != nil && value != "" {
		target[key] = value
	}
}

// PathSearch evaluates a JMESPath expression against input data and returns the result.
func PathSearch(expression string, obj interface{}, defaultValue interface{}) interface{} {
	v, err := jmespath.Search(expression, obj)
	if err != nil || v == nil {
		return defaultValue
	}
	return v
}

func MapContains(rawMap map[string]string, filterMap map[string]interface{}) bool {
	if len(filterMap) < 1 {
		return true
	}

	hasContain := true
	for key, value := range filterMap {
		hasContain = hasContain && hasMapContain(rawMap, key, value.(string))
	}

	return hasContain
}

func hasMapContain(rawMap map[string]string, filterKey, filterValue string) bool {
	if rawTag, ok := rawMap[filterKey]; ok {
		if filterValue != "" {
			filterTagValues := strings.Split(filterValue, ",")
			return StrSliceContains(filterTagValues, rawTag)
		} else {
			return true
		}
	} else {
		return false
	}
}

// StrSliceContainsAnother checks whether a string slice (b) contains another string slice (s).
func StrSliceContainsAnother(b []string, s []string) bool {
	// The empty set is the subset of any set.
	if len(s) < 1 {
		return true
	}
	for _, v := range s {
		if !StrSliceContains(b, v) {
			return false
		}
	}
	return true
}

// Salt generates a random salt according to given size
func Salt(size int) ([]byte, error) {
	salt := make([]byte, size)
	_, err := io.ReadFull(rand.Reader, salt)
	if err != nil {
		return nil, fmt.Errorf("error generating salt: %s", err)
	}

	max := uint8(len(letters))
	arc := uint8(0)
	for i, x := range salt {
		arc = x % max
		salt[i] = letters[arc]
	}
	return salt, nil
}

// TryPasswordEncrypt tries to encrypt given password if it's not encrypted
func TryPasswordEncrypt(password string) (string, error) {
	if _, err := base64.StdEncoding.DecodeString(password); err != nil {
		return PasswordEncrypt(password)
	}
	return password, nil
}

// PasswordEncrypt encrypts given password with sha512
func PasswordEncrypt(password string) (string, error) {
	saltBytes, err := Salt(16)
	if err != nil {
		return "", err
	}
	salt := "$6$" + string(saltBytes) + "$"

	sha512crypt := crypt.SHA512.New()
	passwordEncrypted, err := sha512crypt.Generate([]byte(password), []byte(salt))
	if err != nil {
		return "", fmt.Errorf("error encrypting the password: %s", err)
	}
	return base64.StdEncoding.EncodeToString([]byte(passwordEncrypted)), nil
}
