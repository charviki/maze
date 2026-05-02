# V1SessionTemplate



## Properties



Name | Type | Description | Notes

------------ | ------------- | ------------- | -------------

**Id** | Pointer to **string** |  | [optional]

**Name** | Pointer to **string** |  | [optional]

**Command** | Pointer to **string** |  | [optional]

**RestoreCommand** | Pointer to **string** |  | [optional]

**SessionFilePattern** | Pointer to **string** |  | [optional]

**Description** | Pointer to **string** |  | [optional]

**Icon** | Pointer to **string** |  | [optional]

**Builtin** | Pointer to **bool** |  | [optional]

**Defaults** | Pointer to [**V1ConfigLayer**](V1ConfigLayer.md) |  | [optional]

**SessionSchema** | Pointer to [**V1SessionSchema**](V1SessionSchema.md) |  | [optional]



## Methods



### NewV1SessionTemplate



`func NewV1SessionTemplate() *V1SessionTemplate`



NewV1SessionTemplate instantiates a new V1SessionTemplate object

This constructor will assign default values to properties that have it defined,

and makes sure properties required by API are set, but the set of arguments

will change when the set of required properties is changed



### NewV1SessionTemplateWithDefaults



`func NewV1SessionTemplateWithDefaults() *V1SessionTemplate`



NewV1SessionTemplateWithDefaults instantiates a new V1SessionTemplate object

This constructor will only assign default values to properties that have it defined,

but it doesn't guarantee that properties required by API are set



### GetId



`func (o *V1SessionTemplate) GetId() string`



GetId returns the Id field if non-nil, zero value otherwise.



### GetIdOk



`func (o *V1SessionTemplate) GetIdOk() (*string, bool)`



GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetId



`func (o *V1SessionTemplate) SetId(v string)`



SetId sets Id field to given value.



### HasId



`func (o *V1SessionTemplate) HasId() bool`



HasId returns a boolean if a field has been set.



### GetName



`func (o *V1SessionTemplate) GetName() string`



GetName returns the Name field if non-nil, zero value otherwise.



### GetNameOk



`func (o *V1SessionTemplate) GetNameOk() (*string, bool)`



GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetName



`func (o *V1SessionTemplate) SetName(v string)`



SetName sets Name field to given value.



### HasName



`func (o *V1SessionTemplate) HasName() bool`



HasName returns a boolean if a field has been set.



### GetCommand



`func (o *V1SessionTemplate) GetCommand() string`



GetCommand returns the Command field if non-nil, zero value otherwise.



### GetCommandOk



`func (o *V1SessionTemplate) GetCommandOk() (*string, bool)`



GetCommandOk returns a tuple with the Command field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetCommand



`func (o *V1SessionTemplate) SetCommand(v string)`



SetCommand sets Command field to given value.



### HasCommand



`func (o *V1SessionTemplate) HasCommand() bool`



HasCommand returns a boolean if a field has been set.



### GetRestoreCommand



`func (o *V1SessionTemplate) GetRestoreCommand() string`



GetRestoreCommand returns the RestoreCommand field if non-nil, zero value otherwise.



### GetRestoreCommandOk



`func (o *V1SessionTemplate) GetRestoreCommandOk() (*string, bool)`



GetRestoreCommandOk returns a tuple with the RestoreCommand field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetRestoreCommand



`func (o *V1SessionTemplate) SetRestoreCommand(v string)`



SetRestoreCommand sets RestoreCommand field to given value.



### HasRestoreCommand



`func (o *V1SessionTemplate) HasRestoreCommand() bool`



HasRestoreCommand returns a boolean if a field has been set.



### GetSessionFilePattern



`func (o *V1SessionTemplate) GetSessionFilePattern() string`



GetSessionFilePattern returns the SessionFilePattern field if non-nil, zero value otherwise.



### GetSessionFilePatternOk



`func (o *V1SessionTemplate) GetSessionFilePatternOk() (*string, bool)`



GetSessionFilePatternOk returns a tuple with the SessionFilePattern field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetSessionFilePattern



`func (o *V1SessionTemplate) SetSessionFilePattern(v string)`



SetSessionFilePattern sets SessionFilePattern field to given value.



### HasSessionFilePattern



`func (o *V1SessionTemplate) HasSessionFilePattern() bool`



HasSessionFilePattern returns a boolean if a field has been set.



### GetDescription



`func (o *V1SessionTemplate) GetDescription() string`



GetDescription returns the Description field if non-nil, zero value otherwise.



### GetDescriptionOk



`func (o *V1SessionTemplate) GetDescriptionOk() (*string, bool)`



GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetDescription



`func (o *V1SessionTemplate) SetDescription(v string)`



SetDescription sets Description field to given value.



### HasDescription



`func (o *V1SessionTemplate) HasDescription() bool`



HasDescription returns a boolean if a field has been set.



### GetIcon



`func (o *V1SessionTemplate) GetIcon() string`



GetIcon returns the Icon field if non-nil, zero value otherwise.



### GetIconOk



`func (o *V1SessionTemplate) GetIconOk() (*string, bool)`



GetIconOk returns a tuple with the Icon field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetIcon



`func (o *V1SessionTemplate) SetIcon(v string)`



SetIcon sets Icon field to given value.



### HasIcon



`func (o *V1SessionTemplate) HasIcon() bool`



HasIcon returns a boolean if a field has been set.



### GetBuiltin



`func (o *V1SessionTemplate) GetBuiltin() bool`



GetBuiltin returns the Builtin field if non-nil, zero value otherwise.



### GetBuiltinOk



`func (o *V1SessionTemplate) GetBuiltinOk() (*bool, bool)`



GetBuiltinOk returns a tuple with the Builtin field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetBuiltin



`func (o *V1SessionTemplate) SetBuiltin(v bool)`



SetBuiltin sets Builtin field to given value.



### HasBuiltin



`func (o *V1SessionTemplate) HasBuiltin() bool`



HasBuiltin returns a boolean if a field has been set.



### GetDefaults



`func (o *V1SessionTemplate) GetDefaults() V1ConfigLayer`



GetDefaults returns the Defaults field if non-nil, zero value otherwise.



### GetDefaultsOk



`func (o *V1SessionTemplate) GetDefaultsOk() (*V1ConfigLayer, bool)`



GetDefaultsOk returns a tuple with the Defaults field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetDefaults



`func (o *V1SessionTemplate) SetDefaults(v V1ConfigLayer)`



SetDefaults sets Defaults field to given value.



### HasDefaults



`func (o *V1SessionTemplate) HasDefaults() bool`



HasDefaults returns a boolean if a field has been set.



### GetSessionSchema



`func (o *V1SessionTemplate) GetSessionSchema() V1SessionSchema`



GetSessionSchema returns the SessionSchema field if non-nil, zero value otherwise.



### GetSessionSchemaOk



`func (o *V1SessionTemplate) GetSessionSchemaOk() (*V1SessionSchema, bool)`



GetSessionSchemaOk returns a tuple with the SessionSchema field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetSessionSchema



`func (o *V1SessionTemplate) SetSessionSchema(v V1SessionSchema)`



SetSessionSchema sets SessionSchema field to given value.



### HasSessionSchema



`func (o *V1SessionTemplate) HasSessionSchema() bool`



HasSessionSchema returns a boolean if a field has been set.





[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
