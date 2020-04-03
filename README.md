# rest-client

[rest-client](https://github.com/ncrypthic/rest-client) is a command line utility
to manage and testing http APIs from command line. `rest-client` is heavily inspired
by [vim-rest-console](https://github.com/diepm/vim-rest-console). From its http APIs
collection file and execution.

## 1. Installation

There are multiple ways to install `rest-client` executable:

1. Download binary

   1.1 Linux
   1.2 Windows
   1.3 OSX

2. With go utility

   If you have golang installed, you can simply run `go install github.com/ncrypthic/rest-client`
   This will automatically install `rest-client` binary to `GOBIN` path.

## 2. Usage

To use `rest-client`, first you must have a [HTTP APIs collection file](Collection File)

### Collection File

HTTP APIs collection file have a specific format.

```
[Global Variables]
--
[Endpoint]
--
[Endpoint]
--
[Endpoint]
...
```

The `Global variables` must be in the following format:

```
[Server hostname:port] # Can be replaced per-endpoint

[Variable name]: [Variable value] # to use variable in Endpoint, add a colon (`:`) followed by the variable name (e.g. :token)

```

Variables will be applied on every part of [Endpoint](Endpoint)

### Endpoint

The `[Endpoint]` must be in the following format:

```
[SERVER HOST:PORT] # If not exists, will be using host:port from global variables

[HTTP METHOD] [PATH]

[HTTP HeaderName]: [HTTP Header Value]
[HTTP HeaderName]: [HTTP Header Value]
...

[REQUEST BODY]
```
## License

MIT

copyright (c) 2020 - Lim Afriyadi
