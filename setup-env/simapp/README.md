# simap -- simple app simulator

Usage: 
simapp [options]

options:
  -action string
        [start stop]
  -name string
        App Name
  -port int
        listen port

Examples:
./simapp  -action start -name myapp1 -port 8080

2018/06/11 15:45:38 App started successfully
2018/06/11 15:45:38 Test using: curl http://127.0.0.1:8080/apps/myapp1

curl http://127.0.0.1:8080/apps/myapp1
myapp1 is Alive 
