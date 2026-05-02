# V1HostInfo



## Properties



Name | Type | Description | Notes

------------ | ------------- | ------------- | -------------

**Name** | Pointer to **string** |  | [optional]

**DisplayName** | Pointer to **string** |  | [optional]

**Tools** | Pointer to **[]string** |  | [optional]

**Resources** | Pointer to [**V1ResourceLimits**](V1ResourceLimits.md) |  | [optional]

**CreatedAt** | Pointer to **string** |  | [optional]

**UpdatedAt** | Pointer to **string** |  | [optional]

**Status** | Pointer to **string** |  | [optional]

**ErrorMsg** | Pointer to **string** |  | [optional]

**RetryCount** | Pointer to **int32** |  | [optional]

**Address** | Pointer to **string** |  | [optional]

**SessionCount** | Pointer to **int32** |  | [optional]

**LastHeartbeat** | Pointer to **string** |  | [optional]



## Methods



### NewV1HostInfo



`func NewV1HostInfo() *V1HostInfo`



NewV1HostInfo instantiates a new V1HostInfo object

This constructor will assign default values to properties that have it defined,

and makes sure properties required by API are set, but the set of arguments

will change when the set of required properties is changed



### NewV1HostInfoWithDefaults



`func NewV1HostInfoWithDefaults() *V1HostInfo`



NewV1HostInfoWithDefaults instantiates a new V1HostInfo object

This constructor will only assign default values to properties that have it defined,

but it doesn't guarantee that properties required by API are set



### GetName



`func (o *V1HostInfo) GetName() string`



GetName returns the Name field if non-nil, zero value otherwise.



### GetNameOk



`func (o *V1HostInfo) GetNameOk() (*string, bool)`



GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetName



`func (o *V1HostInfo) SetName(v string)`



SetName sets Name field to given value.



### HasName



`func (o *V1HostInfo) HasName() bool`



HasName returns a boolean if a field has been set.



### GetDisplayName



`func (o *V1HostInfo) GetDisplayName() string`



GetDisplayName returns the DisplayName field if non-nil, zero value otherwise.



### GetDisplayNameOk



`func (o *V1HostInfo) GetDisplayNameOk() (*string, bool)`



GetDisplayNameOk returns a tuple with the DisplayName field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetDisplayName



`func (o *V1HostInfo) SetDisplayName(v string)`



SetDisplayName sets DisplayName field to given value.



### HasDisplayName



`func (o *V1HostInfo) HasDisplayName() bool`



HasDisplayName returns a boolean if a field has been set.



### GetTools



`func (o *V1HostInfo) GetTools() []string`



GetTools returns the Tools field if non-nil, zero value otherwise.



### GetToolsOk



`func (o *V1HostInfo) GetToolsOk() (*[]string, bool)`



GetToolsOk returns a tuple with the Tools field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetTools



`func (o *V1HostInfo) SetTools(v []string)`



SetTools sets Tools field to given value.



### HasTools



`func (o *V1HostInfo) HasTools() bool`



HasTools returns a boolean if a field has been set.



### GetResources



`func (o *V1HostInfo) GetResources() V1ResourceLimits`



GetResources returns the Resources field if non-nil, zero value otherwise.



### GetResourcesOk



`func (o *V1HostInfo) GetResourcesOk() (*V1ResourceLimits, bool)`



GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetResources



`func (o *V1HostInfo) SetResources(v V1ResourceLimits)`



SetResources sets Resources field to given value.



### HasResources



`func (o *V1HostInfo) HasResources() bool`



HasResources returns a boolean if a field has been set.



### GetCreatedAt



`func (o *V1HostInfo) GetCreatedAt() string`



GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.



### GetCreatedAtOk



`func (o *V1HostInfo) GetCreatedAtOk() (*string, bool)`



GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetCreatedAt



`func (o *V1HostInfo) SetCreatedAt(v string)`



SetCreatedAt sets CreatedAt field to given value.



### HasCreatedAt



`func (o *V1HostInfo) HasCreatedAt() bool`



HasCreatedAt returns a boolean if a field has been set.



### GetUpdatedAt



`func (o *V1HostInfo) GetUpdatedAt() string`



GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.



### GetUpdatedAtOk



`func (o *V1HostInfo) GetUpdatedAtOk() (*string, bool)`



GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetUpdatedAt



`func (o *V1HostInfo) SetUpdatedAt(v string)`



SetUpdatedAt sets UpdatedAt field to given value.



### HasUpdatedAt



`func (o *V1HostInfo) HasUpdatedAt() bool`



HasUpdatedAt returns a boolean if a field has been set.



### GetStatus



`func (o *V1HostInfo) GetStatus() string`



GetStatus returns the Status field if non-nil, zero value otherwise.



### GetStatusOk



`func (o *V1HostInfo) GetStatusOk() (*string, bool)`



GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetStatus



`func (o *V1HostInfo) SetStatus(v string)`



SetStatus sets Status field to given value.



### HasStatus



`func (o *V1HostInfo) HasStatus() bool`



HasStatus returns a boolean if a field has been set.



### GetErrorMsg



`func (o *V1HostInfo) GetErrorMsg() string`



GetErrorMsg returns the ErrorMsg field if non-nil, zero value otherwise.



### GetErrorMsgOk



`func (o *V1HostInfo) GetErrorMsgOk() (*string, bool)`



GetErrorMsgOk returns a tuple with the ErrorMsg field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetErrorMsg



`func (o *V1HostInfo) SetErrorMsg(v string)`



SetErrorMsg sets ErrorMsg field to given value.



### HasErrorMsg



`func (o *V1HostInfo) HasErrorMsg() bool`



HasErrorMsg returns a boolean if a field has been set.



### GetRetryCount



`func (o *V1HostInfo) GetRetryCount() int32`



GetRetryCount returns the RetryCount field if non-nil, zero value otherwise.



### GetRetryCountOk



`func (o *V1HostInfo) GetRetryCountOk() (*int32, bool)`



GetRetryCountOk returns a tuple with the RetryCount field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetRetryCount



`func (o *V1HostInfo) SetRetryCount(v int32)`



SetRetryCount sets RetryCount field to given value.



### HasRetryCount



`func (o *V1HostInfo) HasRetryCount() bool`



HasRetryCount returns a boolean if a field has been set.



### GetAddress



`func (o *V1HostInfo) GetAddress() string`



GetAddress returns the Address field if non-nil, zero value otherwise.



### GetAddressOk



`func (o *V1HostInfo) GetAddressOk() (*string, bool)`



GetAddressOk returns a tuple with the Address field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetAddress



`func (o *V1HostInfo) SetAddress(v string)`



SetAddress sets Address field to given value.



### HasAddress



`func (o *V1HostInfo) HasAddress() bool`



HasAddress returns a boolean if a field has been set.



### GetSessionCount



`func (o *V1HostInfo) GetSessionCount() int32`



GetSessionCount returns the SessionCount field if non-nil, zero value otherwise.



### GetSessionCountOk



`func (o *V1HostInfo) GetSessionCountOk() (*int32, bool)`



GetSessionCountOk returns a tuple with the SessionCount field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetSessionCount



`func (o *V1HostInfo) SetSessionCount(v int32)`



SetSessionCount sets SessionCount field to given value.



### HasSessionCount



`func (o *V1HostInfo) HasSessionCount() bool`



HasSessionCount returns a boolean if a field has been set.



### GetLastHeartbeat



`func (o *V1HostInfo) GetLastHeartbeat() string`



GetLastHeartbeat returns the LastHeartbeat field if non-nil, zero value otherwise.



### GetLastHeartbeatOk



`func (o *V1HostInfo) GetLastHeartbeatOk() (*string, bool)`



GetLastHeartbeatOk returns a tuple with the LastHeartbeat field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetLastHeartbeat



`func (o *V1HostInfo) SetLastHeartbeat(v string)`



SetLastHeartbeat sets LastHeartbeat field to given value.



### HasLastHeartbeat



`func (o *V1HostInfo) HasLastHeartbeat() bool`



HasLastHeartbeat returns a boolean if a field has been set.





[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
