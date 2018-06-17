# Using NGINX as an API Gateway for Securing Microservices

## Requirements
**Docker**: You will need Docker installed for going through this short article. Installation instructions can be found here: https://docs.docker.com/install/

**Text Editor**: Any text editor will do, I recommend Visual Code: https://code.visualstudio.com/

**Golang** *(optional)*: https://golang.org/dl/ not mandatory, but makes it easier for writing/testing the go code outside of docker

If you don't want to write any of the code, but just want to follow along and test it out on your own. All code can be retrieved from: https://github.com/Pungyeon/docker-nginx-example

## Introduction
Building Microservices is a really tough thing to do and while there is a shocking amount of hype around how and why one should build Microservices, there is an equally shocking lack of articles on creating (simple) API gateway's for your Microservices. Either that, or I am shit at using Google (which, quite frankly, is a very feasible thesis).

Either way! Let's talk API Gateway! What is it and why do I need it? API Gateways are used for centralising and distributing API calls. This ensures a single point of contact, directing all traffic through the API gateway, rather than traffic going directly to each microservice. This simplifies ttraffic flow and also comes with a lot of other really neat side-effects, some of which, we will explore in this article.

So, what should my API Gateway do? Well, other than being able to redirect requests to the correct service, the API gateway can help us with securing our microservices. This is typically done, by acting as a proxy and adding authentication and encryption for every requests which requires this. This is super helpful, as it simplifies developers workload, making their lives a lot easier. Instead of developers having to implement SSL and authentication into every single service they write, the API gateway can take care of this for them. So every connection is encrypted and also ensured to be authenticated. Since all traffic is going through the same gateway, we can also simplify logging in our architecture (we won't be covering that in this article however).

Now, there are a lot of other ways to achieve this and other tools for this purpose (such as Kong API Gateway and Spring Boot API Gateway)... If you are using Kubernetes, you are probably aware of the super-hyped Istio service-mesh, which comes with some extra features, that are all super cool. However, for now, let's delve into the simple antics of using NGINX as an API Gateway. 

## Project Structure
So the folder structure of this mini-project, will end up looking something like this:
```
.
./auth/                 # Our service for authorization
./coffee/               # Our service for delivering coffee
./tea/                  # Our service for delivering tea
./nginx/                # Files for configuration of our NGINX instance
./docker-compose.yml
```

With our docker-compose file consisting of four services: 
- The NGINX gateway/proxy
- The Coffee & Tea dummy services
- The Authorisation dummy service

## Creating the Tea service
So, first, we are going to create two, more or less, identical services: The Coffee and Tea services, which will be written in golang. Very simply, they will return a response of either Coffe or Tea being served, whenever a request is sent. Let's have a look at our Services:

### tea/main.go
```go
package main

import (
	"log"
	"net/http"
	"os"
)

func teaHandler(w http.ResponseWriter, r *http.Request) {
	servant, err := os.Hostname()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte("Your Tea has been served by - " + servant))
}

func main() {
	http.HandleFunc("/tea", teaHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

So, a very simple HTTP server, listetning on port 8080, which will respond to requests of /tea, with an answer of "Your Tea has been served by - hostname". The hostname is retrieved from the `os.Hostname()` call, which we are using, so we can identify which container has responded to our call. This will be useful later. 

Next step is to create a simple docker file for our tea service:

#### tea/Dockerfile
```docker
FROM golang
WORKDIR /tea
COPY main.go .
RUN go build main.go
EXPOSE 8080
ENTRYPOINT ["./main"]
```

So, in summary, we pull our golang docker image, set our working directory `/tea` copy our main.go file and compile it with `go build main.go`, which will place a `main` executable binary file in our working directory. We expose port `8080`, so that other containers on the network can reach our web service (on port 8080) and finally, we specify that when the docker container is run, we run our `main` binary.

> **NOTE**: To create our coffee service, simply copy both main.go and the Dockerfile into the coffee folder and change the HTTP response from "Your Tea has been..." to "Your Coffee has been..."... or whatever you feel like sending back. There is no need to change the Dockerfile.

## Setting up NGINX
Now that we have both our services that we want to be served by NGINX, we just need to configure our NGINX service. There is no hocus pocus about this. NGINX running in Docker, is configured exactly the same way as normally. Let's begin by creating a file in our nginx folder:

#### nginx/nginx.conf
```nginx
events {
    worker_connections 1024;
}

http {
    server {
        listen 8080;

        location /tea {
            proxy_pass http://tea:8080/tea;
        }
    }
}
```
So essentially, this simple NGINX config file sets the `worker_connections` (the maximum amount of concurrent connections) to 1024 and we define an http server, listening on port 8080. This server, will redirect request on url path /tea to our tea service container on port 8080. So, in other words, if the IP of our NGINX server is 10.10.10.10, if we send a GET request to http://10.10.10.10:8080/tea, this will be redirected to http://tea:8080/tea. Clients of our services will, however, not be aware of this whatsoever.

> **NOTE**: The "tea" service will be registered with docker-compose's service discovery. This works pretty much exactly like DNS, so "tea" will be resolved to the IP address of our tea service container. The tea service is on it's on (docker) network, so we won't be able to directly reach it on http://tea:8080/tea, event if we acquired the ip address of the container. However, we can reach our nginx server on localhost, because we are exposing this containers port 8080 to our localhost (see below).

## Setting up Docker Compose
Cool, and now to finish the first part of our application, we will create a `docker-compose.yml` in our root directory, which will define our application to include our tea service and our nginx proxy:

#### ./dockercompose.yml
```yaml
version: '3'
services:
  tea:
    build: tea/.
  nginx:
    image: nginx
    ports:
      - "8080:8080"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
```

So, the only notable thing we are doing so far, is defining our tea service as `tea` and our NGINX service as `nginx`. On our NGINX service, we are exposing our service on port 8080 (on the docker host) and mapping it to port 8080. We are also adding a volume, which in this case is a single file (our config file), which we are giving the container read-only access to with the `:ro` statement at the end of the volume statement. We are mapping this to `/etc/nginx/nginx.conf`, as this is the default file path of the NGINX configuration file. So, if we run:

> docker-compose up

We will be able to be served tea by calling:

> curl http://localhost:8080/tea

...or going to this site via. a browser. Either way, we should get a result similar to:

> Your Tea has been served by - 1ac000dbfc17

## Adding a Coffee Service
So, now that we have our tea service running, we can implement our almost identical coffee service:

#### coffee/main.go
```go
package main

import (
	"log"
	"net/http"
	"os"
)

func coffeeHandler(w http.ResponseWriter, r *http.Request) {
	servant, err := os.Hostname()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
        return
	}
	w.Write([]byte("Your Coffee has been served by - " + servant))
}

func main() {
	http.HandleFunc("/coffee", coffeeHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

We can simply copy and paste our `tea/Dockerfile` into our coffee folder, and now we have our coffee service. Magic. We can then expand our docker-compose file adding our coffee service:

#### ./docker-compose.yaml
```yaml
version: '3'
services:
  coffee:
    build: coffee/.
  tea:
    build: tea/.
  nginx:
    image: nginx
    ports:
      - "8080:8080"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
```

The final step now, is to add another line (within our server{} section) to our NGINX configuration, to redirect our users to the coffee service:

```nginx
location /coffee {
    proxy_pass http://coffee:8080/coffee;
}
```

> **NOTE**: something to take note of with these locations are, that they are not strict. This means that all subrequest os `/coffee`, will also be passed onto our coffee service. So, if we decide to create a new handler with the URI of http://coffee:8080/coffee/aeropress and another called http://coffee:8080/coffee/pourover. These API endpoints can also be access via. our NGINX gateway, without making any changes to our configuration file.

If we were to run our docker-compose file now. We would be able to access both of tea service on `localhost:8080/tea` and our coffee service on `localhost:8080/coffee`. Which is pretty neat! Unfortunately, our services aren't living up to standard security standards. Most importantly, we are missing out on encryption in transit, as we are using HTTP instead of HTTPS, and there is no authentication/authorization so anyone can access our services. Having to write HTTPS and authentication modules for both/all services, might become a tedious process. If our teams working on our services have to do this independantly, we might also introduce inconsistencies into our environment. Not good. However, this is where our API gateway is going to help us. A lot.

## Implementing SSL via. NGINX
Using NGINX, we can implement SSL on both of our services at the same time, without having to touch the code of either of our services. However, first, we need to do a little preparation, by creating our SSL certificates for encrypting traffic between our clients and our NGINX proxy. Here is an example of how to create of certificate key pair, using openssl:

> sudo openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout nginx/ssl/nginx.key -out nginx/ssl/nginx.crt

Quite simply, we are requesting a x509 standard certificate from OpenSSL using RSA 2048 bit encryption and expires after 1 year. We are placing these certificates in our nginx/ssl folder as our private key (nginx.key) and our public certificate (nginx.crt). SSL Certificates and encryption is a topic complex enough for a lifetime, so I won't cover it too much here. However, to summarise the basics, the publiic certificate will be used by our clients to encrypt traffic sent to our NGINX gateway, which will use the private certificate to decrypt incoming traffic.

TO have NGINX use these certificates, we adjust our configuration as such:

#### nginx/nginx.conf
```nginx
events {
    worker_connections 1024;
}

http {
    server {
        listen 443 ssl;

        ssl_certificate /etc/nginx/ssl/nginx.crt;
        ssl_certificate_key /etc/nginx/ssl/nginx.key;

        location /coffee {
            proxy_pass http://coffee:8080/coffee;
        }

        location /tea {
            proxy_pass http://tea:8080/tea;
        }
    }
}
```

We have removed our `listen 8080` line and replaced it with `listen 443 ssl`. There is nothing wrong with using other ports, however, 443 is the default port for HTTPS traffic, so using it for our service makes life easier for everyone. We are also specifying our public certificate: `ssl_certificate` and our private key: `ssl_certificate_key`. This means, we will also need to refer to these files, in our docker-compose file, as we did with the config, making our NGINX service definition look as such:

#### ./docker-compose.yml
```yaml
    ...
    nginx:
        image: nginx
        ports:
        - "443:443"
        volumes:
        - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
        - ./nginx/ssl:/etc/nginx/ssl:ro
```

Now, we are referring to the entire folder `./nginx/ssl`, rather than the individual files and exposing port 443 instead of 8080. So now, we can run `docker-compose up -d` once more, to test that we can now access our services using HTTPS.

> curl https://localhost/coffee -k

Notice that we are no longer specifying a port (:8080), as mentioned earlier, we are using the default SSL port and therefore it is no longer necessary. However, we do need to use the -k parameter. Our SSL certificate is self-signed, meaning that it is not signed by a trusted certificate authority (a trusted entity) and therefore seen as 'insecure'. The `-k` parameter, will ignore checking the certificate vailidity. If we were to access our service via. a browser, we would get a certificate warning (which we can also ignore).

But boom, we are now encrypting, and essentially all it took was 3 lines of configuration change. 

## Setting up Authentication with NGINX
Ok, we are almost there. As mentioned before, anyone can access our services right now. We don't want anyone to access our coffee and tea, so to prevent that we will setup some authentication. Using the same approach / mentality as with setting up SSL, we really don't want to touch the code of our already existing services. Thankfully, this is completely possible with NGINX. Let's have a look at what that looks like configuration wise.

#### nginx/nginx.conf
```nginx
events {
    worker_connections 1024;
}

http {
    server {
        listen 443 ssl;
        
        ssl_certificate /etc/nginx/ssl/nginx.crt;
        ssl_certificate_key /etc/nginx/ssl/nginx.key;

        location /coffee {
            auth_request /auth;
            auth_request_set $auth_status $upstream_status;

            proxy_pass http://coffee:8080/coffee;
        }

        location /tea {
            auth_request /auth;
            auth_request_set $auth_status $upstream_status;

            proxy_pass http://tea:8080/tea;
        }

        location /auth {
            internal;
            proxy_pass http://auth:8080/authenticated;
        }
    }
}
```

So, this is what our nginx configuration looks like. As you can see, we have added a few things. On both of our services, we have added the line:

> auth request /auth;

This line will pass our incoming request through our `/auth` location. If this auth request is successful, the request will then be sent to our coffee or tea service, as it has been previously, however, if the auth request is unsuccessful, NGINX will return an error status (such as 401). At the bottom of our configuration we have added our auth location. This service is defined as `internal`, which ensures that anyone other than NGINX trying to access this location will get a `404 Not Found`. This location is private to our service. All, we do with this is send the request on to another service, our authentication service, which we shall write now...

## Writing our Example Authentication Service
> **NOTE**: So, just to be clear. This is merely an example service, this is not secure and is exclusively for demo usage. 

Our authentication service will be responsible for one thing, and one thing only. Giving us an answer to whether or not a request has the correct `Authorization` header. 

#### auth/main.go
```go
package main

import (
	"log"
	"net/http"
)

func checkAuth(w http.ResponseWriter, r *http.Request) {
	authString := r.Header.Get("Authorization")
	if authString == "CSlkjdfj3423lkj234jj==" {
        w.Write([]byte("Authenticated: True"))
        return
	}
	http.Error(w, "Authorized: false", http.StatusUnauthorized)
}

func main() {
	http.HandleFunc("/authenticated", checkAuth)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

So, a simple web server, just like our coffee and tea services. With a single handler `/authenticated`, which simply checks whether the authorization header of the incoming request is our statically defined auth string. If it does, a HTTP 200 (http.StatusOK) status is returned and if not, then a HTTP 401 error status will be returned (http.StatusUnauthorized) is returned.

So, of course, this is not how actual authentication services work. However, this is the important part for displaying how to use an authentication service with NGINX. Let's imagine that our authentication service has a login handler (which is open to everyone), on success, this handler will return a JWT token. For every subsequent request, our client must include this JWT token in his Authorization header, granting him access to the rest of our services. Checking whether the JWT token is valid, will be the job of our `/authenticated` handler, returning a 401 or 200, just like our auth service does.

Using this setup, our other services aren't even aware of our authentication and authorization service, making them truly agnostic towards the type of authorization being used. We can switch out authentication methods if appropriate, add external authentication services etc. The only thing important to us, is that our NGINX proxy can check the incoming request parameters for a valid token or equivalent.

So let's update our docker-compose, by adding our auth service, which in it's final form looks like this:

#### ./docker-compose.yml
```yaml
version: '3'
services:
  coffee:
    build: coffee/.
  tea:
    build: tea/.
  auth:
    build: auth/.
  nginx:
    image: nginx
    ports:
      - "8080:8080"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
      - ./nginx/ssl:/etc/nginx/ssl:ro
      - ./nginx/index.html:/app/html:ro
```

> **NOTE**: just like with our coffee and tea services, we can reuse the Dockerfile again for our auth service, due to it's extreme simplicity.

So, now all we need to do is spin up `docker-compose up -d` and afterwards, hit it up with some curl commands.

First check that indeed you don't have access to the services now, with the same curl commands we used before:

> curl https://localhost/tea -k

This should return a 401 response. However, if we include an `Authorization` header....

> curl https://localhost/tea -H "Authorization: CSlkjdfj3423lkj234jj==" -k

Which should return our now familiar message: `our Tea has been served by - ee3192e3a655`. Joy.

So, of course. This last step was a little bit more work than implementing SSL. However, please keep in mind that implementing authentication isn't any easier normally. The difference is, that this authentication method is valid for every new service introduced into our application. If we decide that we need a service for serving a different beverage, we just write that service and with just a few configuration changes, we implement SSL and authentication. This way of working makes it possible for service owners to focus on their service and security owners to focus on making great security implementations, without getting in each others way. It makes development and progress much faster, but at the same time, still ensures that security standards are met, if not heightened (since there is now more time to focus on them).

## Scaling our Services
So, some time passes, and we find out that our coffee service is now wildly popular. No problem!! We are in a docker environment, we can just scale horizontally 👍. So, let's try that:

> docker-compose scale coffee=4

Docker will do it's magic and bing-bang-boom, we now have 4 instances of our coffee service, rather than the single instance we had before. But uh oh.... Something is not working. If we run our curl command against our coffee service, we are always served by the same container...

``` 
$ curl https://localhost/tea -H "Authorization: CSlkjdfj3423lkj234jj==" -k
> Your Coffee has been served by - 64ab8788dfbb
$ curl https://localhost/tea -H "Authorization: CSlkjdfj3423lkj234jj==" -k
> Your Coffee has been served by - 64ab8788dfbb
$ curl https://localhost/tea -H "Authorization: CSlkjdfj3423lkj234jj==" -k
> Your Coffee has been served by - 64ab8788dfbb
```

What? Why? Well, it turns out that NGINX, by default, will store all DNS query results and prevent them from being re-resolved (regardless of TTL) up against the DNS. So, when we start our NGINX service, it will resolve `coffee` and `tea` to the IP address of their respective containers, and will keep this answer until next time NGINX is restarted.... 😓 Oh dear. Well, no worries! There is a workaround.

## Reconfiguring DNS
To get around this, we need to add a specified `resolver` at the top of our config, and setting the `valid` parameter to 10s. The valid parameter simply specifies that the TTL (Time to Live) of the DNS query result. This means, that when we scale up (or down) of services, NGINX will re-resolve our `coffee` and `tea` service addresses and get a response including all the current addresses.

We then need to add a specific resolution, which will then be applied our TTL value of our previously set `valid` value. In the end our configuration will look like this:

#### nginx/nginx.conf
```nginx
events {
    worker_connections 1024;
}

http {

    resolver 127.0.0.11 valid=10s;

    server {
        listen 443 ssl;

        ssl_certificate /etc/nginx/ssl/nginx.crt;
        ssl_certificate_key /etc/nginx/ssl/nginx.key;

        location /coffee {
            auth_request /auth;

            set $coffee_service coffee:8080;
            proxy_pass http://$coffee_service/coffee;
        }

        location /tea {
            auth_request /auth;
            proxy_pass http://tea:8080/tea;
        }

        location /auth {
            internal;
            proxy_pass http://auth:8080/authenticated;
        }
    }
}
```

> **NOTE**: that we have not set a value for our tea service. So, if this service was scaled, NGINX would not update it's records.

So, now let's restart our services with `docker-compose up` and try to call our service again:

``` 
$ curl https://localhost/tea -H "Authorization: CSlkjdfj3423lkj234jj==" -k
> Your Coffee has been served by - 668360b94fa8
$ curl https://localhost/tea -H "Authorization: CSlkjdfj3423lkj234jj==" -k
> Your Coffee has been served by - 64ab8788dfbb
$ curl https://localhost/tea -H "Authorization: CSlkjdfj3423lkj234jj==" -k
> Your Coffee has been served by - 266a341781b3
```

It works! Hurray 🎉 🎊. 

And that sums up it for the content of this article. Let's summarise shortly, what we have achieved: 
1. We created an NGINX configuration which would send requests to our two go services, using path-based routing. Unified all requests to go through our NGINX gateway and letting the gateway handle the Layer 7 routing.

2. We secured both our services with SSL, by having the NGINX gateway handle encryption of all incoming connections.

3. We applied authorization controls on our services, having NGINX check the authenticity of incoming requests, headed towards protected resources.

4. We ensured that if services are scaled (up or down), we can load balance traffic to these services using DNS round-robin. This is a super simple form of load-balancing, but for simple scenarios, it works just fine.

I hope this article was of some use, at the very least giving some insights as to what is possible using NGINX together with Docker. There are tons of more features and there are tons of other tools for implementing this kind of structure into your application architecture, some more focused on providing API gateway features (such as Kong API Gateway). But all in all, NGINX does the job pretty well, has a simple configuration and setup and it's something that most developers are used to working with already. For further reading, here are the official sites of the technologies used in this article.

NGINX: 
* https://www.nginx.com/
* https://www.nginx.com/resources/library/complete-nginx-cookbook/
* https://www.nginx.com/resources/library/designing-deploying-microservices/

Golang: https://golang.org/

Docker: https://www.docker.com/

---

And for further reading on building microservices, I heartily can recommend Sam Newmanns excellent book on the topic:

Building Microservices (Sam Newmann): https://www.amazon.com/Building-Microservices-Designing-Fine-Grained-Systems/dp/1491950358/ref=sr_1_1?ie=UTF8&qid=1527804394&sr=8-1&keywords=building+microservices
