# Using NGINX as an API Gateway for Microservices

### Introduction
Building Microservices is a really tough thing to do and while there is a shocking amount of hype around how and why one should build Microservices, there is an equally shocking lack of articles on creating API gateway's for your Microservices. Either that, or I am shit at using Google (which, quite frankly, is a very feasible thesis).

Either way! Let's talk API Gateway! What is it? Why do I need it? Well, you don't necessarily need an API Gateway for your Microservices, it 100% depends on your architecture. However, in certain cases, an API Gateway is used for centralising and distributing API calls. This ensures that you always contact the API Gateway, instead of having to directly contact each microservice depending on your specific need. This simplifies the flow of traffic and also comes with a lot of other really neat side-effects, which we will explore a little in this article.

So, what should my API Gateway do? Well, other than being able to redirect requests to the correct service, the API gateway can help us with securing our microservices. This is typically done, by acting as a proxy and adding authentication and encryption for every requests which requires this. This is super helpful, as it helps developers develop quickly (no, I refuse to use the word agility). Instead of developers having to implement SSL and authentication into every single service that they write, the API gateway can take care of this for you. So every connection is encrypted and also ensured to be authenticated.

Now, there are a lot of other ways to achieve this. If you are using Kubernetes, you are probably aware of the super-hyped Istio service-mesh, which comes with some extra features, that are all super cool. 

To create SSL certificates
sudo openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout nginx/ssl/nginx.key -out nginx/ssl/nginx.crt