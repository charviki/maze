# V1ListNodesResponse



## Properties



Name | Type | Description | Notes

------------ | ------------- | ------------- | -------------

**Nodes** | Pointer to [**[]V1NodeInfo**](V1NodeInfo.md) |  | [optional]



## Methods



### NewV1ListNodesResponse



`func NewV1ListNodesResponse() *V1ListNodesResponse`



NewV1ListNodesResponse instantiates a new V1ListNodesResponse object

This constructor will assign default values to properties that have it defined,

and makes sure properties required by API are set, but the set of arguments

will change when the set of required properties is changed



### NewV1ListNodesResponseWithDefaults



`func NewV1ListNodesResponseWithDefaults() *V1ListNodesResponse`



NewV1ListNodesResponseWithDefaults instantiates a new V1ListNodesResponse object

This constructor will only assign default values to properties that have it defined,

but it doesn't guarantee that properties required by API are set



### GetNodes



`func (o *V1ListNodesResponse) GetNodes() []V1NodeInfo`



GetNodes returns the Nodes field if non-nil, zero value otherwise.



### GetNodesOk



`func (o *V1ListNodesResponse) GetNodesOk() (*[]V1NodeInfo, bool)`



GetNodesOk returns a tuple with the Nodes field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetNodes



`func (o *V1ListNodesResponse) SetNodes(v []V1NodeInfo)`



SetNodes sets Nodes field to given value.



### HasNodes



`func (o *V1ListNodesResponse) HasNodes() bool`



HasNodes returns a boolean if a field has been set.





[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
