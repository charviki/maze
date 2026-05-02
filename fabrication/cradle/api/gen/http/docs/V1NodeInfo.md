# V1NodeInfo



## Properties



Name | Type | Description | Notes

------------ | ------------- | ------------- | -------------

**Name** | Pointer to **string** |  | [optional]

**Address** | Pointer to **string** |  | [optional]

**ExternalAddr** | Pointer to **string** |  | [optional]

**GrpcAddress** | Pointer to **string** |  | [optional]

**AuthToken** | Pointer to **string** |  | [optional]

**Status** | Pointer to **string** |  | [optional]

**RegisteredAt** | Pointer to **string** |  | [optional]

**LastHeartbeat** | Pointer to **string** |  | [optional]

**Capabilities** | Pointer to [**V1AgentCapabilities**](V1AgentCapabilities.md) |  | [optional]

**AgentStatus** | Pointer to [**V1AgentStatus**](V1AgentStatus.md) |  | [optional]

**Metadata** | Pointer to [**V1AgentMetadata**](V1AgentMetadata.md) |  | [optional]



## Methods



### NewV1NodeInfo



`func NewV1NodeInfo() *V1NodeInfo`



NewV1NodeInfo instantiates a new V1NodeInfo object

This constructor will assign default values to properties that have it defined,

and makes sure properties required by API are set, but the set of arguments

will change when the set of required properties is changed



### NewV1NodeInfoWithDefaults



`func NewV1NodeInfoWithDefaults() *V1NodeInfo`



NewV1NodeInfoWithDefaults instantiates a new V1NodeInfo object

This constructor will only assign default values to properties that have it defined,

but it doesn't guarantee that properties required by API are set



### GetName



`func (o *V1NodeInfo) GetName() string`



GetName returns the Name field if non-nil, zero value otherwise.



### GetNameOk



`func (o *V1NodeInfo) GetNameOk() (*string, bool)`



GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetName



`func (o *V1NodeInfo) SetName(v string)`



SetName sets Name field to given value.



### HasName



`func (o *V1NodeInfo) HasName() bool`



HasName returns a boolean if a field has been set.



### GetAddress



`func (o *V1NodeInfo) GetAddress() string`



GetAddress returns the Address field if non-nil, zero value otherwise.



### GetAddressOk



`func (o *V1NodeInfo) GetAddressOk() (*string, bool)`



GetAddressOk returns a tuple with the Address field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetAddress



`func (o *V1NodeInfo) SetAddress(v string)`



SetAddress sets Address field to given value.



### HasAddress



`func (o *V1NodeInfo) HasAddress() bool`



HasAddress returns a boolean if a field has been set.



### GetExternalAddr



`func (o *V1NodeInfo) GetExternalAddr() string`



GetExternalAddr returns the ExternalAddr field if non-nil, zero value otherwise.



### GetExternalAddrOk



`func (o *V1NodeInfo) GetExternalAddrOk() (*string, bool)`



GetExternalAddrOk returns a tuple with the ExternalAddr field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetExternalAddr



`func (o *V1NodeInfo) SetExternalAddr(v string)`



SetExternalAddr sets ExternalAddr field to given value.



### HasExternalAddr



`func (o *V1NodeInfo) HasExternalAddr() bool`



HasExternalAddr returns a boolean if a field has been set.



### GetGrpcAddress



`func (o *V1NodeInfo) GetGrpcAddress() string`



GetGrpcAddress returns the GrpcAddress field if non-nil, zero value otherwise.



### GetGrpcAddressOk



`func (o *V1NodeInfo) GetGrpcAddressOk() (*string, bool)`



GetGrpcAddressOk returns a tuple with the GrpcAddress field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetGrpcAddress



`func (o *V1NodeInfo) SetGrpcAddress(v string)`



SetGrpcAddress sets GrpcAddress field to given value.



### HasGrpcAddress



`func (o *V1NodeInfo) HasGrpcAddress() bool`



HasGrpcAddress returns a boolean if a field has been set.



### GetAuthToken



`func (o *V1NodeInfo) GetAuthToken() string`



GetAuthToken returns the AuthToken field if non-nil, zero value otherwise.



### GetAuthTokenOk



`func (o *V1NodeInfo) GetAuthTokenOk() (*string, bool)`



GetAuthTokenOk returns a tuple with the AuthToken field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetAuthToken



`func (o *V1NodeInfo) SetAuthToken(v string)`



SetAuthToken sets AuthToken field to given value.



### HasAuthToken



`func (o *V1NodeInfo) HasAuthToken() bool`



HasAuthToken returns a boolean if a field has been set.



### GetStatus



`func (o *V1NodeInfo) GetStatus() string`



GetStatus returns the Status field if non-nil, zero value otherwise.



### GetStatusOk



`func (o *V1NodeInfo) GetStatusOk() (*string, bool)`



GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetStatus



`func (o *V1NodeInfo) SetStatus(v string)`



SetStatus sets Status field to given value.



### HasStatus



`func (o *V1NodeInfo) HasStatus() bool`



HasStatus returns a boolean if a field has been set.



### GetRegisteredAt



`func (o *V1NodeInfo) GetRegisteredAt() string`



GetRegisteredAt returns the RegisteredAt field if non-nil, zero value otherwise.



### GetRegisteredAtOk



`func (o *V1NodeInfo) GetRegisteredAtOk() (*string, bool)`



GetRegisteredAtOk returns a tuple with the RegisteredAt field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetRegisteredAt



`func (o *V1NodeInfo) SetRegisteredAt(v string)`



SetRegisteredAt sets RegisteredAt field to given value.



### HasRegisteredAt



`func (o *V1NodeInfo) HasRegisteredAt() bool`



HasRegisteredAt returns a boolean if a field has been set.



### GetLastHeartbeat



`func (o *V1NodeInfo) GetLastHeartbeat() string`



GetLastHeartbeat returns the LastHeartbeat field if non-nil, zero value otherwise.



### GetLastHeartbeatOk



`func (o *V1NodeInfo) GetLastHeartbeatOk() (*string, bool)`



GetLastHeartbeatOk returns a tuple with the LastHeartbeat field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetLastHeartbeat



`func (o *V1NodeInfo) SetLastHeartbeat(v string)`



SetLastHeartbeat sets LastHeartbeat field to given value.



### HasLastHeartbeat



`func (o *V1NodeInfo) HasLastHeartbeat() bool`



HasLastHeartbeat returns a boolean if a field has been set.



### GetCapabilities



`func (o *V1NodeInfo) GetCapabilities() V1AgentCapabilities`



GetCapabilities returns the Capabilities field if non-nil, zero value otherwise.



### GetCapabilitiesOk



`func (o *V1NodeInfo) GetCapabilitiesOk() (*V1AgentCapabilities, bool)`



GetCapabilitiesOk returns a tuple with the Capabilities field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetCapabilities



`func (o *V1NodeInfo) SetCapabilities(v V1AgentCapabilities)`



SetCapabilities sets Capabilities field to given value.



### HasCapabilities



`func (o *V1NodeInfo) HasCapabilities() bool`



HasCapabilities returns a boolean if a field has been set.



### GetAgentStatus



`func (o *V1NodeInfo) GetAgentStatus() V1AgentStatus`



GetAgentStatus returns the AgentStatus field if non-nil, zero value otherwise.



### GetAgentStatusOk



`func (o *V1NodeInfo) GetAgentStatusOk() (*V1AgentStatus, bool)`



GetAgentStatusOk returns a tuple with the AgentStatus field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetAgentStatus



`func (o *V1NodeInfo) SetAgentStatus(v V1AgentStatus)`



SetAgentStatus sets AgentStatus field to given value.



### HasAgentStatus



`func (o *V1NodeInfo) HasAgentStatus() bool`



HasAgentStatus returns a boolean if a field has been set.



### GetMetadata



`func (o *V1NodeInfo) GetMetadata() V1AgentMetadata`



GetMetadata returns the Metadata field if non-nil, zero value otherwise.



### GetMetadataOk



`func (o *V1NodeInfo) GetMetadataOk() (*V1AgentMetadata, bool)`



GetMetadataOk returns a tuple with the Metadata field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetMetadata



`func (o *V1NodeInfo) SetMetadata(v V1AgentMetadata)`



SetMetadata sets Metadata field to given value.



### HasMetadata



`func (o *V1NodeInfo) HasMetadata() bool`



HasMetadata returns a boolean if a field has been set.





[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
