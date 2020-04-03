# rest-client

[rest-client](https://github.com/ncrypthic/rest-client) is a command line utility
to manage and testing http APIs from command line. `rest-client` is heavily inspired
by [vim-rest-console](https://github.com/diepm/vim-rest-console). From its http APIs
collection file and execution.

## 1. Installation

There are multiple ways to install `rest-client` executable:

1. Download binary (Coming Soon)

   1.1 Linux
   
   1.2 OSX
   
   1.3 Windows

2. With go utility

   If you have golang installed, you can simply run `go install github.com/ncrypthic/rest-client`
   This will automatically install `rest-client` binary to `GOBIN` path.

## 2. Usage

To use `rest-client`, first you must have a [HTTP APIs collection file](#collection-file)

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

[Variable name]: [Variable value] # to use variable in Endpoint, add a colon (`:`) followed by the variable name (e.g. `:token`)

```

Variable placeholder will be subtitude with corresponding value on URL path, header value or payload of an [Endpoint](#endpoint). `rest-client` will also **watch the API collection file for any changes in the file** and reload every endpoints.

### Endpoint

The `[Endpoint]` must be in the following format:

```
[SERVER HOST:PORT] # If not exists, will be using host:port from global variables

[HTTP HeaderName]: [HTTP Header Value]
[HTTP HeaderName]: [HTTP Header Value]
...

[HTTP METHOD] [PATH]

[REQUEST BODY]
```

## 3. Example

Collection APIs files named `collection.rest` contain like this:

```
http://example.com

token: xyz123
user_id: 123
post_id: abcde12345

--

authorization: :token

GET /home

--


http://example1.com

authorization: BEARER :token

PUT /users/:user_id/posts/:post_id

{
    "username": "test",
    "password": "topsecret"
}

--

http://example3.com

POST /users/register

{
    "username": "test",
    "password": "topsecret"
}

--

POST /users/register

{
    "username": "test",
    "password": "topsecret"
}
```

When run `rest-client development.rest` it will list the following APIs:

```
1) Variable
2) GET http://example.com/home
3) PUT http://example1.com/users/123/posts/abcde12345
4) POST http://example3.com/users/register
5) POST http://example.com/users/register
Choose endpoint
```

If we choose the `2` in the `Endpoint menu`, it will show action menu like below:

```
1) *View
2) Execute
Endpoint menu
```

If we choose to view the endpoint (action no.1) from `Action menu`, it will show like below:

```
GET /home

authorization: :token
```

If we choose to execute the http request, it will return the HTTP response header & body like:

```
Status: 404 Not Found

Date: Fri, 03 Apr 2020 14:57:55 GMT
Last-Modified: Wed, 01 Apr 2020 17:02:44 GMT
Server: ECS (oxr/830B)
Vary: Accept-Encoding
Connection: Keep-Alive
Age: 165311
Content-Type: text/html; charset=UTF-8
Cache-Control: max-age=604800, proxy-revalidate
Expires: Fri, 10 Apr 2020 14:57:55 GMT
X-Cache: 404-HIT

<?xml version="1.0" encoding="iso-8859-1"?>
<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN"
         "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<html xmlns="http://www.w3.org/1999/xhtml" xml:lang="en" lang="en">
        <head>
                <title>404 - Not Found</title>
        </head>
        <body>
                <h1>404 - Not Found</h1>
                <script type="text/javascript" src="//wpc.75674.betacdn.net/0075674/www/ec_tpm_bcon.js"></script>
        </body>
</html>
```

## License

MIT

Copyright (c) 2020 - Lim Afriyadi
