# V1TemplateConfigView



## Properties



Name | Type | Description | Notes

------------ | ------------- | ------------- | -------------

**TemplateId** | Pointer to **string** |  | [optional]

**Scope** | Pointer to **string** |  | [optional]

**Files** | Pointer to [**[]V1ConfigFileSnapshot**](V1ConfigFileSnapshot.md) |  | [optional]



## Methods



### NewV1TemplateConfigView



`func NewV1TemplateConfigView() *V1TemplateConfigView`



NewV1TemplateConfigView instantiates a new V1TemplateConfigView object

This constructor will assign default values to properties that have it defined,

and makes sure properties required by API are set, but the set of arguments

will change when the set of required properties is changed



### NewV1TemplateConfigViewWithDefaults



`func NewV1TemplateConfigViewWithDefaults() *V1TemplateConfigView`



NewV1TemplateConfigViewWithDefaults instantiates a new V1TemplateConfigView object

This constructor will only assign default values to properties that have it defined,

but it doesn't guarantee that properties required by API are set



### GetTemplateId



`func (o *V1TemplateConfigView) GetTemplateId() string`



GetTemplateId returns the TemplateId field if non-nil, zero value otherwise.



### GetTemplateIdOk



`func (o *V1TemplateConfigView) GetTemplateIdOk() (*string, bool)`



GetTemplateIdOk returns a tuple with the TemplateId field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetTemplateId



`func (o *V1TemplateConfigView) SetTemplateId(v string)`



SetTemplateId sets TemplateId field to given value.



### HasTemplateId



`func (o *V1TemplateConfigView) HasTemplateId() bool`



HasTemplateId returns a boolean if a field has been set.



### GetScope



`func (o *V1TemplateConfigView) GetScope() string`



GetScope returns the Scope field if non-nil, zero value otherwise.



### GetScopeOk



`func (o *V1TemplateConfigView) GetScopeOk() (*string, bool)`



GetScopeOk returns a tuple with the Scope field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetScope



`func (o *V1TemplateConfigView) SetScope(v string)`



SetScope sets Scope field to given value.



### HasScope



`func (o *V1TemplateConfigView) HasScope() bool`



HasScope returns a boolean if a field has been set.



### GetFiles



`func (o *V1TemplateConfigView) GetFiles() []V1ConfigFileSnapshot`



GetFiles returns the Files field if non-nil, zero value otherwise.



### GetFilesOk



`func (o *V1TemplateConfigView) GetFilesOk() (*[]V1ConfigFileSnapshot, bool)`



GetFilesOk returns a tuple with the Files field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetFiles



`func (o *V1TemplateConfigView) SetFiles(v []V1ConfigFileSnapshot)`



SetFiles sets Files field to given value.



### HasFiles



`func (o *V1TemplateConfigView) HasFiles() bool`



HasFiles returns a boolean if a field has been set.





[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
