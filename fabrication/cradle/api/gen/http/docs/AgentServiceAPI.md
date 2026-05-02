# \AgentServiceAPI



All URIs are relative to *http://localhost*



Method | HTTP request | Description

------------- | ------------- | -------------

[**AgentServiceHeartbeat**](AgentServiceAPI.md#AgentServiceHeartbeat) | **Post** /api/v1/nodes/heartbeat |

[**AgentServiceRegister**](AgentServiceAPI.md#AgentServiceRegister) | **Post** /api/v1/nodes/register |







## AgentServiceHeartbeat



> V1HeartbeatResponse AgentServiceHeartbeat(ctx).Body(body).Execute()







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

	body := *openapiclient.NewV1HeartbeatRequest() // V1HeartbeatRequest |



	configuration := openapiclient.NewConfiguration()

	apiClient := openapiclient.NewAPIClient(configuration)

	resp, r, err := apiClient.AgentServiceAPI.AgentServiceHeartbeat(context.Background()).Body(body).Execute()

	if err != nil {

		fmt.Fprintf(os.Stderr, "Error when calling `AgentServiceAPI.AgentServiceHeartbeat``: %v\n", err)

		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)

	}

	// response from `AgentServiceHeartbeat`: V1HeartbeatResponse

	fmt.Fprintf(os.Stdout, "Response from `AgentServiceAPI.AgentServiceHeartbeat`: %v\n", resp)

}

```



### Path Parameters







### Other Parameters



Other parameters are passed through a pointer to a apiAgentServiceHeartbeatRequest struct via the builder pattern





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------

 **body** | [**V1HeartbeatRequest**](V1HeartbeatRequest.md) |  |



### Return type



[**V1HeartbeatResponse**](V1HeartbeatResponse.md)



### Authorization



No authorization required



### HTTP request headers



- **Content-Type**: application/json

- **Accept**: application/json



[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)

[[Back to Model list]](../README.md#documentation-for-models)

[[Back to README]](../README.md)





## AgentServiceRegister



> V1RegisterResponse AgentServiceRegister(ctx).Body(body).Execute()







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

	body := *openapiclient.NewV1RegisterRequest() // V1RegisterRequest |



	configuration := openapiclient.NewConfiguration()

	apiClient := openapiclient.NewAPIClient(configuration)

	resp, r, err := apiClient.AgentServiceAPI.AgentServiceRegister(context.Background()).Body(body).Execute()

	if err != nil {

		fmt.Fprintf(os.Stderr, "Error when calling `AgentServiceAPI.AgentServiceRegister``: %v\n", err)

		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)

	}

	// response from `AgentServiceRegister`: V1RegisterResponse

	fmt.Fprintf(os.Stdout, "Response from `AgentServiceAPI.AgentServiceRegister`: %v\n", resp)

}

```



### Path Parameters







### Other Parameters



Other parameters are passed through a pointer to a apiAgentServiceRegisterRequest struct via the builder pattern





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------

 **body** | [**V1RegisterRequest**](V1RegisterRequest.md) |  |



### Return type



[**V1RegisterResponse**](V1RegisterResponse.md)



### Authorization



No authorization required



### HTTP request headers



- **Content-Type**: application/json

- **Accept**: application/json



[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)

[[Back to Model list]](../README.md#documentation-for-models)

[[Back to README]](../README.md)
