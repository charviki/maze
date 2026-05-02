# V1ListHostsResponse



## Properties



Name | Type | Description | Notes

------------ | ------------- | ------------- | -------------

**Hosts** | Pointer to [**[]V1HostInfo**](V1HostInfo.md) |  | [optional]



## Methods



### NewV1ListHostsResponse



`func NewV1ListHostsResponse() *V1ListHostsResponse`



NewV1ListHostsResponse instantiates a new V1ListHostsResponse object

This constructor will assign default values to properties that have it defined,

and makes sure properties required by API are set, but the set of arguments

will change when the set of required properties is changed



### NewV1ListHostsResponseWithDefaults



`func NewV1ListHostsResponseWithDefaults() *V1ListHostsResponse`



NewV1ListHostsResponseWithDefaults instantiates a new V1ListHostsResponse object

This constructor will only assign default values to properties that have it defined,

but it doesn't guarantee that properties required by API are set



### GetHosts



`func (o *V1ListHostsResponse) GetHosts() []V1HostInfo`



GetHosts returns the Hosts field if non-nil, zero value otherwise.



### GetHostsOk



`func (o *V1ListHostsResponse) GetHostsOk() (*[]V1HostInfo, bool)`



GetHostsOk returns a tuple with the Hosts field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetHosts



`func (o *V1ListHostsResponse) SetHosts(v []V1HostInfo)`



SetHosts sets Hosts field to given value.



### HasHosts



`func (o *V1ListHostsResponse) HasHosts() bool`



HasHosts returns a boolean if a field has been set.





[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
