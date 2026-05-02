# SessionServiceCreateSessionBody



## Properties



Name | Type | Description | Notes

------------ | ------------- | ------------- | -------------

**Name** | Pointer to **string** |  | [optional]

**Command** | Pointer to **string** |  | [optional]

**WorkingDir** | Pointer to **string** |  | [optional]

**SessionConfs** | Pointer to [**[]V1ConfigItem**](V1ConfigItem.md) |  | [optional]

**RestoreStrategy** | Pointer to **string** |  | [optional]

**TemplateId** | Pointer to **string** |  | [optional]



## Methods



### NewSessionServiceCreateSessionBody



`func NewSessionServiceCreateSessionBody() *SessionServiceCreateSessionBody`



NewSessionServiceCreateSessionBody instantiates a new SessionServiceCreateSessionBody object

This constructor will assign default values to properties that have it defined,

and makes sure properties required by API are set, but the set of arguments

will change when the set of required properties is changed



### NewSessionServiceCreateSessionBodyWithDefaults



`func NewSessionServiceCreateSessionBodyWithDefaults() *SessionServiceCreateSessionBody`



NewSessionServiceCreateSessionBodyWithDefaults instantiates a new SessionServiceCreateSessionBody object

This constructor will only assign default values to properties that have it defined,

but it doesn't guarantee that properties required by API are set



### GetName



`func (o *SessionServiceCreateSessionBody) GetName() string`



GetName returns the Name field if non-nil, zero value otherwise.



### GetNameOk



`func (o *SessionServiceCreateSessionBody) GetNameOk() (*string, bool)`



GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetName



`func (o *SessionServiceCreateSessionBody) SetName(v string)`



SetName sets Name field to given value.



### HasName



`func (o *SessionServiceCreateSessionBody) HasName() bool`



HasName returns a boolean if a field has been set.



### GetCommand



`func (o *SessionServiceCreateSessionBody) GetCommand() string`



GetCommand returns the Command field if non-nil, zero value otherwise.



### GetCommandOk



`func (o *SessionServiceCreateSessionBody) GetCommandOk() (*string, bool)`



GetCommandOk returns a tuple with the Command field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetCommand



`func (o *SessionServiceCreateSessionBody) SetCommand(v string)`



SetCommand sets Command field to given value.



### HasCommand



`func (o *SessionServiceCreateSessionBody) HasCommand() bool`



HasCommand returns a boolean if a field has been set.



### GetWorkingDir



`func (o *SessionServiceCreateSessionBody) GetWorkingDir() string`



GetWorkingDir returns the WorkingDir field if non-nil, zero value otherwise.



### GetWorkingDirOk



`func (o *SessionServiceCreateSessionBody) GetWorkingDirOk() (*string, bool)`



GetWorkingDirOk returns a tuple with the WorkingDir field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetWorkingDir



`func (o *SessionServiceCreateSessionBody) SetWorkingDir(v string)`



SetWorkingDir sets WorkingDir field to given value.



### HasWorkingDir



`func (o *SessionServiceCreateSessionBody) HasWorkingDir() bool`



HasWorkingDir returns a boolean if a field has been set.



### GetSessionConfs



`func (o *SessionServiceCreateSessionBody) GetSessionConfs() []V1ConfigItem`



GetSessionConfs returns the SessionConfs field if non-nil, zero value otherwise.



### GetSessionConfsOk



`func (o *SessionServiceCreateSessionBody) GetSessionConfsOk() (*[]V1ConfigItem, bool)`



GetSessionConfsOk returns a tuple with the SessionConfs field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetSessionConfs



`func (o *SessionServiceCreateSessionBody) SetSessionConfs(v []V1ConfigItem)`



SetSessionConfs sets SessionConfs field to given value.



### HasSessionConfs



`func (o *SessionServiceCreateSessionBody) HasSessionConfs() bool`



HasSessionConfs returns a boolean if a field has been set.



### GetRestoreStrategy



`func (o *SessionServiceCreateSessionBody) GetRestoreStrategy() string`



GetRestoreStrategy returns the RestoreStrategy field if non-nil, zero value otherwise.



### GetRestoreStrategyOk



`func (o *SessionServiceCreateSessionBody) GetRestoreStrategyOk() (*string, bool)`



GetRestoreStrategyOk returns a tuple with the RestoreStrategy field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetRestoreStrategy



`func (o *SessionServiceCreateSessionBody) SetRestoreStrategy(v string)`



SetRestoreStrategy sets RestoreStrategy field to given value.



### HasRestoreStrategy



`func (o *SessionServiceCreateSessionBody) HasRestoreStrategy() bool`



HasRestoreStrategy returns a boolean if a field has been set.



### GetTemplateId



`func (o *SessionServiceCreateSessionBody) GetTemplateId() string`



GetTemplateId returns the TemplateId field if non-nil, zero value otherwise.



### GetTemplateIdOk



`func (o *SessionServiceCreateSessionBody) GetTemplateIdOk() (*string, bool)`



GetTemplateIdOk returns a tuple with the TemplateId field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetTemplateId



`func (o *SessionServiceCreateSessionBody) SetTemplateId(v string)`



SetTemplateId sets TemplateId field to given value.



### HasTemplateId



`func (o *SessionServiceCreateSessionBody) HasTemplateId() bool`



HasTemplateId returns a boolean if a field has been set.





[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
