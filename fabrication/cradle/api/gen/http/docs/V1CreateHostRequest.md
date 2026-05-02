# V1CreateHostRequest



## Properties



Name | Type | Description | Notes

------------ | ------------- | ------------- | -------------

**Name** | Pointer to **string** |  | [optional]

**Tools** | Pointer to **[]string** |  | [optional]

**DisplayName** | Pointer to **string** |  | [optional]

**Resources** | Pointer to [**V1ResourceLimits**](V1ResourceLimits.md) |  | [optional]



## Methods



### NewV1CreateHostRequest



`func NewV1CreateHostRequest() *V1CreateHostRequest`



NewV1CreateHostRequest instantiates a new V1CreateHostRequest object

This constructor will assign default values to properties that have it defined,

and makes sure properties required by API are set, but the set of arguments

will change when the set of required properties is changed



### NewV1CreateHostRequestWithDefaults



`func NewV1CreateHostRequestWithDefaults() *V1CreateHostRequest`



NewV1CreateHostRequestWithDefaults instantiates a new V1CreateHostRequest object

This constructor will only assign default values to properties that have it defined,

but it doesn't guarantee that properties required by API are set



### GetName



`func (o *V1CreateHostRequest) GetName() string`



GetName returns the Name field if non-nil, zero value otherwise.



### GetNameOk



`func (o *V1CreateHostRequest) GetNameOk() (*string, bool)`



GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetName



`func (o *V1CreateHostRequest) SetName(v string)`



SetName sets Name field to given value.



### HasName



`func (o *V1CreateHostRequest) HasName() bool`



HasName returns a boolean if a field has been set.



### GetTools



`func (o *V1CreateHostRequest) GetTools() []string`



GetTools returns the Tools field if non-nil, zero value otherwise.



### GetToolsOk



`func (o *V1CreateHostRequest) GetToolsOk() (*[]string, bool)`



GetToolsOk returns a tuple with the Tools field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetTools



`func (o *V1CreateHostRequest) SetTools(v []string)`



SetTools sets Tools field to given value.



### HasTools



`func (o *V1CreateHostRequest) HasTools() bool`



HasTools returns a boolean if a field has been set.



### GetDisplayName



`func (o *V1CreateHostRequest) GetDisplayName() string`



GetDisplayName returns the DisplayName field if non-nil, zero value otherwise.



### GetDisplayNameOk



`func (o *V1CreateHostRequest) GetDisplayNameOk() (*string, bool)`



GetDisplayNameOk returns a tuple with the DisplayName field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetDisplayName



`func (o *V1CreateHostRequest) SetDisplayName(v string)`



SetDisplayName sets DisplayName field to given value.



### HasDisplayName



`func (o *V1CreateHostRequest) HasDisplayName() bool`



HasDisplayName returns a boolean if a field has been set.



### GetResources



`func (o *V1CreateHostRequest) GetResources() V1ResourceLimits`



GetResources returns the Resources field if non-nil, zero value otherwise.



### GetResourcesOk



`func (o *V1CreateHostRequest) GetResourcesOk() (*V1ResourceLimits, bool)`



GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetResources



`func (o *V1CreateHostRequest) SetResources(v V1ResourceLimits)`



SetResources sets Resources field to given value.



### HasResources



`func (o *V1CreateHostRequest) HasResources() bool`



HasResources returns a boolean if a field has been set.





[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
