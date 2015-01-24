wiki
====

prerequisites
* install go
* set go path
* go get code.google.com/p/go.crypto/bcrypt
* go get github.com/russross/blackfriday

installation
* git clone https://github.com/raygenbogen/wiki
* cd ./wiki/
* go build wiki.go

running first time
* ./wiki --host=$hostname (to create self signed certificate)

announcements
* dont use with root account
* access in browser via https://$hostname:10443
