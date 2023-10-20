
./appdemo register -u A -p test -i 100

Usage:
  appdemo register [flags]

Flags:
  -i, --initbalance string   init initbalance (default "0")
  -p, --protectpwd string    protect pwd
  -u, --userid string        user id


./appdemo transaction -u A -p test -i 100 -b -t 10

Usage:
  appdemo transaction [flags]

Flags:
  -b, --AddrB string        B' addr
  -t, --Tx string           Transaction Num (default "0")
  -p, --protectpwd string   protect pwd
  -u, --userid string       user id



