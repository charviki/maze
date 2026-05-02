# V1SessionState



## Properties



Name | Type | Description | Notes

------------ | ------------- | ------------- | -------------

**SessionName** | Pointer to **string** |  | [optional]

**Pipeline** | Pointer to **string** |  | [optional]

**RestoreStrategy** | Pointer to **string** |  | [optional]

**RestoreCommand** | Pointer to **string** |  | [optional]

**WorkingDir** | Pointer to **string** |  | [optional]

**TemplateId** | Pointer to **string** |  | [optional]

**CliSessionId** | Pointer to **string** |  | [optional]

**EnvSnapshot** | Pointer to **map[string]string** |  | [optional]

**TerminalSnapshot** | Pointer to **string** |  | [optional]

**SavedAt** | Pointer to **string** |  | [optional]



## Methods



### NewV1SessionState



`func NewV1SessionState() *V1SessionState`



NewV1SessionState instantiates a new V1SessionState object

This constructor will assign default values to properties that have it defined,

and makes sure properties required by API are set, but the set of arguments

will change when the set of required properties is changed



### NewV1SessionStateWithDefaults



`func NewV1SessionStateWithDefaults() *V1SessionState`



NewV1SessionStateWithDefaults instantiates a new V1SessionState object

This constructor will only assign default values to properties that have it defined,

but it doesn't guarantee that properties required by API are set



### GetSessionName



`func (o *V1SessionState) GetSessionName() string`



GetSessionName returns the SessionName field if non-nil, zero value otherwise.



### GetSessionNameOk



`func (o *V1SessionState) GetSessionNameOk() (*string, bool)`



GetSessionNameOk returns a tuple with the SessionName field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetSessionName



`func (o *V1SessionState) SetSessionName(v string)`



SetSessionName sets SessionName field to given value.



### HasSessionName



`func (o *V1SessionState) HasSessionName() bool`



HasSessionName returns a boolean if a field has been set.



### GetPipeline



`func (o *V1SessionState) GetPipeline() string`



GetPipeline returns the Pipeline field if non-nil, zero value otherwise.



### GetPipelineOk



`func (o *V1SessionState) GetPipelineOk() (*string, bool)`



GetPipelineOk returns a tuple with the Pipeline field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetPipeline



`func (o *V1SessionState) SetPipeline(v string)`



SetPipeline sets Pipeline field to given value.



### HasPipeline



`func (o *V1SessionState) HasPipeline() bool`



HasPipeline returns a boolean if a field has been set.



### GetRestoreStrategy



`func (o *V1SessionState) GetRestoreStrategy() string`



GetRestoreStrategy returns the RestoreStrategy field if non-nil, zero value otherwise.



### GetRestoreStrategyOk



`func (o *V1SessionState) GetRestoreStrategyOk() (*string, bool)`



GetRestoreStrategyOk returns a tuple with the RestoreStrategy field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetRestoreStrategy



`func (o *V1SessionState) SetRestoreStrategy(v string)`



SetRestoreStrategy sets RestoreStrategy field to given value.



### HasRestoreStrategy



`func (o *V1SessionState) HasRestoreStrategy() bool`



HasRestoreStrategy returns a boolean if a field has been set.



### GetRestoreCommand



`func (o *V1SessionState) GetRestoreCommand() string`



GetRestoreCommand returns the RestoreCommand field if non-nil, zero value otherwise.



### GetRestoreCommandOk



`func (o *V1SessionState) GetRestoreCommandOk() (*string, bool)`



GetRestoreCommandOk returns a tuple with the RestoreCommand field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetRestoreCommand



`func (o *V1SessionState) SetRestoreCommand(v string)`



SetRestoreCommand sets RestoreCommand field to given value.



### HasRestoreCommand



`func (o *V1SessionState) HasRestoreCommand() bool`



HasRestoreCommand returns a boolean if a field has been set.



### GetWorkingDir



`func (o *V1SessionState) GetWorkingDir() string`



GetWorkingDir returns the WorkingDir field if non-nil, zero value otherwise.



### GetWorkingDirOk



`func (o *V1SessionState) GetWorkingDirOk() (*string, bool)`



GetWorkingDirOk returns a tuple with the WorkingDir field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetWorkingDir



`func (o *V1SessionState) SetWorkingDir(v string)`



SetWorkingDir sets WorkingDir field to given value.



### HasWorkingDir



`func (o *V1SessionState) HasWorkingDir() bool`



HasWorkingDir returns a boolean if a field has been set.



### GetTemplateId



`func (o *V1SessionState) GetTemplateId() string`



GetTemplateId returns the TemplateId field if non-nil, zero value otherwise.



### GetTemplateIdOk



`func (o *V1SessionState) GetTemplateIdOk() (*string, bool)`



GetTemplateIdOk returns a tuple with the TemplateId field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetTemplateId



`func (o *V1SessionState) SetTemplateId(v string)`



SetTemplateId sets TemplateId field to given value.



### HasTemplateId



`func (o *V1SessionState) HasTemplateId() bool`



HasTemplateId returns a boolean if a field has been set.



### GetCliSessionId



`func (o *V1SessionState) GetCliSessionId() string`



GetCliSessionId returns the CliSessionId field if non-nil, zero value otherwise.



### GetCliSessionIdOk



`func (o *V1SessionState) GetCliSessionIdOk() (*string, bool)`



GetCliSessionIdOk returns a tuple with the CliSessionId field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetCliSessionId



`func (o *V1SessionState) SetCliSessionId(v string)`



SetCliSessionId sets CliSessionId field to given value.



### HasCliSessionId



`func (o *V1SessionState) HasCliSessionId() bool`



HasCliSessionId returns a boolean if a field has been set.



### GetEnvSnapshot



`func (o *V1SessionState) GetEnvSnapshot() map[string]string`



GetEnvSnapshot returns the EnvSnapshot field if non-nil, zero value otherwise.



### GetEnvSnapshotOk



`func (o *V1SessionState) GetEnvSnapshotOk() (*map[string]string, bool)`



GetEnvSnapshotOk returns a tuple with the EnvSnapshot field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetEnvSnapshot



`func (o *V1SessionState) SetEnvSnapshot(v map[string]string)`



SetEnvSnapshot sets EnvSnapshot field to given value.



### HasEnvSnapshot



`func (o *V1SessionState) HasEnvSnapshot() bool`



HasEnvSnapshot returns a boolean if a field has been set.



### GetTerminalSnapshot



`func (o *V1SessionState) GetTerminalSnapshot() string`



GetTerminalSnapshot returns the TerminalSnapshot field if non-nil, zero value otherwise.



### GetTerminalSnapshotOk



`func (o *V1SessionState) GetTerminalSnapshotOk() (*string, bool)`



GetTerminalSnapshotOk returns a tuple with the TerminalSnapshot field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetTerminalSnapshot



`func (o *V1SessionState) SetTerminalSnapshot(v string)`



SetTerminalSnapshot sets TerminalSnapshot field to given value.



### HasTerminalSnapshot



`func (o *V1SessionState) HasTerminalSnapshot() bool`



HasTerminalSnapshot returns a boolean if a field has been set.



### GetSavedAt



`func (o *V1SessionState) GetSavedAt() string`



GetSavedAt returns the SavedAt field if non-nil, zero value otherwise.



### GetSavedAtOk



`func (o *V1SessionState) GetSavedAtOk() (*string, bool)`



GetSavedAtOk returns a tuple with the SavedAt field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetSavedAt



`func (o *V1SessionState) SetSavedAt(v string)`



SetSavedAt sets SavedAt field to given value.



### HasSavedAt



`func (o *V1SessionState) HasSavedAt() bool`



HasSavedAt returns a boolean if a field has been set.





[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
