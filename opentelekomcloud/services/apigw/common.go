package apigw

type (
	PolicyType string
	PeriodUnit string

	ApiType         string
	RequestMethod   string
	ApiAuthType     string
	ParamLocation   string
	ParamType       string
	MatchMode       string
	InvocationType  string
	EffectiveMode   string
	ConditionSource string
	ConditionType   string
	ParameterType   string
	SystemParamType string
	BackendType     string
	AppCodeAuthType string
	ProtocolType    string
	NetworkType     string

	SecretAction string
	RouteConfig  struct {
		UserRoutes []interface{} `json:"user_routes"`
	}
)

const (
	keyClientV2         = "apigw-v2-client"
	errCreationV2Client = "error creating OpenTelekomCloud APIGW V2 client: %w"

	PeriodUnitSecond PeriodUnit = "SECOND"
	PeriodUnitMinute PeriodUnit = "MINUTE"
	PeriodUnitHour   PeriodUnit = "HOUR"
	PeriodUnitDay    PeriodUnit = "DAY"

	PolicyTypeExclusive   PolicyType = "API-based"
	PolicyTypeShared      PolicyType = "API-shared"
	PolicyTypeUser        PolicyType = "USER"
	PolicyTypeApplication PolicyType = "APP"

	includeSpecialThrottle int = 1

	ApiTypePublic  ApiType = "Public"
	ApiTypePrivate ApiType = "Private"

	RequestMethodGet     RequestMethod = "GET"
	RequestMethodPost    RequestMethod = "POST"
	RequestMethodPut     RequestMethod = "PUT"
	RequestMethodDelete  RequestMethod = "DELETE"
	RequestMethodHead    RequestMethod = "HEAD"
	RequestMethodPatch   RequestMethod = "PATCH"
	RequestMethodOptions RequestMethod = "OPTIONS"
	RequestMethodAny     RequestMethod = "ANY"

	ApiAuthTypeNone       ApiAuthType = "NONE"
	ApiAuthTypeApp        ApiAuthType = "APP"
	ApiAuthTypeIam        ApiAuthType = "IAM"
	ApiAuthTypeAuthorizer ApiAuthType = "AUTHORIZER"

	ParamLocationPath   ParamLocation = "PATH"
	ParamLocationHeader ParamLocation = "HEADER"
	ParamLocationQuery  ParamLocation = "QUERY"

	ParamTypeString ParamType = "STRING"
	ParamTypeNumber ParamType = "NUMBER"

	MatchModePrefix MatchMode = "PREFIX"
	MatchModeExact  MatchMode = "EXACT"

	InvocationTypeAsync InvocationType = "async"
	InvocationTypeSync  InvocationType = "sync"

	EffectiveModeAll EffectiveMode = "ALL"
	EffectiveModeAny EffectiveMode = "ANY"

	ConditionSourceParam              ConditionSource = "param"
	ConditionSourceSource             ConditionSource = "source"
	ConditionSourceSystem             ConditionSource = "system"
	ConditionSourceCookie             ConditionSource = "cookie"
	ConditionSourceFrontendAuthorizer ConditionSource = "frontend_authorizer"

	ConditionTypeEqual      ConditionType = "EXACT"
	ConditionTypeEnumerated ConditionType = "ENUM"
	ConditionTypeMatching   ConditionType = "PATTERN"

	ParameterTypeRequest  ParameterType = "REQUEST"
	ParameterTypeConstant ParameterType = "CONSTANT"
	ParameterTypeSystem   ParameterType = "SYSTEM"

	SystemParamTypeFrontend SystemParamType = "frontend"
	SystemParamTypeBackend  SystemParamType = "backend"
	SystemParamTypeInternal SystemParamType = "internal"

	BackendTypeHttp     BackendType = "HTTP"
	BackendTypeFunction BackendType = "FUNCTION"
	BackendTypeMock     BackendType = "MOCK"

	AppCodeAuthTypeDisable AppCodeAuthType = "DISABLE"
	AppCodeAuthTypeEnable  AppCodeAuthType = "HEADER"

	ProtocolTypeTCP   ProtocolType = "TCP"
	ProtocolTypeHTTP  ProtocolType = "HTTP"
	ProtocolTypeHTTPS ProtocolType = "HTTPS"
	ProtocolTypeBoth  ProtocolType = "BOTH"

	strBoolEnabled  int = 1
	strBoolDisabled int = 2

	NetworkTypeV1 NetworkType = "NON-VPC"
	NetworkTypeV2 NetworkType = "VPC"

	ChargingModeBandwidth = "bandwidth"
	ChargingModeTraffic   = "traffic"

	SecretActionReset SecretAction = "RESET"
)

var (
	policyType = map[string]int{
		string(PolicyTypeExclusive): 1,
		string(PolicyTypeShared):    2,
	}
	matching = map[string]string{
		string(MatchModePrefix): "SWA",
		string(MatchModeExact):  "NORMAL",
	}
	conditionType = map[string]string{
		string(ConditionTypeEqual):      "exact",
		string(ConditionTypeEnumerated): "enum",
		string(ConditionTypeMatching):   "pattern",
	}
)
