# V1RegisterRequest



## Properties



Name | Type | Description | Notes

------------ | ------------- | ------------- | -------------

**Name** | Pointer to **string** |  | [optional]

**Address** | Pointer to **string** |  | [optional]

**ExternalAddr** | Pointer to **string** |  | [optional]

**GrpcAddress** | Pointer to **string** |  | [optional]

**Capabilities** | Pointer to [**V1AgentCapabilities**](V1AgentCapabilities.md) |  | [optional]

**Status** | Pointer to [**V1AgentStatus**](V1AgentStatus.md) |  | [optional]

**Metadata** | Pointer to [**V1AgentMetadata**](V1AgentMetadata.md) |  | [optional]



## Methods



### NewV1RegisterRequest



`func NewV1RegisterRequest() *V1RegisterRequest`



NewV1RegisterRequest instantiates a new V1RegisterRequest object

This constructor will assign default values to properties that have it defined,

and makes sure properties required by API are set, but the set of arguments

will change when the set of required properties is changed



### NewV1RegisterRequestWithDefaults



`func NewV1RegisterRequestWithDefaults() *V1RegisterRequest`



NewV1RegisterRequestWithDefaults instantiates a new V1RegisterRequest object

This constructor will only assign default values to properties that have it defined,

but it doesn't guarantee that properties required by API are set



### GetName



`func (o *V1RegisterRequest) GetName() string`



GetName returns the Name field if non-nil, zero value otherwise.



### GetNameOk



`func (o *V1RegisterRequest) GetNameOk() (*string, bool)`



GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetName



`func (o *V1RegisterRequest) SetName(v string)`



SetName sets Name field to given value.



### HasName



`func (o *V1RegisterRequest) HasName() bool`



HasName returns a boolean if a field has been set.



### GetAddress



`func (o *V1RegisterRequest) GetAddress() string`



GetAddress returns the Address field if non-nil, zero value otherwise.



### GetAddressOk



`func (o *V1RegisterRequest) GetAddressOk() (*string, bool)`



GetAddressOk returns a tuple with the Address field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetAddress



`func (o *V1RegisterRequest) SetAddress(v string)`



SetAddress sets Address field to given value.



### HasAddress



`func (o *V1RegisterRequest) HasAddress() bool`



HasAddress returns a boolean if a field has been set.



### GetExternalAddr



`func (o *V1RegisterRequest) GetExternalAddr() string`



GetExternalAddr returns the ExternalAddr field if non-nil, zero value otherwise.



### GetExternalAddrOk



`func (o *V1RegisterRequest) GetExternalAddrOk() (*string, bool)`



GetExternalAddrOk returns a tuple with the ExternalAddr field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetExternalAddr



`func (o *V1RegisterRequest) SetExternalAddr(v string)`



SetExternalAddr sets ExternalAddr field to given value.



### HasExternalAddr



`func (o *V1RegisterRequest) HasExternalAddr() bool`



HasExternalAddr returns a boolean if a field has been set.



### GetGrpcAddress



`func (o *V1RegisterRequest) GetGrpcAddress() string`



GetGrpcAddress returns the GrpcAddress field if non-nil, zero value otherwise.



### GetGrpcAddressOk



`func (o *V1RegisterRequest) GetGrpcAddressOk() (*string, bool)`



GetGrpcAddressOk returns a tuple with the GrpcAddress field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetGrpcAddress



`func (o *V1RegisterRequest) SetGrpcAddress(v string)`



SetGrpcAddress sets GrpcAddress field to given value.



### HasGrpcAddress



`func (o *V1RegisterRequest) HasGrpcAddress() bool`



HasGrpcAddress returns a boolean if a field has been set.



### GetCapabilities



`func (o *V1RegisterRequest) GetCapabilities() V1AgentCapabilities`



GetCapabilities returns the Capabilities field if non-nil, zero value otherwise.



### GetCapabilitiesOk



`func (o *V1RegisterRequest) GetCapabilitiesOk() (*V1AgentCapabilities, bool)`



GetCapabilitiesOk returns a tuple with the Capabilities field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetCapabilities



`func (o *V1RegisterRequest) SetCapabilities(v V1AgentCapabilities)`



SetCapabilities sets Capabilities field to given value.



### HasCapabilities



`func (o *V1RegisterRequest) HasCapabilities() bool`



HasCapabilities returns a boolean if a field has been set.



### GetStatus



`func (o *V1RegisterRequest) GetStatus() V1AgentStatus`



GetStatus returns the Status field if non-nil, zero value otherwise.



### GetStatusOk



`func (o *V1RegisterRequest) GetStatusOk() (*V1AgentStatus, bool)`



GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetStatus



`func (o *V1RegisterRequest) SetStatus(v V1AgentStatus)`



SetStatus sets Status field to given value.



### HasStatus



`func (o *V1RegisterRequest) HasStatus() bool`



HasStatus returns a boolean if a field has been set.



### GetMetadata



`func (o *V1RegisterRequest) GetMetadata() V1AgentMetadata`



GetMetadata returns the Metadata field if non-nil, zero value otherwise.



### GetMetadataOk



`func (o *V1RegisterRequest) GetMetadataOk() (*V1AgentMetadata, bool)`



GetMetadataOk returns a tuple with the Metadata field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetMetadata



`func (o *V1RegisterRequest) SetMetadata(v V1AgentMetadata)`



SetMetadata sets Metadata field to given value.



### HasMetadata



`func (o *V1RegisterRequest) HasMetadata() bool`



HasMetadata returns a boolean if a field has been set.





[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
