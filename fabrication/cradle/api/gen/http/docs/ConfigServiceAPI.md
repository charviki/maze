# \ConfigServiceAPI



All URIs are relative to *http://localhost*



Method | HTTP request | Description

------------- | ------------- | -------------

[**ConfigServiceGetConfig**](ConfigServiceAPI.md#ConfigServiceGetConfig) | **Get** /api/v1/nodes/{nodeName}/local-config |

[**ConfigServiceUpdateConfig**](ConfigServiceAPI.md#ConfigServiceUpdateConfig) | **Put** /api/v1/nodes/{nodeName}/local-config |







## ConfigServiceGetConfig



> V1LocalAgentConfig ConfigServiceGetConfig(ctx, nodeName).Execute()







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

	nodeName := "nodeName_example" // string |



	configuration := openapiclient.NewConfiguration()

	apiClient := openapiclient.NewAPIClient(configuration)

	resp, r, err := apiClient.ConfigServiceAPI.ConfigServiceGetConfig(context.Background(), nodeName).Execute()

	if err != nil {

		fmt.Fprintf(os.Stderr, "Error when calling `ConfigServiceAPI.ConfigServiceGetConfig``: %v\n", err)

		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)

	}

	// response from `ConfigServiceGetConfig`: V1LocalAgentConfig

	fmt.Fprintf(os.Stdout, "Response from `ConfigServiceAPI.ConfigServiceGetConfig`: %v\n", resp)

}

```



### Path Parameters





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------

**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.

**nodeName** | **string** |  |



### Other Parameters



Other parameters are passed through a pointer to a apiConfigServiceGetConfigRequest struct via the builder pattern





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------





### Return type



[**V1LocalAgentConfig**](V1LocalAgentConfig.md)



### Authorization



No authorization required



### HTTP request headers



- **Content-Type**: Not defined

- **Accept**: application/json



[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)

[[Back to Model list]](../README.md#documentation-for-models)

[[Back to README]](../README.md)





## ConfigServiceUpdateConfig



> V1LocalAgentConfig ConfigServiceUpdateConfig(ctx, nodeName).Body(body).Execute()







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

	nodeName := "nodeName_example" // string |

	body := *openapiclient.NewConfigServiceUpdateConfigBody() // ConfigServiceUpdateConfigBody |



	configuration := openapiclient.NewConfiguration()

	apiClient := openapiclient.NewAPIClient(configuration)

	resp, r, err := apiClient.ConfigServiceAPI.ConfigServiceUpdateConfig(context.Background(), nodeName).Body(body).Execute()

	if err != nil {

		fmt.Fprintf(os.Stderr, "Error when calling `ConfigServiceAPI.ConfigServiceUpdateConfig``: %v\n", err)

		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)

	}

	// response from `ConfigServiceUpdateConfig`: V1LocalAgentConfig

	fmt.Fprintf(os.Stdout, "Response from `ConfigServiceAPI.ConfigServiceUpdateConfig`: %v\n", resp)

}

```



### Path Parameters





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------

**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.

**nodeName** | **string** |  |



### Other Parameters



Other parameters are passed through a pointer to a apiConfigServiceUpdateConfigRequest struct via the builder pattern





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------



 **body** | [**ConfigServiceUpdateConfigBody**](ConfigServiceUpdateConfigBody.md) |  |



### Return type



[**V1LocalAgentConfig**](V1LocalAgentConfig.md)



### Authorization



No authorization required



### HTTP request headers



- **Content-Type**: application/json

- **Accept**: application/json



[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)

[[Back to Model list]](../README.md#documentation-for-models)

[[Back to README]](../README.md)
