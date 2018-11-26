# Azure Functions for Go

[![Build Status](https://travis-ci.com/Azure/azure-functions-go.svg?token=pzfiiBDjqjzLCtQCMpq1&branch=dev)](https://travis-ci.com/Azure/azure-functions-go)

This project adds Go support to Azure Functions by implementing a [language worker][] for Go. It requires the following:

- Go 1.10+
- Linux or OS X
- Docker
- [Azure CLI](https://github.com/Azure/azure-cli)

[language worker]: https://github.com/Azure/azure-functions-host/wiki/Language-Extensibility

## Contents:

- [Run a Go Functions instance](#run-a-go-functions-instance)
- [Write and deploy a Go Function](#write-and-deploy-a-go-function)

# Run a Go Functions instance

During preview you must build and deploy your own Functions instance with the
Go worker included. To do so clone this repo and run one of the following to
deploy an instance with bundled samples from the `sample` directory:

- To deploy to a local Docker daemon: `make local-instance`
- To deploy to Azure App Service: `make azure-instance`

Deployment to App Service relies on its support for [custom images][].

[custom images]: https://docs.microsoft.com/en-us/azure/azure-functions/functions-create-function-linux-custom-image

**NOTE** that to use Azure App Service you must specify a public registry which
you have push access to for `RUNTIME_IMAGE_REGISTRY` in `.env` rather than
`local`.

The make tasks utilize Azure CLI and will deploy to the logged-in
subscription/account using the logged-in user's credentials. You must install
the CLI and login with `az login` before running `make`, and you must have
rights to create the needed resources.

**NOTE** that each instance created this way includes a connected Storage
account, CosmosDB account, Service Bus namespace, and Event Hubs namespace. The
azure-instance also creates and utilizes an App Service plan and functionapp.

Some triggers are triggered after the instance is created. (TODO: make them
skippable; and verify success.)

Default instance configuration options can be overriden by maintaining your own
`.env` file in the root of your clone. When no `.env` file is found one is
created based on `.env.tpl`.

## Other Options

You can also build and run the container image manually, or even build and run
the runtime and worker locally without containers.

### Build and run in a container

1. Build the Functions runtime with Go worker into a container image with
   `test/build_container.sh`. Add `1` as a parameter to also push to a
   registry, and add 'sample' or 'usr' to also build user functions in those
   directories. The image name is built from configuration in `.env` by
   default.

1. Run an instance of the runtime with Docker, providing connection strings in
   env vars to connect to Azure services, as in the following commands; more
   details on these connection strings in the following section. The image name
   chosen reflects the defaults in `.env.tpl`.

   ```bash
   run_image_uri=local/azure-functions-go-with-samples:dev
   published_port=8080
   docker container run --rm --detach \
       --name functions-tester \
       --publish "${published_port}:80" \
       --env "AzureWebJobsStorage=$sa_connstr" \
       --env "ServiceBusConnectionString=$sb_connstr" \
       --env "EventHubsConnectionString=$eh_connstr" \
       --env "CosmosDBConnectionString=$cdb_connstr" \
       "${run_image_uri}"
   ```

### Connections to Azure Services

For the Functions runtime to handle their events it must be able to connect to
Azure services. This is faciliated by connection strings retrieved by the
runtime from environment variables. _Connection strings_ in Azure are
collections of semicolon-delimited name-value pairs with details for
connecting to a service. The names of env vars to look for are as specified in
`function.json` files; the following names are used in all our samples and
scripts, and we recommend you don't change them.

For your convenience, CLI commands for getting these strings are also listed.

- AzureWebJobsStorage - `az storage account show-connection-string ...`
- ServiceBusConnectionString - `az servicebus namespace authorization-rule keys list ...`
- EventHubsConnectionString - `az eventhubs namespace authorization-rule keys list ...`
- CosmosDBConnectionString - `az cosmosdb list-keys ...`, formatted into
  `AccountEndpoint=https://${account_name}.documents.azure.com:443/;AccountKey=${account_key};"`

### Build and run locally without containers

1. Build the worker and the samples: `test/build.sh native sample`
1. Get and install the [functions
   runtime](https://github.com/Azure/azure-functions-host) per instructions in
   that repo.
1. Set environment variables:

   ```bash
   FUNCTIONS_WORKER_RUNTIME=golang              # intended target language worker
   AZURE_FUNCTIONS_ENVIRONMENT=Development      # Needed so logs are sent to STDOUT not just to files
   AzureWebJobsScriptRoot=/home/site/wwwroot    # path in container fs to user code
   AzureWebJobsStorage=                         # Storage account connection string
   EventHubsConnectionString=                   # Event Hubs namespace connection string
   ServiceBusConnectionString=                  # Service Bus namespace connection string
   CosmosDBConnectionString=                    # CosmosDB connection string
   ```

1. In `github.com/Azure/azure-functions-host`, modify
   `src/WebJobs.Script.WebHost/appsettings.json` as follows to specify the path
   to the Go worker:

   ```json
   "langaugeWorkers": {
     "workersDirectory":
        "/home/functions-user/go/src/github.com/Azure/azure-functions-go/workers"
   }
   ```

# Write and deploy a Go Function

Follow these steps to create Go Functions:

1.  Write a Go Function.
2.  Deploy it.
3.  Trigger and watch it.

Following are step-by-step instructions to prepare a Go Function triggered by
an HttpTrigger, as demonstrated in [the HttpTrigger sample][].

> See [the wiki][] and [Things to Note](#things-to-note) below for more details
> on the programming model.

[the httptrigger sample]: ./sample/HttpTrigger
[the wiki]: https://github.com/Azure/azure-functions-go/wiki/Programming-Model

## Write a Go Function

1.  Create a directory with the files for your Go Function: `mkdir myfunc && cd myfunc && touch main.go; touch function.json`.

1.  Put the following code in `main.go`.

    ```go
    package main

    import (
        "encoding/json"
        "fmt"
        "io/ioutil"
        "net/http"

        "github.com/Azure/azure-functions-go/azfunc"
    )

    // Run runs this Azure Function if/because it is specified in `function.json` as
    // the entryPoint. Fields of the function's parameters are also bound to
    // incoming and outgoing event properties as specified in `function.json`.
    func Run(ctx azfunc.Context, req *http.Request) (*User, error) {

        // additional properties are bound to ctx by Azure Functions
        ctx.Log(azfunc.LogInformation,"function invoked: function %v, invocation %v", ctx.FunctionID(), ctx.InvocationID())

        // use Go's standard library to:
        //  handle incoming request:
        body, _ := ioutil.ReadAll(req.Body)

        // to deserialize JSON content:
        var data map[string]interface{}
        var err error
        err = json.Unmarshal(body, &data)
        if err != nil {
            return nil, fmt.Errorf("failed to unmarshal JSON: %v\n", err)
        }

        // and to get query param values:
        name := req.URL.Query().Get("name")

        if name == "" {
            return nil, fmt.Errorf("missing required query parameter: name")
        }

        // Prepare a struct to return. The special output binding name
        // `$return` transforms the struct into near-equivalent JSON.
        u := &User{
            Name:     name,
            Greeting: fmt.Sprintf("Hello %s. %s\n", name, data["greeting"].(string)),
        }

        return u, nil
    }

    // User exemplifies a struct to be returned. You can use any struct or *struct.
    type User struct {
        Name     string
        Greeting string
    }
    ```

1.  Put the following configuration in the `function.json` file next to
    `main.go`. `function.json` specifies bindings between incoming and outgoing
    event properties and the structs and types in your code.

    For more details see [the function.json
    wiki](https://github.com/Azure/azure-functions-host/wiki/function.json).

    ```json
    {
      "entryPoint": "Run",
      "bindings": [
        {
          "name": "req",
          "type": "httpTrigger",
          "direction": "in",
          "authLevel": "anonymous"
        },
        {
          "name": "$return",
          "type": "http",
          "direction": "out"
        }
      ],
      "disabled": false
    }
    ```

## Deploy it

With your Function written, you're ready to package and deploy it to a Go Functions instance.

During preview the recommended pattern for deployment is to build an image with
runtime, Golang worker, and user functions included. To facilitate this we've
provided a `usr` directory where you can put properly structured function
files, and then run `make local-instance-with-usr` (or `make azure-instance-with-usr`) to automatically build and include your functions in
the image.

Each function should be in a directory bearing its intended name. Within that
directory there should be a main.go and function.json file. On build, a file
`bin/${function_name}.so` will also be added to the directory. All three files
(even the source in main.go) are used by the runtime.

The structure of a function is as follows (paths marked \* are added by builder):

```
| UserFunc1
 \ main.go
 | function.json
 | bin*
 \  UserFunc1.so*
```

### Other options

If you prefer to deploy your functions to a running instance you can consider
the following options.

- Put your structured functions in a directory and mount that directory into
  `/home/site/wwwroot` in the local container.

- For Azure instances, FTP files with `curl` or zip-deploy them with `az functionapp deployment source config-zip --src zippedfunc.zip ...`. You'll
  also need to change the functionapp's appsetting (aka environment variable)
  `WEBSITES_ENABLE_APP_SERVICE_STORAGE` to `true`.

  The script `test/deploy_function.json` takes a path to a user function which
  it builds and FTPs to the configured Azure instance.

## Trigger and watch it

Now your Function is live and ready to handle events. Time to trigger it!

1.  Use a tool like [Postman](https://www.getpostman.com/apps) or `curl` to
    execute a request with the following parameters (in this case for a local
    instance on port 8080). The path after `/api/` is the name of your
    function.

    ```
    HTTP Method: `POST`
    URL: `http://localhost:8080/api/HttpTrigger?name=world`
    Headers: `Content-Type: application/json`
    Body: `{"greeting": "How are you?"}`
    ```

    ```bash
    declare PORT=8080 FUNCTION_NAME=HttpTrigger PERSON_NAME=world
    curl -L "http://localhost:${PORT}/api/${FUNCTION_NAME}?name=${PERSON_NAME}" \
        --data '{ "greeting": "How are you?" }' \
        --header 'Content-Type: application/json' \
    ```

    The `Run` method from the sample should be executed and a User object with
    Name and Greeting properties like the following should be returned:

    ```json
    {
      "Name": "world",
      "Greeting": "Hello world. How are you?\n"
    }
    ```

# More information

## Things to note

- `function.json::entryPoint` names the Go function in package main to be used
  as the Azure Function entry point. In this example that function is named
  `Run` but any name is okay as long as it is also specified in
  `function.json`.
- `main.go` is the required name for the file containing the entry point Go
  function.
- You can use any dependencies you want in your app since they'll be compiled
  into the built binary.
- Structs in the function signature are initialized based on properties in the
  incoming event and specifications in function.json. In the example signature
  of `func Run(ctx azfunc.Context, req *http.Request) (User, error)`; `ctx azfunc.Context`, `req *http.Request` and `User` are automatically bound to
  incoming and outgoing message properties. Properties received from the GRPC
  channel are bound to properties on the Go structs, and any Go struct with the
  named properties can be used; that is, there's nothing special about the
  default types provided in package azfunc. This is illustrated by the returned
  `User` struct in the example.
- **Properties are bound to parameters based on the name of the parameter! You
  can change the order, but the name has to be consistent with the name of the
  binding defined in `function.json`!**
- You can specify a named return type, which then needs to match an output
  binding in `function.json`. Alternatively, you can have 1 unnamed return type
  which will match the special `$return` binding.
- You can also have an optional `error` return (named or anonymous) value to
  signal that the function execution failed for whatever reason.
- Having pointer types is preferred, but you can also have parameters and
  return values as non-pointer types for your functions.
- Logging to console is enabled by default for logs from functions. To enable logs
  from the worker have a look in the `host.json` file. To disable logging to console
  entirely, set the `AzureFunctionsJobHost__Logging__Console__IsEnabled` to `false` in
  the dockerfiles.

## Disclaimer

- The project is in development; problems and frequent changes are expected.
- This project has not been evaluated for production use.
- We will reply to issues in this repo; but this project is in limited preview
  and not otherwise supported.

## Contributing

This project welcomes contributions and suggestions. Most contributions require
you to agree to a Contributor License Agreement (CLA) declaring that you have
the right to, and actually do, grant us the rights to use your contribution.
For details, visit https://cla.microsoft.com.

When you submit a pull request, a CLA-bot will automatically determine whether
you need to provide a CLA and decorate the PR appropriately (e.g., label,
comment). Simply follow the instructions provided by the bot. You will only
need to do this once across all repos using our CLA.

This project has adopted the [Microsoft Open Source Code of
Conduct](https://opensource.microsoft.com/codeofconduct/). For more
information see the [Code of Conduct
FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or contact
[opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional
questions or comments.
