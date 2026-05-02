# \TemplateServiceAPI



All URIs are relative to *http://localhost*



Method | HTTP request | Description

------------- | ------------- | -------------

[**TemplateServiceCreateTemplate**](TemplateServiceAPI.md#TemplateServiceCreateTemplate) | **Post** /api/v1/nodes/{nodeName}/templates |

[**TemplateServiceDeleteTemplate**](TemplateServiceAPI.md#TemplateServiceDeleteTemplate) | **Delete** /api/v1/nodes/{nodeName}/templates/{id} |

[**TemplateServiceGetTemplate**](TemplateServiceAPI.md#TemplateServiceGetTemplate) | **Get** /api/v1/nodes/{nodeName}/templates/{id} |

[**TemplateServiceGetTemplateConfig**](TemplateServiceAPI.md#TemplateServiceGetTemplateConfig) | **Get** /api/v1/nodes/{nodeName}/templates/{id}/config |

[**TemplateServiceListTemplates**](TemplateServiceAPI.md#TemplateServiceListTemplates) | **Get** /api/v1/nodes/{nodeName}/templates |

[**TemplateServiceUpdateTemplate**](TemplateServiceAPI.md#TemplateServiceUpdateTemplate) | **Put** /api/v1/nodes/{nodeName}/templates/{id} |

[**TemplateServiceUpdateTemplateConfig**](TemplateServiceAPI.md#TemplateServiceUpdateTemplateConfig) | **Put** /api/v1/nodes/{nodeName}/templates/{id}/config |







## TemplateServiceCreateTemplate



> V1SessionTemplate TemplateServiceCreateTemplate(ctx, nodeName).Body(body).Execute()







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

	body := *openapiclient.NewTemplateServiceCreateTemplateBody() // TemplateServiceCreateTemplateBody |



	configuration := openapiclient.NewConfiguration()

	apiClient := openapiclient.NewAPIClient(configuration)

	resp, r, err := apiClient.TemplateServiceAPI.TemplateServiceCreateTemplate(context.Background(), nodeName).Body(body).Execute()

	if err != nil {

		fmt.Fprintf(os.Stderr, "Error when calling `TemplateServiceAPI.TemplateServiceCreateTemplate``: %v\n", err)

		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)

	}

	// response from `TemplateServiceCreateTemplate`: V1SessionTemplate

	fmt.Fprintf(os.Stdout, "Response from `TemplateServiceAPI.TemplateServiceCreateTemplate`: %v\n", resp)

}

```



### Path Parameters





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------

**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.

**nodeName** | **string** |  |



### Other Parameters



Other parameters are passed through a pointer to a apiTemplateServiceCreateTemplateRequest struct via the builder pattern





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------



 **body** | [**TemplateServiceCreateTemplateBody**](TemplateServiceCreateTemplateBody.md) |  |



### Return type



[**V1SessionTemplate**](V1SessionTemplate.md)



### Authorization



No authorization required



### HTTP request headers



- **Content-Type**: application/json

- **Accept**: application/json



[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)

[[Back to Model list]](../README.md#documentation-for-models)

[[Back to README]](../README.md)





## TemplateServiceDeleteTemplate



> map[string]interface{} TemplateServiceDeleteTemplate(ctx, nodeName, id).Execute()







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

	resp, r, err := apiClient.TemplateServiceAPI.TemplateServiceDeleteTemplate(context.Background(), nodeName, id).Execute()

	if err != nil {

		fmt.Fprintf(os.Stderr, "Error when calling `TemplateServiceAPI.TemplateServiceDeleteTemplate``: %v\n", err)

		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)

	}

	// response from `TemplateServiceDeleteTemplate`: map[string]interface{}

	fmt.Fprintf(os.Stdout, "Response from `TemplateServiceAPI.TemplateServiceDeleteTemplate`: %v\n", resp)

}

```



### Path Parameters





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------

**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.

**nodeName** | **string** |  |

**id** | **string** |  |



### Other Parameters



Other parameters are passed through a pointer to a apiTemplateServiceDeleteTemplateRequest struct via the builder pattern





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





## TemplateServiceGetTemplate



> V1SessionTemplate TemplateServiceGetTemplate(ctx, nodeName, id).Execute()







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

	resp, r, err := apiClient.TemplateServiceAPI.TemplateServiceGetTemplate(context.Background(), nodeName, id).Execute()

	if err != nil {

		fmt.Fprintf(os.Stderr, "Error when calling `TemplateServiceAPI.TemplateServiceGetTemplate``: %v\n", err)

		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)

	}

	// response from `TemplateServiceGetTemplate`: V1SessionTemplate

	fmt.Fprintf(os.Stdout, "Response from `TemplateServiceAPI.TemplateServiceGetTemplate`: %v\n", resp)

}

```



### Path Parameters





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------

**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.

**nodeName** | **string** |  |

**id** | **string** |  |



### Other Parameters



Other parameters are passed through a pointer to a apiTemplateServiceGetTemplateRequest struct via the builder pattern





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------







### Return type



[**V1SessionTemplate**](V1SessionTemplate.md)



### Authorization



No authorization required



### HTTP request headers



- **Content-Type**: Not defined

- **Accept**: application/json



[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)

[[Back to Model list]](../README.md#documentation-for-models)

[[Back to README]](../README.md)





## TemplateServiceGetTemplateConfig



> V1TemplateConfigView TemplateServiceGetTemplateConfig(ctx, nodeName, id).Execute()







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

	resp, r, err := apiClient.TemplateServiceAPI.TemplateServiceGetTemplateConfig(context.Background(), nodeName, id).Execute()

	if err != nil {

		fmt.Fprintf(os.Stderr, "Error when calling `TemplateServiceAPI.TemplateServiceGetTemplateConfig``: %v\n", err)

		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)

	}

	// response from `TemplateServiceGetTemplateConfig`: V1TemplateConfigView

	fmt.Fprintf(os.Stdout, "Response from `TemplateServiceAPI.TemplateServiceGetTemplateConfig`: %v\n", resp)

}

```



### Path Parameters





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------

**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.

**nodeName** | **string** |  |

**id** | **string** |  |



### Other Parameters



Other parameters are passed through a pointer to a apiTemplateServiceGetTemplateConfigRequest struct via the builder pattern





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------







### Return type



[**V1TemplateConfigView**](V1TemplateConfigView.md)



### Authorization



No authorization required



### HTTP request headers



- **Content-Type**: Not defined

- **Accept**: application/json



[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)

[[Back to Model list]](../README.md#documentation-for-models)

[[Back to README]](../README.md)





## TemplateServiceListTemplates



> V1ListTemplatesResponse TemplateServiceListTemplates(ctx, nodeName).Execute()







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

	resp, r, err := apiClient.TemplateServiceAPI.TemplateServiceListTemplates(context.Background(), nodeName).Execute()

	if err != nil {

		fmt.Fprintf(os.Stderr, "Error when calling `TemplateServiceAPI.TemplateServiceListTemplates``: %v\n", err)

		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)

	}

	// response from `TemplateServiceListTemplates`: V1ListTemplatesResponse

	fmt.Fprintf(os.Stdout, "Response from `TemplateServiceAPI.TemplateServiceListTemplates`: %v\n", resp)

}

```



### Path Parameters





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------

**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.

**nodeName** | **string** |  |



### Other Parameters



Other parameters are passed through a pointer to a apiTemplateServiceListTemplatesRequest struct via the builder pattern





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------





### Return type



[**V1ListTemplatesResponse**](V1ListTemplatesResponse.md)



### Authorization



No authorization required



### HTTP request headers



- **Content-Type**: Not defined

- **Accept**: application/json



[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)

[[Back to Model list]](../README.md#documentation-for-models)

[[Back to README]](../README.md)





## TemplateServiceUpdateTemplate



> V1SessionTemplate TemplateServiceUpdateTemplate(ctx, nodeName, id).Body(body).Execute()







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

	body := *openapiclient.NewTemplateServiceUpdateTemplateBody() // TemplateServiceUpdateTemplateBody |



	configuration := openapiclient.NewConfiguration()

	apiClient := openapiclient.NewAPIClient(configuration)

	resp, r, err := apiClient.TemplateServiceAPI.TemplateServiceUpdateTemplate(context.Background(), nodeName, id).Body(body).Execute()

	if err != nil {

		fmt.Fprintf(os.Stderr, "Error when calling `TemplateServiceAPI.TemplateServiceUpdateTemplate``: %v\n", err)

		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)

	}

	// response from `TemplateServiceUpdateTemplate`: V1SessionTemplate

	fmt.Fprintf(os.Stdout, "Response from `TemplateServiceAPI.TemplateServiceUpdateTemplate`: %v\n", resp)

}

```



### Path Parameters





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------

**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.

**nodeName** | **string** |  |

**id** | **string** |  |



### Other Parameters



Other parameters are passed through a pointer to a apiTemplateServiceUpdateTemplateRequest struct via the builder pattern





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------





 **body** | [**TemplateServiceUpdateTemplateBody**](TemplateServiceUpdateTemplateBody.md) |  |



### Return type



[**V1SessionTemplate**](V1SessionTemplate.md)



### Authorization



No authorization required



### HTTP request headers



- **Content-Type**: application/json

- **Accept**: application/json



[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)

[[Back to Model list]](../README.md#documentation-for-models)

[[Back to README]](../README.md)





## TemplateServiceUpdateTemplateConfig



> V1TemplateConfigView TemplateServiceUpdateTemplateConfig(ctx, nodeName, id).Body(body).Execute()







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

	body := *openapiclient.NewTemplateServiceUpdateTemplateConfigBody() // TemplateServiceUpdateTemplateConfigBody |



	configuration := openapiclient.NewConfiguration()

	apiClient := openapiclient.NewAPIClient(configuration)

	resp, r, err := apiClient.TemplateServiceAPI.TemplateServiceUpdateTemplateConfig(context.Background(), nodeName, id).Body(body).Execute()

	if err != nil {

		fmt.Fprintf(os.Stderr, "Error when calling `TemplateServiceAPI.TemplateServiceUpdateTemplateConfig``: %v\n", err)

		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)

	}

	// response from `TemplateServiceUpdateTemplateConfig`: V1TemplateConfigView

	fmt.Fprintf(os.Stdout, "Response from `TemplateServiceAPI.TemplateServiceUpdateTemplateConfig`: %v\n", resp)

}

```



### Path Parameters





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------

**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.

**nodeName** | **string** |  |

**id** | **string** |  |



### Other Parameters



Other parameters are passed through a pointer to a apiTemplateServiceUpdateTemplateConfigRequest struct via the builder pattern





Name | Type | Description  | Notes

------------- | ------------- | ------------- | -------------





 **body** | [**TemplateServiceUpdateTemplateConfigBody**](TemplateServiceUpdateTemplateConfigBody.md) |  |



### Return type



[**V1TemplateConfigView**](V1TemplateConfigView.md)



### Authorization



No authorization required



### HTTP request headers



- **Content-Type**: application/json

- **Accept**: application/json



[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)

[[Back to Model list]](../README.md#documentation-for-models)

[[Back to README]](../README.md)
