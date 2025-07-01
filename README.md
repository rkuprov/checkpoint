# checkpoint

Checkpoint is package meant to ease the testing of REST APIs handlers. 

The Test method accepts the following parameters that allow you to customize the shape of the request and the expected response.
The function accepts the following parameters:
- a context 
- a URL path, including query parameters
- a URL pattern
- a collection of headers
- a collection of middlewares
- a request method
- a handler function.
- a request body

It outputs a Result struct, which contains the following fields:
- Headers: map[string]string
- StatusCode: int
- Body: []byte

as well as the error, if any.