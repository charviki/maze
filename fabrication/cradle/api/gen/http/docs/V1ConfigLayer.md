# V1ConfigLayer



## Properties



Name | Type | Description | Notes

------------ | ------------- | ------------- | -------------

**Env** | Pointer to **map[string]string** |  | [optional]

**Files** | Pointer to [**[]V1ConfigFile**](V1ConfigFile.md) |  | [optional]



## Methods



### NewV1ConfigLayer



`func NewV1ConfigLayer() *V1ConfigLayer`



NewV1ConfigLayer instantiates a new V1ConfigLayer object

This constructor will assign default values to properties that have it defined,

and makes sure properties required by API are set, but the set of arguments

will change when the set of required properties is changed



### NewV1ConfigLayerWithDefaults



`func NewV1ConfigLayerWithDefaults() *V1ConfigLayer`



NewV1ConfigLayerWithDefaults instantiates a new V1ConfigLayer object

This constructor will only assign default values to properties that have it defined,

but it doesn't guarantee that properties required by API are set



### GetEnv



`func (o *V1ConfigLayer) GetEnv() map[string]string`



GetEnv returns the Env field if non-nil, zero value otherwise.



### GetEnvOk



`func (o *V1ConfigLayer) GetEnvOk() (*map[string]string, bool)`



GetEnvOk returns a tuple with the Env field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetEnv



`func (o *V1ConfigLayer) SetEnv(v map[string]string)`



SetEnv sets Env field to given value.



### HasEnv



`func (o *V1ConfigLayer) HasEnv() bool`



HasEnv returns a boolean if a field has been set.



### GetFiles



`func (o *V1ConfigLayer) GetFiles() []V1ConfigFile`



GetFiles returns the Files field if non-nil, zero value otherwise.



### GetFilesOk



`func (o *V1ConfigLayer) GetFilesOk() (*[]V1ConfigFile, bool)`



GetFilesOk returns a tuple with the Files field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetFiles



`func (o *V1ConfigLayer) SetFiles(v []V1ConfigFile)`



SetFiles sets Files field to given value.



### HasFiles



`func (o *V1ConfigLayer) HasFiles() bool`



HasFiles returns a boolean if a field has been set.





[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
