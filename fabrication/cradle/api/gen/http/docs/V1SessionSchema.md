# V1SessionSchema



## Properties



Name | Type | Description | Notes

------------ | ------------- | ------------- | -------------

**EnvDefs** | Pointer to [**[]V1EnvDef**](V1EnvDef.md) |  | [optional]

**FileDefs** | Pointer to [**[]V1FileDef**](V1FileDef.md) |  | [optional]



## Methods



### NewV1SessionSchema



`func NewV1SessionSchema() *V1SessionSchema`



NewV1SessionSchema instantiates a new V1SessionSchema object

This constructor will assign default values to properties that have it defined,

and makes sure properties required by API are set, but the set of arguments

will change when the set of required properties is changed



### NewV1SessionSchemaWithDefaults



`func NewV1SessionSchemaWithDefaults() *V1SessionSchema`



NewV1SessionSchemaWithDefaults instantiates a new V1SessionSchema object

This constructor will only assign default values to properties that have it defined,

but it doesn't guarantee that properties required by API are set



### GetEnvDefs



`func (o *V1SessionSchema) GetEnvDefs() []V1EnvDef`



GetEnvDefs returns the EnvDefs field if non-nil, zero value otherwise.



### GetEnvDefsOk



`func (o *V1SessionSchema) GetEnvDefsOk() (*[]V1EnvDef, bool)`



GetEnvDefsOk returns a tuple with the EnvDefs field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetEnvDefs



`func (o *V1SessionSchema) SetEnvDefs(v []V1EnvDef)`



SetEnvDefs sets EnvDefs field to given value.



### HasEnvDefs



`func (o *V1SessionSchema) HasEnvDefs() bool`



HasEnvDefs returns a boolean if a field has been set.



### GetFileDefs



`func (o *V1SessionSchema) GetFileDefs() []V1FileDef`



GetFileDefs returns the FileDefs field if non-nil, zero value otherwise.



### GetFileDefsOk



`func (o *V1SessionSchema) GetFileDefsOk() (*[]V1FileDef, bool)`



GetFileDefsOk returns a tuple with the FileDefs field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetFileDefs



`func (o *V1SessionSchema) SetFileDefs(v []V1FileDef)`



SetFileDefs sets FileDefs field to given value.



### HasFileDefs



`func (o *V1SessionSchema) HasFileDefs() bool`



HasFileDefs returns a boolean if a field has been set.





[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
