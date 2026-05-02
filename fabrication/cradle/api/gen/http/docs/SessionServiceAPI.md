# \SessionServiceAPI



All URIs are relative to *http://localhost*



Method | HTTP request | Description

------------- | ------------- | -------------

[**SessionServiceCreateSession**](SessionServiceAPI.md#SessionServiceCreateSession) | **Post** /api/v1/nodes/{nodeName}/sessions |

[**SessionServiceDeleteSession**](SessionServiceAPI.md#SessionServiceDeleteSession) | **Delete** /api/v1/nodes/{nodeName}/sessions/{id} |

[**SessionServiceGetEnv**](SessionServiceAPI.md#SessionServiceGetEnv) | **Get** /api/v1/nodes/{nodeName}/sessions/{id}/env |

[**SessionServiceGetOutput**](SessionServiceAPI.md#SessionServiceGetOutput) | **Get** /api/v1/nodes/{nodeName}/sessions/{id}/output |

[**SessionServiceGetSavedSessions**](SessionServiceAPI.md#SessionServiceGetSavedSessions) | **Get** /api/v1/nodes/{nodeName}/sessions/saved |

[**SessionServiceGetSession**](SessionServiceAPI.md#SessionServiceGetSession) | **Get** /api/v1/nodes/{nodeName}/sessions/{id} |

[**SessionServiceGetSessionConfig**](SessionServiceAPI.md#SessionServiceGetSessionConfig) | **Get** /api/v1/nodes/{nodeName}/sessions/{id}/config |

[**SessionServiceListSessions**](SessionServiceAPI.md#SessionServiceListSessions) | **Get** /api/v1/nodes/{nodeName}/sessions |

[**SessionServiceRestoreSession**](SessionServiceAPI.md#SessionServiceRestoreSession) | **Post** /api/v1/nodes/{nodeName}/sessions/{id}/restore |

[**SessionServiceSaveSessions**](SessionServiceAPI.md#SessionServiceSaveSessions) | **Post** /api/v1/nodes/{nodeName}/sessions/save |

[**SessionServiceSendInput**](SessionServiceAPI.md#SessionServiceSendInput) | **Post** /api/v1/nodes/{nodeName}/sessions/{id}/input |

[**SessionServiceSendSignal**](SessionServiceAPI.md#SessionServiceSendSignal) | **Post** /api/v1/nodes/{nodeName}/sessions/{id}/signal |

[**SessionServiceUpdateSessionConfig**](SessionServiceAPI.md#SessionServiceUpdateSessionConfig) | **Put** /api/v1/nodes/{nodeName}/sessions/{id}/config |







## SessionServiceCreateSession



> V1Session SessionServiceCreateSession(ctx, nodeName).Body(body).Execute()







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

	body := *openapiclient.NewSessionServiceCreateSessionBody() // SessionServiceCreateSessionBody |



	configuration := openapiclient.NewConfiguration()

	apiClient := openapiclient.NewAPIClient(configuration)

	resp, r, err := apiClient.SessionServiceAPI.SessionServiceCreateSession(context.Background(), nodeName).Body(body).Execute()

	if err != nil {

		fmt.Fprintf(os.Stderr, "Error when calling `SessionServiceAPI.SessionServiceCreateSession``: %v\n", err)

		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)

	}

	// response from `SessionServiceCreateSession`: V1Session

	fmt.Fprintf(os.Stdout, "Response from `SessionServiceAPI.SessionServiceCreateSession`: %v\n", resp)

}

```



### Path Parameters





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------

**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.

**nodeName** | **string** |  |



### Other Parameters



Other parameters are passed through a pointer to a apiSessionServiceCreateSessionRequest struct via the builder pattern





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------



 **body** | [**SessionServiceCreateSessionBody**](SessionServiceCreateSessionBody.md) |  |



### Return type



[**V1Session**](V1Session.md)



### Authorization



No authorization required



### HTTP request headers



- **Content-Type**: application/json

- **Accept**: application/json



[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)

[[Back to Model list]](../README.md#documentation-for-models)

[[Back to README]](../README.md)





## SessionServiceDeleteSession



> map[string]interface{} SessionServiceDeleteSession(ctx, nodeName, id).Execute()







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

	id := "id_example" // string |



	configuration := openapiclient.NewConfiguration()

	apiClient := openapiclient.NewAPIClient(configuration)

	resp, r, err := apiClient.SessionServiceAPI.SessionServiceDeleteSession(context.Background(), nodeName, id).Execute()

	if err != nil {

		fmt.Fprintf(os.Stderr, "Error when calling `SessionServiceAPI.SessionServiceDeleteSession``: %v\n", err)

		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)

	}

	// response from `SessionServiceDeleteSession`: map[string]interface{}

	fmt.Fprintf(os.Stdout, "Response from `SessionServiceAPI.SessionServiceDeleteSession`: %v\n", resp)

}

```



### Path Parameters





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------

**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.

**nodeName** | **string** |  |

**id** | **string** |  |



### Other Parameters



Other parameters are passed through a pointer to a apiSessionServiceDeleteSessionRequest struct via the builder pattern





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





## SessionServiceGetEnv



> V1GetEnvResponse SessionServiceGetEnv(ctx, nodeName, id).Execute()







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

	id := "id_example" // string |



	configuration := openapiclient.NewConfiguration()

	apiClient := openapiclient.NewAPIClient(configuration)

	resp, r, err := apiClient.SessionServiceAPI.SessionServiceGetEnv(context.Background(), nodeName, id).Execute()

	if err != nil {

		fmt.Fprintf(os.Stderr, "Error when calling `SessionServiceAPI.SessionServiceGetEnv``: %v\n", err)

		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)

	}

	// response from `SessionServiceGetEnv`: V1GetEnvResponse

	fmt.Fprintf(os.Stdout, "Response from `SessionServiceAPI.SessionServiceGetEnv`: %v\n", resp)

}

```



### Path Parameters





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------

**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.

**nodeName** | **string** |  |

**id** | **string** |  |



### Other Parameters



Other parameters are passed through a pointer to a apiSessionServiceGetEnvRequest struct via the builder pattern





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------







### Return type



[**V1GetEnvResponse**](V1GetEnvResponse.md)



### Authorization



No authorization required



### HTTP request headers



- **Content-Type**: Not defined

- **Accept**: application/json



[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)

[[Back to Model list]](../README.md#documentation-for-models)

[[Back to README]](../README.md)





## SessionServiceGetOutput



> V1TerminalOutput SessionServiceGetOutput(ctx, nodeName, id).Lines(lines).Execute()







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

	id := "id_example" // string |

	lines := int32(56) // int32 |  (optional)



	configuration := openapiclient.NewConfiguration()

	apiClient := openapiclient.NewAPIClient(configuration)

	resp, r, err := apiClient.SessionServiceAPI.SessionServiceGetOutput(context.Background(), nodeName, id).Lines(lines).Execute()

	if err != nil {

		fmt.Fprintf(os.Stderr, "Error when calling `SessionServiceAPI.SessionServiceGetOutput``: %v\n", err)

		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)

	}

	// response from `SessionServiceGetOutput`: V1TerminalOutput

	fmt.Fprintf(os.Stdout, "Response from `SessionServiceAPI.SessionServiceGetOutput`: %v\n", resp)

}

```



### Path Parameters





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------

**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.

**nodeName** | **string** |  |

**id** | **string** |  |



### Other Parameters



Other parameters are passed through a pointer to a apiSessionServiceGetOutputRequest struct via the builder pattern





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------





 **lines** | **int32** |  |



### Return type



[**V1TerminalOutput**](V1TerminalOutput.md)



### Authorization



No authorization required



### HTTP request headers



- **Content-Type**: Not defined

- **Accept**: application/json



[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)

[[Back to Model list]](../README.md#documentation-for-models)

[[Back to README]](../README.md)





## SessionServiceGetSavedSessions



> V1GetSavedSessionsResponse SessionServiceGetSavedSessions(ctx, nodeName).Execute()







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

	resp, r, err := apiClient.SessionServiceAPI.SessionServiceGetSavedSessions(context.Background(), nodeName).Execute()

	if err != nil {

		fmt.Fprintf(os.Stderr, "Error when calling `SessionServiceAPI.SessionServiceGetSavedSessions``: %v\n", err)

		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)

	}

	// response from `SessionServiceGetSavedSessions`: V1GetSavedSessionsResponse

	fmt.Fprintf(os.Stdout, "Response from `SessionServiceAPI.SessionServiceGetSavedSessions`: %v\n", resp)

}

```



### Path Parameters





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------

**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.

**nodeName** | **string** |  |



### Other Parameters



Other parameters are passed through a pointer to a apiSessionServiceGetSavedSessionsRequest struct via the builder pattern





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------





### Return type



[**V1GetSavedSessionsResponse**](V1GetSavedSessionsResponse.md)



### Authorization



No authorization required



### HTTP request headers



- **Content-Type**: Not defined

- **Accept**: application/json



[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)

[[Back to Model list]](../README.md#documentation-for-models)

[[Back to README]](../README.md)





## SessionServiceGetSession



> V1Session SessionServiceGetSession(ctx, nodeName, id).Execute()







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

	id := "id_example" // string |



	configuration := openapiclient.NewConfiguration()

	apiClient := openapiclient.NewAPIClient(configuration)

	resp, r, err := apiClient.SessionServiceAPI.SessionServiceGetSession(context.Background(), nodeName, id).Execute()

	if err != nil {

		fmt.Fprintf(os.Stderr, "Error when calling `SessionServiceAPI.SessionServiceGetSession``: %v\n", err)

		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)

	}

	// response from `SessionServiceGetSession`: V1Session

	fmt.Fprintf(os.Stdout, "Response from `SessionServiceAPI.SessionServiceGetSession`: %v\n", resp)

}

```



### Path Parameters





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------

**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.

**nodeName** | **string** |  |

**id** | **string** |  |



### Other Parameters



Other parameters are passed through a pointer to a apiSessionServiceGetSessionRequest struct via the builder pattern





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------







### Return type



[**V1Session**](V1Session.md)



### Authorization



No authorization required



### HTTP request headers



- **Content-Type**: Not defined

- **Accept**: application/json



[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)

[[Back to Model list]](../README.md#documentation-for-models)

[[Back to README]](../README.md)





## SessionServiceGetSessionConfig



> V1SessionConfigView SessionServiceGetSessionConfig(ctx, nodeName, id).Execute()







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

	id := "id_example" // string |



	configuration := openapiclient.NewConfiguration()

	apiClient := openapiclient.NewAPIClient(configuration)

	resp, r, err := apiClient.SessionServiceAPI.SessionServiceGetSessionConfig(context.Background(), nodeName, id).Execute()

	if err != nil {

		fmt.Fprintf(os.Stderr, "Error when calling `SessionServiceAPI.SessionServiceGetSessionConfig``: %v\n", err)

		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)

	}

	// response from `SessionServiceGetSessionConfig`: V1SessionConfigView

	fmt.Fprintf(os.Stdout, "Response from `SessionServiceAPI.SessionServiceGetSessionConfig`: %v\n", resp)

}

```



### Path Parameters





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------

**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.

**nodeName** | **string** |  |

**id** | **string** |  |



### Other Parameters



Other parameters are passed through a pointer to a apiSessionServiceGetSessionConfigRequest struct via the builder pattern





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------







### Return type



[**V1SessionConfigView**](V1SessionConfigView.md)



### Authorization



No authorization required



### HTTP request headers



- **Content-Type**: Not defined

- **Accept**: application/json



[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)

[[Back to Model list]](../README.md#documentation-for-models)

[[Back to README]](../README.md)





## SessionServiceListSessions



> V1ListSessionsResponse SessionServiceListSessions(ctx, nodeName).Execute()







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

	resp, r, err := apiClient.SessionServiceAPI.SessionServiceListSessions(context.Background(), nodeName).Execute()

	if err != nil {

		fmt.Fprintf(os.Stderr, "Error when calling `SessionServiceAPI.SessionServiceListSessions``: %v\n", err)

		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)

	}

	// response from `SessionServiceListSessions`: V1ListSessionsResponse

	fmt.Fprintf(os.Stdout, "Response from `SessionServiceAPI.SessionServiceListSessions`: %v\n", resp)

}

```



### Path Parameters





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------

**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.

**nodeName** | **string** |  |



### Other Parameters



Other parameters are passed through a pointer to a apiSessionServiceListSessionsRequest struct via the builder pattern





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------





### Return type



[**V1ListSessionsResponse**](V1ListSessionsResponse.md)



### Authorization



No authorization required



### HTTP request headers



- **Content-Type**: Not defined

- **Accept**: application/json



[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)

[[Back to Model list]](../README.md#documentation-for-models)

[[Back to README]](../README.md)





## SessionServiceRestoreSession



> map[string]interface{} SessionServiceRestoreSession(ctx, nodeName, id).Body(body).Execute()







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

	id := "id_example" // string |

	body := map[string]interface{}{ ... } // map[string]interface{} |



	configuration := openapiclient.NewConfiguration()

	apiClient := openapiclient.NewAPIClient(configuration)

	resp, r, err := apiClient.SessionServiceAPI.SessionServiceRestoreSession(context.Background(), nodeName, id).Body(body).Execute()

	if err != nil {

		fmt.Fprintf(os.Stderr, "Error when calling `SessionServiceAPI.SessionServiceRestoreSession``: %v\n", err)

		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)

	}

	// response from `SessionServiceRestoreSession`: map[string]interface{}

	fmt.Fprintf(os.Stdout, "Response from `SessionServiceAPI.SessionServiceRestoreSession`: %v\n", resp)

}

```



### Path Parameters





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------

**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.

**nodeName** | **string** |  |

**id** | **string** |  |



### Other Parameters



Other parameters are passed through a pointer to a apiSessionServiceRestoreSessionRequest struct via the builder pattern





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------





 **body** | **map[string]interface{}** |  |



### Return type



**map[string]interface{}**



### Authorization



No authorization required



### HTTP request headers



- **Content-Type**: application/json

- **Accept**: application/json



[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)

[[Back to Model list]](../README.md#documentation-for-models)

[[Back to README]](../README.md)





## SessionServiceSaveSessions



> V1SaveSessionsResponse SessionServiceSaveSessions(ctx, nodeName).Execute()







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

	resp, r, err := apiClient.SessionServiceAPI.SessionServiceSaveSessions(context.Background(), nodeName).Execute()

	if err != nil {

		fmt.Fprintf(os.Stderr, "Error when calling `SessionServiceAPI.SessionServiceSaveSessions``: %v\n", err)

		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)

	}

	// response from `SessionServiceSaveSessions`: V1SaveSessionsResponse

	fmt.Fprintf(os.Stdout, "Response from `SessionServiceAPI.SessionServiceSaveSessions`: %v\n", resp)

}

```



### Path Parameters





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------

**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.

**nodeName** | **string** |  |



### Other Parameters



Other parameters are passed through a pointer to a apiSessionServiceSaveSessionsRequest struct via the builder pattern





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------





### Return type



[**V1SaveSessionsResponse**](V1SaveSessionsResponse.md)



### Authorization



No authorization required



### HTTP request headers



- **Content-Type**: Not defined

- **Accept**: application/json



[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)

[[Back to Model list]](../README.md#documentation-for-models)

[[Back to README]](../README.md)





## SessionServiceSendInput



> map[string]interface{} SessionServiceSendInput(ctx, nodeName, id).Body(body).Execute()







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

	id := "id_example" // string |

	body := *openapiclient.NewSessionServiceSendInputBody() // SessionServiceSendInputBody |



	configuration := openapiclient.NewConfiguration()

	apiClient := openapiclient.NewAPIClient(configuration)

	resp, r, err := apiClient.SessionServiceAPI.SessionServiceSendInput(context.Background(), nodeName, id).Body(body).Execute()

	if err != nil {

		fmt.Fprintf(os.Stderr, "Error when calling `SessionServiceAPI.SessionServiceSendInput``: %v\n", err)

		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)

	}

	// response from `SessionServiceSendInput`: map[string]interface{}

	fmt.Fprintf(os.Stdout, "Response from `SessionServiceAPI.SessionServiceSendInput`: %v\n", resp)

}

```



### Path Parameters





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------

**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.

**nodeName** | **string** |  |

**id** | **string** |  |



### Other Parameters



Other parameters are passed through a pointer to a apiSessionServiceSendInputRequest struct via the builder pattern





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------





 **body** | [**SessionServiceSendInputBody**](SessionServiceSendInputBody.md) |  |



### Return type



**map[string]interface{}**



### Authorization



No authorization required



### HTTP request headers



- **Content-Type**: application/json

- **Accept**: application/json



[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)

[[Back to Model list]](../README.md#documentation-for-models)

[[Back to README]](../README.md)





## SessionServiceSendSignal



> map[string]interface{} SessionServiceSendSignal(ctx, nodeName, id).Body(body).Execute()







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

	id := "id_example" // string |

	body := *openapiclient.NewSessionServiceSendSignalBody() // SessionServiceSendSignalBody |



	configuration := openapiclient.NewConfiguration()

	apiClient := openapiclient.NewAPIClient(configuration)

	resp, r, err := apiClient.SessionServiceAPI.SessionServiceSendSignal(context.Background(), nodeName, id).Body(body).Execute()

	if err != nil {

		fmt.Fprintf(os.Stderr, "Error when calling `SessionServiceAPI.SessionServiceSendSignal``: %v\n", err)

		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)

	}

	// response from `SessionServiceSendSignal`: map[string]interface{}

	fmt.Fprintf(os.Stdout, "Response from `SessionServiceAPI.SessionServiceSendSignal`: %v\n", resp)

}

```



### Path Parameters





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------

**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.

**nodeName** | **string** |  |

**id** | **string** |  |



### Other Parameters



Other parameters are passed through a pointer to a apiSessionServiceSendSignalRequest struct via the builder pattern





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------





 **body** | [**SessionServiceSendSignalBody**](SessionServiceSendSignalBody.md) |  |



### Return type



**map[string]interface{}**



### Authorization



No authorization required



### HTTP request headers



- **Content-Type**: application/json

- **Accept**: application/json



[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)

[[Back to Model list]](../README.md#documentation-for-models)

[[Back to README]](../README.md)





## SessionServiceUpdateSessionConfig



> V1SessionConfigView SessionServiceUpdateSessionConfig(ctx, nodeName, id).Body(body).Execute()







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

	id := "id_example" // string |

	body := *openapiclient.NewSessionServiceUpdateSessionConfigBody() // SessionServiceUpdateSessionConfigBody |



	configuration := openapiclient.NewConfiguration()

	apiClient := openapiclient.NewAPIClient(configuration)

	resp, r, err := apiClient.SessionServiceAPI.SessionServiceUpdateSessionConfig(context.Background(), nodeName, id).Body(body).Execute()

	if err != nil {

		fmt.Fprintf(os.Stderr, "Error when calling `SessionServiceAPI.SessionServiceUpdateSessionConfig``: %v\n", err)

		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)

	}

	// response from `SessionServiceUpdateSessionConfig`: V1SessionConfigView

	fmt.Fprintf(os.Stdout, "Response from `SessionServiceAPI.SessionServiceUpdateSessionConfig`: %v\n", resp)

}

```



### Path Parameters





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------

**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.

**nodeName** | **string** |  |

**id** | **string** |  |



### Other Parameters



Other parameters are passed through a pointer to a apiSessionServiceUpdateSessionConfigRequest struct via the builder pattern





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------





 **body** | [**SessionServiceUpdateSessionConfigBody**](SessionServiceUpdateSessionConfigBody.md) |  |



### Return type



[**V1SessionConfigView**](V1SessionConfigView.md)



### Authorization



No authorization required



### HTTP request headers



- **Content-Type**: application/json

- **Accept**: application/json



[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)

[[Back to Model list]](../README.md#documentation-for-models)

[[Back to README]](../README.md)
