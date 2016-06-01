# godyndns
Dynamic DNS server written in go

I got the idea from this project from http://mkaczanowski.com/golang-build-dynamic-dns-service-go/ but I wanted
to make it easier to update the dns records from clients by making the updates via HTTPs rather than dns update.

The final exe takes a number of parameters 

`--logfile` If blank logs to stdout  otherwise logs to that file. (Default Blank)

`--port` is the dns port to listen on (Default 53)

`--httpport` which https port to listen on (Default 8000)

`--bind` IP address to bind server to (default all IPs)

`--dbpath` Path to and make of the bolt db file to store the users and dns records in (Default ./dyndns.db)

`--rootdomain` domain name that is the root of the domain name i.e. `dyndns.example.co.uk.` in this case the host names will be
        `server1.dyndns.example.com` (Default blank)
        
`--adminuser` Admin username (needed for adding new users who can update dns entries) (default blank)

`--adminPass` Admin user password (default blank)

The SSL certificate and key need to be in the current directory and names `server.crt` and `server.key`

To add a user you go to

`https://servername:8000/add` 

and enter the admin username and password then fill in the form

To set the address go to

`https://servername:8000/set`

enter the username of the user added in the last step

this will set the entry for that user to the source IP address of the request

You can use curl to do that like this

`curl -k -u user:password https://server:8000/set`

to see the current records you can go to

`https://servername:8000/get`

that will show the current IP settings for that user
	