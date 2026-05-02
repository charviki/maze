# \NodeServiceAPI



All URIs are relative to *http://localhost*



Method | HTTP request | Description

------------- | ------------- | -------------

[**NodeServiceDeleteNode**](NodeServiceAPI.md#NodeServiceDeleteNode) | **Delete** /api/v1/nodes/{name} |

[**NodeServiceGetNode**](NodeServiceAPI.md#NodeServiceGetNode) | **Get** /api/v1/nodes/{name} |

[**NodeServiceListNodes**](NodeServiceAPI.md#NodeServiceListNodes) | **Get** /api/v1/nodes |







## NodeServiceDeleteNode



> map[string]interface{} NodeServiceDeleteNode(ctx, name).Execute()







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

	name := "name_example" // string |



	configuration := openapiclient.NewConfiguration()

	apiClient := openapiclient.NewAPIClient(configuration)

	resp, r, err := apiClient.NodeServiceAPI.NodeServiceDeleteNode(context.Background(), name).Execute()

	if err != nil {

		fmt.Fprintf(os.Stderr, "Error when calling `NodeServiceAPI.NodeServiceDeleteNode``: %v\n", err)

		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)

	}

	// response from `NodeServiceDeleteNode`: map[string]interface{}

	fmt.Fprintf(os.Stdout, "Response from `NodeServiceAPI.NodeServiceDeleteNode`: %v\n", resp)

}

```



### Path Parameters





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------

**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.

**name** | **string** |  |



### Other Parameters



Other parameters are passed through a pointer to a apiNodeServiceDeleteNodeRequest struct via the builder pattern





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------





### Return type



**map[string]interface{}**



### Authorization



No authorization required



### HTTP request headers



- **Content-Type**: Not defined

- **Accept**: application/json



[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)

[[Back to Model list]](../README.md#documentation-for-models)

[[Back to README]](../README.md)





## NodeServiceGetNode



> V1NodeInfo NodeServiceGetNode(ctx, name).Execute()







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

	name := "name_example" // string |



	configuration := openapiclient.NewConfiguration()

	apiClient := openapiclient.NewAPIClient(configuration)

	resp, r, err := apiClient.NodeServiceAPI.NodeServiceGetNode(context.Background(), name).Execute()

	if err != nil {

		fmt.Fprintf(os.Stderr, "Error when calling `NodeServiceAPI.NodeServiceGetNode``: %v\n", err)

		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)

	}

	// response from `NodeServiceGetNode`: V1NodeInfo

	fmt.Fprintf(os.Stdout, "Response from `NodeServiceAPI.NodeServiceGetNode`: %v\n", resp)

}

```



### Path Parameters





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------

**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.

**name** | **string** |  |



### Other Parameters



Other parameters are passed through a pointer to a apiNodeServiceGetNodeRequest struct via the builder pattern





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------





### Return type



[**V1NodeInfo**](V1NodeInfo.md)



### Authorization



No authorization required



### HTTP request headers



- **Content-Type**: Not defined

- **Accept**: application/json



[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)

[[Back to Model list]](../README.md#documentation-for-models)

[[Back to README]](../README.md)





## NodeServiceListNodes



> V1ListNodesResponse NodeServiceListNodes(ctx).Execute()







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



	configuration := openapiclient.NewConfiguration()

	apiClient := openapiclient.NewAPIClient(configuration)

	resp, r, err := apiClient.NodeServiceAPI.NodeServiceListNodes(context.Background()).Execute()

	if err != nil {

		fmt.Fprintf(os.Stderr, "Error when calling `NodeServiceAPI.NodeServiceListNodes``: %v\n", err)

		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)

	}

	// response from `NodeServiceListNodes`: V1ListNodesResponse

	fmt.Fprintf(os.Stdout, "Response from `NodeServiceAPI.NodeServiceListNodes`: %v\n", resp)

}

```



### Path Parameters



This endpoint does not need any parameter.



### Other Parameters



Other parameters are passed through a pointer to a apiNodeServiceListNodesRequest struct via the builder pattern





### Return type



[**V1ListNodesResponse**](V1ListNodesResponse.md)



### Authorization



No authorization required



### HTTP request headers



- **Content-Type**: Not defined

- **Accept**: application/json



[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)

[[Back to Model list]](../README.md#documentation-for-models)

[[Back to README]](../README.md)
