# \HostServiceAPI



All URIs are relative to *http://localhost*



Method | HTTP request | Description

------------- | ------------- | -------------

[**HostServiceCreateHost**](HostServiceAPI.md#HostServiceCreateHost) | **Post** /api/v1/hosts |

[**HostServiceDeleteHost**](HostServiceAPI.md#HostServiceDeleteHost) | **Delete** /api/v1/hosts/{name} |

[**HostServiceGetBuildLog**](HostServiceAPI.md#HostServiceGetBuildLog) | **Get** /api/v1/hosts/{name}/logs/build |

[**HostServiceGetHost**](HostServiceAPI.md#HostServiceGetHost) | **Get** /api/v1/hosts/{name} |

[**HostServiceGetRuntimeLog**](HostServiceAPI.md#HostServiceGetRuntimeLog) | **Get** /api/v1/hosts/{name}/logs/runtime |

[**HostServiceListHosts**](HostServiceAPI.md#HostServiceListHosts) | **Get** /api/v1/hosts |

[**HostServiceListTools**](HostServiceAPI.md#HostServiceListTools) | **Get** /api/v1/host/tools |







## HostServiceCreateHost



> V1HostSpec HostServiceCreateHost(ctx).Body(body).Execute()







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

	body := *openapiclient.NewV1CreateHostRequest() // V1CreateHostRequest |



	configuration := openapiclient.NewConfiguration()

	apiClient := openapiclient.NewAPIClient(configuration)

	resp, r, err := apiClient.HostServiceAPI.HostServiceCreateHost(context.Background()).Body(body).Execute()

	if err != nil {

		fmt.Fprintf(os.Stderr, "Error when calling `HostServiceAPI.HostServiceCreateHost``: %v\n", err)

		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)

	}

	// response from `HostServiceCreateHost`: V1HostSpec

	fmt.Fprintf(os.Stdout, "Response from `HostServiceAPI.HostServiceCreateHost`: %v\n", resp)

}

```



### Path Parameters







### Other Parameters



Other parameters are passed through a pointer to a apiHostServiceCreateHostRequest struct via the builder pattern





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------

 **body** | [**V1CreateHostRequest**](V1CreateHostRequest.md) |  |



### Return type



[**V1HostSpec**](V1HostSpec.md)



### Authorization



No authorization required



### HTTP request headers



- **Content-Type**: application/json

- **Accept**: application/json



[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)

[[Back to Model list]](../README.md#documentation-for-models)

[[Back to README]](../README.md)





## HostServiceDeleteHost



> map[string]interface{} HostServiceDeleteHost(ctx, name).Execute()







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

	resp, r, err := apiClient.HostServiceAPI.HostServiceDeleteHost(context.Background(), name).Execute()

	if err != nil {

		fmt.Fprintf(os.Stderr, "Error when calling `HostServiceAPI.HostServiceDeleteHost``: %v\n", err)

		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)

	}

	// response from `HostServiceDeleteHost`: map[string]interface{}

	fmt.Fprintf(os.Stdout, "Response from `HostServiceAPI.HostServiceDeleteHost`: %v\n", resp)

}

```



### Path Parameters





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------

**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.

**name** | **string** |  |



### Other Parameters



Other parameters are passed through a pointer to a apiHostServiceDeleteHostRequest struct via the builder pattern





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





## HostServiceGetBuildLog



> V1GetBuildLogResponse HostServiceGetBuildLog(ctx, name).Execute()







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

	resp, r, err := apiClient.HostServiceAPI.HostServiceGetBuildLog(context.Background(), name).Execute()

	if err != nil {

		fmt.Fprintf(os.Stderr, "Error when calling `HostServiceAPI.HostServiceGetBuildLog``: %v\n", err)

		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)

	}

	// response from `HostServiceGetBuildLog`: V1GetBuildLogResponse

	fmt.Fprintf(os.Stdout, "Response from `HostServiceAPI.HostServiceGetBuildLog`: %v\n", resp)

}

```



### Path Parameters





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------

**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.

**name** | **string** |  |



### Other Parameters



Other parameters are passed through a pointer to a apiHostServiceGetBuildLogRequest struct via the builder pattern





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------





### Return type



[**V1GetBuildLogResponse**](V1GetBuildLogResponse.md)



### Authorization



No authorization required



### HTTP request headers



- **Content-Type**: Not defined

- **Accept**: application/json



[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)

[[Back to Model list]](../README.md#documentation-for-models)

[[Back to README]](../README.md)





## HostServiceGetHost



> V1HostInfo HostServiceGetHost(ctx, name).Execute()







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

	resp, r, err := apiClient.HostServiceAPI.HostServiceGetHost(context.Background(), name).Execute()

	if err != nil {

		fmt.Fprintf(os.Stderr, "Error when calling `HostServiceAPI.HostServiceGetHost``: %v\n", err)

		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)

	}

	// response from `HostServiceGetHost`: V1HostInfo

	fmt.Fprintf(os.Stdout, "Response from `HostServiceAPI.HostServiceGetHost`: %v\n", resp)

}

```



### Path Parameters





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------

**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.

**name** | **string** |  |



### Other Parameters



Other parameters are passed through a pointer to a apiHostServiceGetHostRequest struct via the builder pattern





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------





### Return type



[**V1HostInfo**](V1HostInfo.md)



### Authorization



No authorization required



### HTTP request headers



- **Content-Type**: Not defined

- **Accept**: application/json



[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)

[[Back to Model list]](../README.md#documentation-for-models)

[[Back to README]](../README.md)





## HostServiceGetRuntimeLog



> V1GetRuntimeLogResponse HostServiceGetRuntimeLog(ctx, name).Execute()







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

	resp, r, err := apiClient.HostServiceAPI.HostServiceGetRuntimeLog(context.Background(), name).Execute()

	if err != nil {

		fmt.Fprintf(os.Stderr, "Error when calling `HostServiceAPI.HostServiceGetRuntimeLog``: %v\n", err)

		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)

	}

	// response from `HostServiceGetRuntimeLog`: V1GetRuntimeLogResponse

	fmt.Fprintf(os.Stdout, "Response from `HostServiceAPI.HostServiceGetRuntimeLog`: %v\n", resp)

}

```



### Path Parameters





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------

**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.

**name** | **string** |  |



### Other Parameters



Other parameters are passed through a pointer to a apiHostServiceGetRuntimeLogRequest struct via the builder pattern





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------





### Return type



[**V1GetRuntimeLogResponse**](V1GetRuntimeLogResponse.md)



### Authorization



No authorization required



### HTTP request headers



- **Content-Type**: Not defined

- **Accept**: application/json



[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)

[[Back to Model list]](../README.md#documentation-for-models)

[[Back to README]](../README.md)





## HostServiceListHosts



> V1ListHostsResponse HostServiceListHosts(ctx).Execute()







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

	resp, r, err := apiClient.HostServiceAPI.HostServiceListHosts(context.Background()).Execute()

	if err != nil {

		fmt.Fprintf(os.Stderr, "Error when calling `HostServiceAPI.HostServiceListHosts``: %v\n", err)

		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)

	}

	// response from `HostServiceListHosts`: V1ListHostsResponse

	fmt.Fprintf(os.Stdout, "Response from `HostServiceAPI.HostServiceListHosts`: %v\n", resp)

}

```



### Path Parameters



This endpoint does not need any parameter.



### Other Parameters



Other parameters are passed through a pointer to a apiHostServiceListHostsRequest struct via the builder pattern





### Return type



[**V1ListHostsResponse**](V1ListHostsResponse.md)



### Authorization



No authorization required



### HTTP request headers



- **Content-Type**: Not defined

- **Accept**: application/json



[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)

[[Back to Model list]](../README.md#documentation-for-models)

[[Back to README]](../README.md)





## HostServiceListTools



> V1ListToolsResponse HostServiceListTools(ctx).Execute()







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

	resp, r, err := apiClient.HostServiceAPI.HostServiceListTools(context.Background()).Execute()

	if err != nil {

		fmt.Fprintf(os.Stderr, "Error when calling `HostServiceAPI.HostServiceListTools``: %v\n", err)

		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)

	}

	// response from `HostServiceListTools`: V1ListToolsResponse

	fmt.Fprintf(os.Stdout, "Response from `HostServiceAPI.HostServiceListTools`: %v\n", resp)

}

```



### Path Parameters



This endpoint does not need any parameter.



### Other Parameters



Other parameters are passed through a pointer to a apiHostServiceListToolsRequest struct via the builder pattern





### Return type



[**V1ListToolsResponse**](V1ListToolsResponse.md)



### Authorization



No authorization required



### HTTP request headers



- **Content-Type**: Not defined

- **Accept**: application/json



[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)

[[Back to Model list]](../README.md#documentation-for-models)

[[Back to README]](../README.md)
