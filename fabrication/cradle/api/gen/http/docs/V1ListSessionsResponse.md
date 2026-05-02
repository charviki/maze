# V1ListSessionsResponse



## Properties



Name | Type | Description | Notes

------------ | ------------- | ------------- | -------------

**Sessions** | Pointer to [**[]V1Session**](V1Session.md) |  | [optional]



## Methods



### NewV1ListSessionsResponse



`func NewV1ListSessionsResponse() *V1ListSessionsResponse`



NewV1ListSessionsResponse instantiates a new V1ListSessionsResponse object

This constructor will assign default values to properties that have it defined,

and makes sure properties required by API are set, but the set of arguments

will change when the set of required properties is changed



### NewV1ListSessionsResponseWithDefaults



`func NewV1ListSessionsResponseWithDefaults() *V1ListSessionsResponse`



NewV1ListSessionsResponseWithDefaults instantiates a new V1ListSessionsResponse object

This constructor will only assign default values to properties that have it defined,

but it doesn't guarantee that properties required by API are set



### GetSessions



`func (o *V1ListSessionsResponse) GetSessions() []V1Session`



GetSessions returns the Sessions field if non-nil, zero value otherwise.



### GetSessionsOk



`func (o *V1ListSessionsResponse) GetSessionsOk() (*[]V1Session, bool)`



GetSessionsOk returns a tuple with the Sessions field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetSessions



`func (o *V1ListSessionsResponse) SetSessions(v []V1Session)`



SetSessions sets Sessions field to given value.



### HasSessions



`func (o *V1ListSessionsResponse) HasSessions() bool`



HasSessions returns a boolean if a field has been set.





[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
