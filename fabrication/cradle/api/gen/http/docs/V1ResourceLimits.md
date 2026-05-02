# V1ResourceLimits



## Properties



Name | Type | Description | Notes

------------ | ------------- | ------------- | -------------

**CpuLimit** | Pointer to **string** |  | [optional]

**MemoryLimit** | Pointer to **string** |  | [optional]



## Methods



### NewV1ResourceLimits



`func NewV1ResourceLimits() *V1ResourceLimits`



NewV1ResourceLimits instantiates a new V1ResourceLimits object

This constructor will assign default values to properties that have it defined,

and makes sure properties required by API are set, but the set of arguments

will change when the set of required properties is changed



### NewV1ResourceLimitsWithDefaults



`func NewV1ResourceLimitsWithDefaults() *V1ResourceLimits`



NewV1ResourceLimitsWithDefaults instantiates a new V1ResourceLimits object

This constructor will only assign default values to properties that have it defined,

but it doesn't guarantee that properties required by API are set



### GetCpuLimit



`func (o *V1ResourceLimits) GetCpuLimit() string`



GetCpuLimit returns the CpuLimit field if non-nil, zero value otherwise.



### GetCpuLimitOk



`func (o *V1ResourceLimits) GetCpuLimitOk() (*string, bool)`



GetCpuLimitOk returns a tuple with the CpuLimit field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetCpuLimit



`func (o *V1ResourceLimits) SetCpuLimit(v string)`



SetCpuLimit sets CpuLimit field to given value.



### HasCpuLimit



`func (o *V1ResourceLimits) HasCpuLimit() bool`



HasCpuLimit returns a boolean if a field has been set.



### GetMemoryLimit



`func (o *V1ResourceLimits) GetMemoryLimit() string`



GetMemoryLimit returns the MemoryLimit field if non-nil, zero value otherwise.



### GetMemoryLimitOk



`func (o *V1ResourceLimits) GetMemoryLimitOk() (*string, bool)`



GetMemoryLimitOk returns a tuple with the MemoryLimit field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetMemoryLimit



`func (o *V1ResourceLimits) SetMemoryLimit(v string)`



SetMemoryLimit sets MemoryLimit field to given value.



### HasMemoryLimit



`func (o *V1ResourceLimits) HasMemoryLimit() bool`



HasMemoryLimit returns a boolean if a field has been set.





[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
