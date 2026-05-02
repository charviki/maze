# V1HostSpec



## Properties



Name | Type | Description | Notes

------------ | ------------- | ------------- | -------------

**Name** | Pointer to **string** |  | [optional]

**DisplayName** | Pointer to **string** |  | [optional]

**Tools** | Pointer to **[]string** |  | [optional]

**Resources** | Pointer to [**V1ResourceLimits**](V1ResourceLimits.md) |  | [optional]

**AuthToken** | Pointer to **string** |  | [optional]

**CreatedAt** | Pointer to **string** |  | [optional]

**UpdatedAt** | Pointer to **string** |  | [optional]

**Status** | Pointer to **string** |  | [optional]

**ErrorMsg** | Pointer to **string** |  | [optional]

**RetryCount** | Pointer to **int32** |  | [optional]



## Methods



### NewV1HostSpec



`func NewV1HostSpec() *V1HostSpec`



NewV1HostSpec instantiates a new V1HostSpec object

This constructor will assign default values to properties that have it defined,

and makes sure properties required by API are set, but the set of arguments

will change when the set of required properties is changed



### NewV1HostSpecWithDefaults



`func NewV1HostSpecWithDefaults() *V1HostSpec`



NewV1HostSpecWithDefaults instantiates a new V1HostSpec object

This constructor will only assign default values to properties that have it defined,

but it doesn't guarantee that properties required by API are set



### GetName



`func (o *V1HostSpec) GetName() string`



GetName returns the Name field if non-nil, zero value otherwise.



### GetNameOk



`func (o *V1HostSpec) GetNameOk() (*string, bool)`



GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetName



`func (o *V1HostSpec) SetName(v string)`



SetName sets Name field to given value.



### HasName



`func (o *V1HostSpec) HasName() bool`



HasName returns a boolean if a field has been set.



### GetDisplayName



`func (o *V1HostSpec) GetDisplayName() string`



GetDisplayName returns the DisplayName field if non-nil, zero value otherwise.



### GetDisplayNameOk



`func (o *V1HostSpec) GetDisplayNameOk() (*string, bool)`



GetDisplayNameOk returns a tuple with the DisplayName field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetDisplayName



`func (o *V1HostSpec) SetDisplayName(v string)`



SetDisplayName sets DisplayName field to given value.



### HasDisplayName



`func (o *V1HostSpec) HasDisplayName() bool`



HasDisplayName returns a boolean if a field has been set.



### GetTools



`func (o *V1HostSpec) GetTools() []string`



GetTools returns the Tools field if non-nil, zero value otherwise.



### GetToolsOk



`func (o *V1HostSpec) GetToolsOk() (*[]string, bool)`



GetToolsOk returns a tuple with the Tools field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetTools



`func (o *V1HostSpec) SetTools(v []string)`



SetTools sets Tools field to given value.



### HasTools



`func (o *V1HostSpec) HasTools() bool`



HasTools returns a boolean if a field has been set.



### GetResources



`func (o *V1HostSpec) GetResources() V1ResourceLimits`



GetResources returns the Resources field if non-nil, zero value otherwise.



### GetResourcesOk



`func (o *V1HostSpec) GetResourcesOk() (*V1ResourceLimits, bool)`



GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetResources



`func (o *V1HostSpec) SetResources(v V1ResourceLimits)`



SetResources sets Resources field to given value.



### HasResources



`func (o *V1HostSpec) HasResources() bool`



HasResources returns a boolean if a field has been set.



### GetAuthToken



`func (o *V1HostSpec) GetAuthToken() string`



GetAuthToken returns the AuthToken field if non-nil, zero value otherwise.



### GetAuthTokenOk



`func (o *V1HostSpec) GetAuthTokenOk() (*string, bool)`



GetAuthTokenOk returns a tuple with the AuthToken field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetAuthToken



`func (o *V1HostSpec) SetAuthToken(v string)`



SetAuthToken sets AuthToken field to given value.



### HasAuthToken



`func (o *V1HostSpec) HasAuthToken() bool`



HasAuthToken returns a boolean if a field has been set.



### GetCreatedAt



`func (o *V1HostSpec) GetCreatedAt() string`



GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.



### GetCreatedAtOk



`func (o *V1HostSpec) GetCreatedAtOk() (*string, bool)`



GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetCreatedAt



`func (o *V1HostSpec) SetCreatedAt(v string)`



SetCreatedAt sets CreatedAt field to given value.



### HasCreatedAt



`func (o *V1HostSpec) HasCreatedAt() bool`



HasCreatedAt returns a boolean if a field has been set.



### GetUpdatedAt



`func (o *V1HostSpec) GetUpdatedAt() string`



GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.



### GetUpdatedAtOk



`func (o *V1HostSpec) GetUpdatedAtOk() (*string, bool)`



GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetUpdatedAt



`func (o *V1HostSpec) SetUpdatedAt(v string)`



SetUpdatedAt sets UpdatedAt field to given value.



### HasUpdatedAt



`func (o *V1HostSpec) HasUpdatedAt() bool`



HasUpdatedAt returns a boolean if a field has been set.



### GetStatus



`func (o *V1HostSpec) GetStatus() string`



GetStatus returns the Status field if non-nil, zero value otherwise.



### GetStatusOk



`func (o *V1HostSpec) GetStatusOk() (*string, bool)`



GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetStatus



`func (o *V1HostSpec) SetStatus(v string)`



SetStatus sets Status field to given value.



### HasStatus



`func (o *V1HostSpec) HasStatus() bool`



HasStatus returns a boolean if a field has been set.



### GetErrorMsg



`func (o *V1HostSpec) GetErrorMsg() string`



GetErrorMsg returns the ErrorMsg field if non-nil, zero value otherwise.



### GetErrorMsgOk



`func (o *V1HostSpec) GetErrorMsgOk() (*string, bool)`



GetErrorMsgOk returns a tuple with the ErrorMsg field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetErrorMsg



`func (o *V1HostSpec) SetErrorMsg(v string)`



SetErrorMsg sets ErrorMsg field to given value.



### HasErrorMsg



`func (o *V1HostSpec) HasErrorMsg() bool`



HasErrorMsg returns a boolean if a field has been set.



### GetRetryCount



`func (o *V1HostSpec) GetRetryCount() int32`



GetRetryCount returns the RetryCount field if non-nil, zero value otherwise.



### GetRetryCountOk



`func (o *V1HostSpec) GetRetryCountOk() (*int32, bool)`



GetRetryCountOk returns a tuple with the RetryCount field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetRetryCount



`func (o *V1HostSpec) SetRetryCount(v int32)`



SetRetryCount sets RetryCount field to given value.



### HasRetryCount



`func (o *V1HostSpec) HasRetryCount() bool`



HasRetryCount returns a boolean if a field has been set.





[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
