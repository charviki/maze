# V1AgentStatus



## Properties



Name | Type | Description | Notes

------------ | ------------- | ------------- | -------------

**ActiveSessions** | Pointer to **int32** |  | [optional]

**CpuUsage** | Pointer to **float64** |  | [optional]

**MemoryUsageMb** | Pointer to **float64** |  | [optional]

**WorkspaceRoot** | Pointer to **string** |  | [optional]

**SessionDetails** | Pointer to [**[]V1SessionDetail**](V1SessionDetail.md) |  | [optional]

**LocalConfig** | Pointer to [**V1LocalAgentConfig**](V1LocalAgentConfig.md) |  | [optional]



## Methods



### NewV1AgentStatus



`func NewV1AgentStatus() *V1AgentStatus`



NewV1AgentStatus instantiates a new V1AgentStatus object

This constructor will assign default values to properties that have it defined,

and makes sure properties required by API are set, but the set of arguments

will change when the set of required properties is changed



### NewV1AgentStatusWithDefaults



`func NewV1AgentStatusWithDefaults() *V1AgentStatus`



NewV1AgentStatusWithDefaults instantiates a new V1AgentStatus object

This constructor will only assign default values to properties that have it defined,

but it doesn't guarantee that properties required by API are set



### GetActiveSessions



`func (o *V1AgentStatus) GetActiveSessions() int32`



GetActiveSessions returns the ActiveSessions field if non-nil, zero value otherwise.



### GetActiveSessionsOk



`func (o *V1AgentStatus) GetActiveSessionsOk() (*int32, bool)`



GetActiveSessionsOk returns a tuple with the ActiveSessions field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetActiveSessions



`func (o *V1AgentStatus) SetActiveSessions(v int32)`



SetActiveSessions sets ActiveSessions field to given value.



### HasActiveSessions



`func (o *V1AgentStatus) HasActiveSessions() bool`



HasActiveSessions returns a boolean if a field has been set.



### GetCpuUsage



`func (o *V1AgentStatus) GetCpuUsage() float64`



GetCpuUsage returns the CpuUsage field if non-nil, zero value otherwise.



### GetCpuUsageOk



`func (o *V1AgentStatus) GetCpuUsageOk() (*float64, bool)`



GetCpuUsageOk returns a tuple with the CpuUsage field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetCpuUsage



`func (o *V1AgentStatus) SetCpuUsage(v float64)`



SetCpuUsage sets CpuUsage field to given value.



### HasCpuUsage



`func (o *V1AgentStatus) HasCpuUsage() bool`



HasCpuUsage returns a boolean if a field has been set.



### GetMemoryUsageMb



`func (o *V1AgentStatus) GetMemoryUsageMb() float64`



GetMemoryUsageMb returns the MemoryUsageMb field if non-nil, zero value otherwise.



### GetMemoryUsageMbOk



`func (o *V1AgentStatus) GetMemoryUsageMbOk() (*float64, bool)`



GetMemoryUsageMbOk returns a tuple with the MemoryUsageMb field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetMemoryUsageMb



`func (o *V1AgentStatus) SetMemoryUsageMb(v float64)`



SetMemoryUsageMb sets MemoryUsageMb field to given value.



### HasMemoryUsageMb



`func (o *V1AgentStatus) HasMemoryUsageMb() bool`



HasMemoryUsageMb returns a boolean if a field has been set.



### GetWorkspaceRoot



`func (o *V1AgentStatus) GetWorkspaceRoot() string`



GetWorkspaceRoot returns the WorkspaceRoot field if non-nil, zero value otherwise.



### GetWorkspaceRootOk



`func (o *V1AgentStatus) GetWorkspaceRootOk() (*string, bool)`



GetWorkspaceRootOk returns a tuple with the WorkspaceRoot field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetWorkspaceRoot



`func (o *V1AgentStatus) SetWorkspaceRoot(v string)`



SetWorkspaceRoot sets WorkspaceRoot field to given value.



### HasWorkspaceRoot



`func (o *V1AgentStatus) HasWorkspaceRoot() bool`



HasWorkspaceRoot returns a boolean if a field has been set.



### GetSessionDetails



`func (o *V1AgentStatus) GetSessionDetails() []V1SessionDetail`



GetSessionDetails returns the SessionDetails field if non-nil, zero value otherwise.



### GetSessionDetailsOk



`func (o *V1AgentStatus) GetSessionDetailsOk() (*[]V1SessionDetail, bool)`



GetSessionDetailsOk returns a tuple with the SessionDetails field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetSessionDetails



`func (o *V1AgentStatus) SetSessionDetails(v []V1SessionDetail)`



SetSessionDetails sets SessionDetails field to given value.



### HasSessionDetails



`func (o *V1AgentStatus) HasSessionDetails() bool`



HasSessionDetails returns a boolean if a field has been set.



### GetLocalConfig



`func (o *V1AgentStatus) GetLocalConfig() V1LocalAgentConfig`



GetLocalConfig returns the LocalConfig field if non-nil, zero value otherwise.



### GetLocalConfigOk



`func (o *V1AgentStatus) GetLocalConfigOk() (*V1LocalAgentConfig, bool)`



GetLocalConfigOk returns a tuple with the LocalConfig field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetLocalConfig



`func (o *V1AgentStatus) SetLocalConfig(v V1LocalAgentConfig)`



SetLocalConfig sets LocalConfig field to given value.



### HasLocalConfig



`func (o *V1AgentStatus) HasLocalConfig() bool`



HasLocalConfig returns a boolean if a field has been set.





[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
