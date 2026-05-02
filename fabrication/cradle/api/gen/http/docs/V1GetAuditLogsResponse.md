# V1GetAuditLogsResponse



## Properties



Name | Type | Description | Notes

------------ | ------------- | ------------- | -------------

**Logs** | Pointer to [**[]V1AuditLogEntry**](V1AuditLogEntry.md) |  | [optional]

**Total** | Pointer to **int32** |  | [optional]

**Page** | Pointer to **int32** |  | [optional]

**PageSize** | Pointer to **int32** |  | [optional]



## Methods



### NewV1GetAuditLogsResponse



`func NewV1GetAuditLogsResponse() *V1GetAuditLogsResponse`



NewV1GetAuditLogsResponse instantiates a new V1GetAuditLogsResponse object

This constructor will assign default values to properties that have it defined,

and makes sure properties required by API are set, but the set of arguments

will change when the set of required properties is changed



### NewV1GetAuditLogsResponseWithDefaults



`func NewV1GetAuditLogsResponseWithDefaults() *V1GetAuditLogsResponse`



NewV1GetAuditLogsResponseWithDefaults instantiates a new V1GetAuditLogsResponse object

This constructor will only assign default values to properties that have it defined,

but it doesn't guarantee that properties required by API are set



### GetLogs



`func (o *V1GetAuditLogsResponse) GetLogs() []V1AuditLogEntry`



GetLogs returns the Logs field if non-nil, zero value otherwise.



### GetLogsOk



`func (o *V1GetAuditLogsResponse) GetLogsOk() (*[]V1AuditLogEntry, bool)`



GetLogsOk returns a tuple with the Logs field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetLogs



`func (o *V1GetAuditLogsResponse) SetLogs(v []V1AuditLogEntry)`



SetLogs sets Logs field to given value.



### HasLogs



`func (o *V1GetAuditLogsResponse) HasLogs() bool`



HasLogs returns a boolean if a field has been set.



### GetTotal



`func (o *V1GetAuditLogsResponse) GetTotal() int32`



GetTotal returns the Total field if non-nil, zero value otherwise.



### GetTotalOk



`func (o *V1GetAuditLogsResponse) GetTotalOk() (*int32, bool)`



GetTotalOk returns a tuple with the Total field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetTotal



`func (o *V1GetAuditLogsResponse) SetTotal(v int32)`



SetTotal sets Total field to given value.



### HasTotal



`func (o *V1GetAuditLogsResponse) HasTotal() bool`



HasTotal returns a boolean if a field has been set.



### GetPage



`func (o *V1GetAuditLogsResponse) GetPage() int32`



GetPage returns the Page field if non-nil, zero value otherwise.



### GetPageOk



`func (o *V1GetAuditLogsResponse) GetPageOk() (*int32, bool)`



GetPageOk returns a tuple with the Page field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetPage



`func (o *V1GetAuditLogsResponse) SetPage(v int32)`



SetPage sets Page field to given value.



### HasPage



`func (o *V1GetAuditLogsResponse) HasPage() bool`



HasPage returns a boolean if a field has been set.



### GetPageSize



`func (o *V1GetAuditLogsResponse) GetPageSize() int32`



GetPageSize returns the PageSize field if non-nil, zero value otherwise.



### GetPageSizeOk



`func (o *V1GetAuditLogsResponse) GetPageSizeOk() (*int32, bool)`



GetPageSizeOk returns a tuple with the PageSize field if it's non-nil, zero value otherwise

and a boolean to check if the value has been set.



### SetPageSize



`func (o *V1GetAuditLogsResponse) SetPageSize(v int32)`



SetPageSize sets PageSize field to given value.



### HasPageSize



`func (o *V1GetAuditLogsResponse) HasPageSize() bool`



HasPageSize returns a boolean if a field has been set.





[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
