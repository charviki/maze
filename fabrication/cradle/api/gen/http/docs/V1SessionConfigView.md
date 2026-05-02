# V1SessionConfigView



## Properties



Name | Type | Description | Notes

------------ | ------------- | ------------- | -------------

**SessionId** | Pointer to **string** |  | [optional]

**TemplateId** | Pointer to **string** |  | [optional]

**WorkingDir** | Pointer to **string** |  | [optional]

**Scope** | Pointer to **string** |  | [optional]

**Files** | Pointer to [**[]V1ConfigFileSnapshot**](V1ConfigFileSnapshot.md) |  | [optional]



## Methods



### NewV1SessionConfigView



`func NewV1SessionConfigView() *V1SessionConfigView`



NewV1SessionConfigView instantiates a new V1SessionConfigView object

This constructor will assign default values to properties that have it defined,

and makes sure properties required by API are set, but the set of arguments

will change when the set of required properties is changed



### NewV1SessionConfigViewWithDefaults



`func NewV1SessionConfigViewWithDefaults() *V1SessionConfigView`



NewV1SessionConfigViewWithDefaults instantiates a new V1SessionConfigView object

This constructor will only assign default values to properties that have it defined,

but it doesn't guarantee that properties required by API are set



### GetSessionId



`func (o *V1SessionConfigView) GetSessionId() string`



GetSessionId returns the SessionId field if non-nil, zero value otherwise.



### GetSessionIdOk



`func (o *V1SessionConfigView) GetSessionIdOk() (*string, bool)`



GetSessionIdOk returns a tuple with the SessionId field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetSessionId



`func (o *V1SessionConfigView) SetSessionId(v string)`



SetSessionId sets SessionId field to given value.



### HasSessionId



`func (o *V1SessionConfigView) HasSessionId() bool`



HasSessionId returns a boolean if a field has been set.



### GetTemplateId



`func (o *V1SessionConfigView) GetTemplateId() string`



GetTemplateId returns the TemplateId field if non-nil, zero value otherwise.



### GetTemplateIdOk



`func (o *V1SessionConfigView) GetTemplateIdOk() (*string, bool)`



GetTemplateIdOk returns a tuple with the TemplateId field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetTemplateId



`func (o *V1SessionConfigView) SetTemplateId(v string)`



SetTemplateId sets TemplateId field to given value.



### HasTemplateId



`func (o *V1SessionConfigView) HasTemplateId() bool`



HasTemplateId returns a boolean if a field has been set.



### GetWorkingDir



`func (o *V1SessionConfigView) GetWorkingDir() string`



GetWorkingDir returns the WorkingDir field if non-nil, zero value otherwise.



### GetWorkingDirOk



`func (o *V1SessionConfigView) GetWorkingDirOk() (*string, bool)`



GetWorkingDirOk returns a tuple with the WorkingDir field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetWorkingDir



`func (o *V1SessionConfigView) SetWorkingDir(v string)`



SetWorkingDir sets WorkingDir field to given value.



### HasWorkingDir



`func (o *V1SessionConfigView) HasWorkingDir() bool`



HasWorkingDir returns a boolean if a field has been set.



### GetScope



`func (o *V1SessionConfigView) GetScope() string`



GetScope returns the Scope field if non-nil, zero value otherwise.



### GetScopeOk



`func (o *V1SessionConfigView) GetScopeOk() (*string, bool)`



GetScopeOk returns a tuple with the Scope field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetScope



`func (o *V1SessionConfigView) SetScope(v string)`



SetScope sets Scope field to given value.



### HasScope



`func (o *V1SessionConfigView) HasScope() bool`



HasScope returns a boolean if a field has been set.



### GetFiles



`func (o *V1SessionConfigView) GetFiles() []V1ConfigFileSnapshot`



GetFiles returns the Files field if non-nil, zero value otherwise.



### GetFilesOk



`func (o *V1SessionConfigView) GetFilesOk() (*[]V1ConfigFileSnapshot, bool)`



GetFilesOk returns a tuple with the Files field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetFiles



`func (o *V1SessionConfigView) SetFiles(v []V1ConfigFileSnapshot)`



SetFiles sets Files field to given value.



### HasFiles



`func (o *V1SessionConfigView) HasFiles() bool`



HasFiles returns a boolean if a field has been set.





[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
