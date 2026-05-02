# SessionServiceUpdateSessionConfigBody



## Properties



Name | Type | Description | Notes

------------ | ------------- | ------------- | -------------

**Files** | Pointer to [**[]V1ConfigFileUpdate**](V1ConfigFileUpdate.md) |  | [optional]



## Methods



### NewSessionServiceUpdateSessionConfigBody



`func NewSessionServiceUpdateSessionConfigBody() *SessionServiceUpdateSessionConfigBody`



NewSessionServiceUpdateSessionConfigBody instantiates a new SessionServiceUpdateSessionConfigBody object

This constructor will assign default values to properties that have it defined,

and makes sure properties required by API are set, but the set of arguments

will change when the set of required properties is changed



### NewSessionServiceUpdateSessionConfigBodyWithDefaults



`func NewSessionServiceUpdateSessionConfigBodyWithDefaults() *SessionServiceUpdateSessionConfigBody`



NewSessionServiceUpdateSessionConfigBodyWithDefaults instantiates a new SessionServiceUpdateSessionConfigBody object

This constructor will only assign default values to properties that have it defined,

but it doesn't guarantee that properties required by API are set



### GetFiles



`func (o *SessionServiceUpdateSessionConfigBody) GetFiles() []V1ConfigFileUpdate`



GetFiles returns the Files field if non-nil, zero value otherwise.



### GetFilesOk



`func (o *SessionServiceUpdateSessionConfigBody) GetFilesOk() (*[]V1ConfigFileUpdate, bool)`



GetFilesOk returns a tuple with the Files field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetFiles



`func (o *SessionServiceUpdateSessionConfigBody) SetFiles(v []V1ConfigFileUpdate)`



SetFiles sets Files field to given value.



### HasFiles



`func (o *SessionServiceUpdateSessionConfigBody) HasFiles() bool`



HasFiles returns a boolean if a field has been set.





[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
