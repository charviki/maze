# ConfigServiceUpdateConfigBody



## Properties



Name | Type | Description | Notes

------------ | ------------- | ------------- | -------------

**WorkingDir** | Pointer to **string** |  | [optional]

**Env** | Pointer to **map[string]string** |  | [optional]



## Methods



### NewConfigServiceUpdateConfigBody



`func NewConfigServiceUpdateConfigBody() *ConfigServiceUpdateConfigBody`



NewConfigServiceUpdateConfigBody instantiates a new ConfigServiceUpdateConfigBody object

This constructor will assign default values to properties that have it defined,

and makes sure properties required by API are set, but the set of arguments

will change when the set of required properties is changed



### NewConfigServiceUpdateConfigBodyWithDefaults



`func NewConfigServiceUpdateConfigBodyWithDefaults() *ConfigServiceUpdateConfigBody`



NewConfigServiceUpdateConfigBodyWithDefaults instantiates a new ConfigServiceUpdateConfigBody object

This constructor will only assign default values to properties that have it defined,

but it doesn't guarantee that properties required by API are set



### GetWorkingDir



`func (o *ConfigServiceUpdateConfigBody) GetWorkingDir() string`



GetWorkingDir returns the WorkingDir field if non-nil, zero value otherwise.



### GetWorkingDirOk



`func (o *ConfigServiceUpdateConfigBody) GetWorkingDirOk() (*string, bool)`



GetWorkingDirOk returns a tuple with the WorkingDir field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetWorkingDir



`func (o *ConfigServiceUpdateConfigBody) SetWorkingDir(v string)`



SetWorkingDir sets WorkingDir field to given value.



### HasWorkingDir



`func (o *ConfigServiceUpdateConfigBody) HasWorkingDir() bool`



HasWorkingDir returns a boolean if a field has been set.



### GetEnv



`func (o *ConfigServiceUpdateConfigBody) GetEnv() map[string]string`



GetEnv returns the Env field if non-nil, zero value otherwise.



### GetEnvOk



`func (o *ConfigServiceUpdateConfigBody) GetEnvOk() (*map[string]string, bool)`



GetEnvOk returns a tuple with the Env field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetEnv



`func (o *ConfigServiceUpdateConfigBody) SetEnv(v map[string]string)`



SetEnv sets Env field to given value.



### HasEnv



`func (o *ConfigServiceUpdateConfigBody) HasEnv() bool`



HasEnv returns a boolean if a field has been set.





[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
