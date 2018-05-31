# Using NGINX as an API Gateway for Microservices

## Introduction
Building Microservices is a really tough thing to do and while there is a shocking amount of hype around how and why one should build Microservices, there is an equally shocking lack of articles on creating API gateway's for your Microservices. Either that, or I am shit at using Google (which, quite frankly, is a very feasible thesis).

Either way! Let's talk API Gateway! What is it? Why do I need it? Well, you don't necessarily need an API Gateway for your Microservices, it 100% depends on your architecture. However, in certain cases, an API Gateway is used for centralising and distributing API calls. This ensures that you always contact the API Gateway, instead of having to directly contact each microservice depending on your specific need. This simplifies the flow of traffic and also comes with a lot of other really neat side-effects, which we will explore a little in this article.

So, what should my API Gateway do? Well, other than being able to redirect requests to the correct service, the API gateway can help us with securing our microservices. This is typically done, by acting as a proxy and adding authentication and encryption for every requests which requires this. This is super helpful, as it helps developers develop quickly (no, I refuse to use the word agility). Instead of developers having to implement SSL and authentication into every single service that they write, the API gateway can take care of this for you. So every connection is encrypted and also ensured to be authenticated.

Now, there are a lot of other ways to achieve this and other tools for this purpose (such as Kong API Gateway and Spring Boot API Gateway)... If you are using Kubernetes, you are probably aware of the super-hyped Istio service-mesh, which comes with some extra features, that are all super cool. However, for now, let's delve into the simple antics of using NGINX as an API Gateway. 

## Project Structure
So the folder structure of this mini-project, will end up looking something like this:
.
./auth/                 # Our service for authorization
./coffee/               # Our service for delivering coffee
./tea/                  # Our service for delivering tea
./nginx/                # Files for configuration of our NGINX instance
.docker-compose.yml

With our docker-compose file consiting of four services: 
- The NGINX gateway/proxy
- The Coffe e& Tea dummy services
- The Authorisation dummy service

## Creating a very basic service
So, first, we are going to create two, more or less identical services: The Coffee and Tea services, which will be written in golang. Very simply, they will return a response of either Coffe or Tea being served, whenever a request is sent. Let's have a look at our Services:

### tea/main.go
```go
package main

import (
	"log"
	"net/http"
	"os"
)

func teaHandler(w http.ResponseWriter, r *http.Request) {
	servant, err := os.Hostname() // get the hostname of the docker container
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error, no Tea for your :("))
	}
    // return the message, together with container hostname
    w.WriteHeader(http.StatusOK)
	w.Write([]byte("Your Tea has been served by - " + servant))
}

func main() {
    // instatiate /tea routing endpoint
    http.HandleFunc("/tea", teaHandler)
    // start the server on port 8080 and log (and exit) if an error should occur.
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

So, a very simple HTTP server, listetning on port 8080, which will respond to requests of /tea. But we will get back to our Coffee service within too long. But for now let's create a simple docker file for our tea service:

#### tea/Dockerfile
```docker
FROM golang
WORKDIR /tea
COPY main.go .
RUN go build main.go
EXPOSE 8080
ENTRYPOINT ["./main"]
```

So, in summary, we pull our golang docker image, set our working directory `/tea` copy our main.go file and compile it with `go build build main.go`, which will place a `main` executable binary file in our working directory. We expose port `8080`, so that other containers on the network can reach our web service (on port 8080) and finally, we specify that when the docker container is run, we run our `main` binary.

`NOTE: To create our coffee service, simply copy both main.go and the Dockerfile into the coffee folder and change the HTTP response from "Your Tea has been..." to "Your Coffee has been..."... or whatever you feel like sending back. There is no need to change the Dockerfile.`

## Setting up NGINX
Now that we have both our services that we want to be served by NGINX, we just need to configure our NGINX service. There is no hocus pocus about this and NGINX run in Docker, is configured exactly the same way as normally. Let's begin by creating a file in our nginx folder:

#### nginx/nginx.conf
```
events {
    worker_connections 1024;
}

http {
    server {
        listen 80;

        location /tea {
            proxy_pass http://tea:8080/tea;
        }
}
```
So essentially, this simple NGINX config file sets the `worker_connections` (the maximum amount of concurrent connections) to 1024 and we define an http server, listening on port 8080. This server, will redirect request on url path /tea to our tea service container on port 8080. So, in other words, if the IP of our NGINX server is 10.10.10.10, if we send a GET request to http://10.10.10.10:8080/tea, this will be redirected to http://tea:8080/tea. The user will not be aware of this whatsoever.

`NOTE: The "tea" service will be registered with docker-compose's service discovery. This works pretty much exactly like DNS, so "tea" will be resolved to the IP address of our tea service container`

Cool, and now to finish the first part of our application, we will create a `docker-compose.yml` in our root directory, which will define our application to include our tea service and our nginx proxy:

```docker
version: '3'
services:
  tea:
    build: tea/.
  nginx:
    image: nginx
    ports:
      - "8080:8080"
```

So, the only notable thing we are doing so far, is defining our tea service as `tea` and our NGINX service as `nginx`. On our NGINX service, we are exposing our service on port 8080 (on the docker host) and mapping it to port 8080. So, if we run this:

> docker-compose up

We will be able to be served tea by calling:

> curl http://localhost:8080/tea

...or going to this site via. a browser.



To create SSL certificates
sudo openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout nginx/ssl/nginx.key -out nginx/ssl/nginx.crt