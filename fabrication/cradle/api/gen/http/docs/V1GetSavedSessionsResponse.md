# V1GetSavedSessionsResponse



## Properties



Name | Type | Description | Notes

------------ | ------------- | ------------- | -------------

**Sessions** | Pointer to [**[]V1SessionState**](V1SessionState.md) |  | [optional]



## Methods



### NewV1GetSavedSessionsResponse



`func NewV1GetSavedSessionsResponse() *V1GetSavedSessionsResponse`



NewV1GetSavedSessionsResponse instantiates a new V1GetSavedSessionsResponse object

This constructor will assign default values to properties that have it defined,

and makes sure properties required by API are set, but the set of arguments

will change when the set of required properties is changed



### NewV1GetSavedSessionsResponseWithDefaults



`func NewV1GetSavedSessionsResponseWithDefaults() *V1GetSavedSessionsResponse`



NewV1GetSavedSessionsResponseWithDefaults instantiates a new V1GetSavedSessionsResponse object

This constructor will only assign default values to properties that have it defined,

but it doesn't guarantee that properties required by API are set



### GetSessions



`func (o *V1GetSavedSessionsResponse) GetSessions() []V1SessionState`



GetSessions returns the Sessions field if non-nil, zero value otherwise.



### GetSessionsOk



`func (o *V1GetSavedSessionsResponse) GetSessionsOk() (*[]V1SessionState, bool)`



GetSessionsOk returns a tuple with the Sessions field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetSessions



`func (o *V1GetSavedSessionsResponse) SetSessions(v []V1SessionState)`



SetSessions sets Sessions field to given value.



### HasSessions



`func (o *V1GetSavedSessionsResponse) HasSessions() bool`



HasSessions returns a boolean if a field has been set.





[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
