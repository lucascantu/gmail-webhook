# gmail-webhook
A webhook server and associated commands to manage gmail push notifications

# Available commands
```
NAME:
   gmail-webhook - Manage gmail push notifications

USAGE:
   gmail-webhook [global options] command [command options] [arguments...]

VERSION:
   1.0.0

COMMANDS:
   subscribe	create a new subscription to a topic
   authorize	authorize gmail client
   watch	setup watch request for subscribed topic
   run		starts the webhook server for gmail push notifications
   help, h	Shows a list of commands or help for one command
   
GLOBAL OPTIONS:
   --help, -h		show help
   --version, -v	print the version
```

## Sub commands
   
```
NAME:
   subscribe - create a new subscription to a topic

USAGE:
   command subscribe [command options] [arguments...]

DESCRIPTION:
   

OPTIONS:
   --key-file, -k 	key file for authorization(required)
   --project-id, --id 	unique id of the project(required)
   --topic, -t 		Name of a topic that is already created
   --name, -n 		Name of the subscription
   --endpoint, -e 	Name of the subscription endpoint
```
   
```
NAME:
   authorize - authorize gmail client

USAGE:
   command authorize [command options] [arguments...]

DESCRIPTION:
   

OPTIONS:
   --gmail-secret, --gs 	gmail client secret json file
   --cache-file, --cf 		location of cached gmail token file, defaults to ~/.credentials/gmail.json [$CACHE_TOKEN_FILE]
```
   
```
NAME:
   watch - setup watch request for subscribed topic

USAGE:
   command watch [command options] [arguments...]

DESCRIPTION:
   

OPTIONS:
   --topic, -t 			Name of the topic
   --project, -p 		Name of the project
   --cache-file, --cf 		location of cached gmail token file, defaults to ~/.credentials/gmail.json [$CACHE_TOKEN_FILE]
   --gmail-secret, --gs 	gmail client secret json file
``` 

```
NAME:
   run - starts the webhook server for gmail push notifications

USAGE:
   command run [command options] [arguments...]

DESCRIPTION:
   

OPTIONS:
   --subscription, -s 		Name of the subscription
   --project, -p 		Name of the project
   --cache-file, --cf 		location of cached gmail token file, defaults to ~/.credentials/gmail.json [$CACHE_TOKEN_FILE]
   --gmail-secret, --gs 	gmail client secret json file
   --gh-token, --ght 		github personal access token file, defaults to ~/.credentials/github.json
   --log, -l 			Name of the log file(optional), default goes to stderr
   --port '9998'		port on which the server listen
   --label 			Gmail label which will be filtered for messages
   --repository, -r 		Github repository
   --owner 			Github repository owner
``` 
