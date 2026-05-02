# \AuditServiceAPI



All URIs are relative to *http://localhost*



Method | HTTP request | Description

------------- | ------------- | -------------

[**AuditServiceGetAuditLogs**](AuditServiceAPI.md#AuditServiceGetAuditLogs) | **Get** /api/v1/audit/logs |







## AuditServiceGetAuditLogs



> V1GetAuditLogsResponse AuditServiceGetAuditLogs(ctx).Page(page).PageSize(pageSize).Execute()







### Example



```go

package main



import (

	"context"

	"fmt"

	"os"

	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"

)



func main() {

	page := int32(56) // int32 |  (optional)

	pageSize := int32(56) // int32 |  (optional)



	configuration := openapiclient.NewConfiguration()

	apiClient := openapiclient.NewAPIClient(configuration)

	resp, r, err := apiClient.AuditServiceAPI.AuditServiceGetAuditLogs(context.Background()).Page(page).PageSize(pageSize).Execute()

	if err != nil {

		fmt.Fprintf(os.Stderr, "Error when calling `AuditServiceAPI.AuditServiceGetAuditLogs``: %v\n", err)

		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)

	}

	// response from `AuditServiceGetAuditLogs`: V1GetAuditLogsResponse

	fmt.Fprintf(os.Stdout, "Response from `AuditServiceAPI.AuditServiceGetAuditLogs`: %v\n", resp)

}

```



### Path Parameters







### Other Parameters



Other parameters are passed through a pointer to a apiAuditServiceGetAuditLogsRequest struct via the builder pattern





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------

 **page** | **int32** |  |

 **pageSize** | **int32** |  |



### Return type



[**V1GetAuditLogsResponse**](V1GetAuditLogsResponse.md)



### Authorization



No authorization required



### HTTP request headers



- **Content-Type**: Not defined

- **Accept**: application/json



[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)

[[Back to Model list]](../README.md#documentation-for-models)

[[Back to README]](../README.md)
