# V1ListToolsResponse



## Properties



Name | Type | Description | Notes

------------ | ------------- | ------------- | -------------

**Tools** | Pointer to [**[]V1ToolConfig**](V1ToolConfig.md) |  | [optional]



## Methods



### NewV1ListToolsResponse



`func NewV1ListToolsResponse() *V1ListToolsResponse`



NewV1ListToolsResponse instantiates a new V1ListToolsResponse object

This constructor will assign default values to properties that have it defined,

and makes sure properties required by API are set, but the set of arguments

will change when the set of required properties is changed



### NewV1ListToolsResponseWithDefaults



`func NewV1ListToolsResponseWithDefaults() *V1ListToolsResponse`



NewV1ListToolsResponseWithDefaults instantiates a new V1ListToolsResponse object

This constructor will only assign default values to properties that have it defined,

but it doesn't guarantee that properties required by API are set



### GetTools



`func (o *V1ListToolsResponse) GetTools() []V1ToolConfig`



GetTools returns the Tools field if non-nil, zero value otherwise.



### GetToolsOk



`func (o *V1ListToolsResponse) GetToolsOk() (*[]V1ToolConfig, bool)`



GetToolsOk returns a tuple with the Tools field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetTools



`func (o *V1ListToolsResponse) SetTools(v []V1ToolConfig)`



SetTools sets Tools field to given value.



### HasTools



`func (o *V1ListToolsResponse) HasTools() bool`



HasTools returns a boolean if a field has been set.





[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
