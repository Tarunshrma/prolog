**Post Request**

`
encoded_value=$(echo -n "Sanchit Sharma" | base64)
curl -X POST localhost:8080/produce \
     -H "Content-Type: application/json" \
     -d '{"record": {"value": "'"$encoded_value"'"}}'
`

**Get Request**
`curl -X GET localhost:8080/produce -d '{"offset": 2}'`

****chapter-2****

1. Install protoc-gen-go
First, ensure that you have the protoc-gen-go plugin installed. Run the following command:

bash
Copy code
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
This command installs the protoc-gen-go binary into your Go binary directory, typically $GOPATH/bin or $HOME/go/bin if GOPATH is not explicitly set.

2. Add GOPATH/bin to Your PATH Environment Variable
To make the protoc-gen-go plugin accessible to protoc, you need to add the directory containing the protoc-gen-go binary to your system's PATH.

For Unix/Linux/MacOS:

Add the following line to your shell profile file (~/.bashrc, ~/.bash_profile, ~/.zshrc, or similar):

bash
Copy code
export PATH="$PATH:$(go env GOPATH)/bin"
After adding the line, reload your shell configuration:

bash
Copy code
source ~/.bashrc   # or ~/.bash_profile or ~/.zshrc
