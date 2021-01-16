# tw-moderation

This is a microservice framework that is supposed to be easily deployed in a Docker environment as well as easily to extend with new microservices that handle tasks based on events created by individual or multithreaded Teeworlds Monitoring Services.

The basic idea is to have multiple small services, each connecting to a Teewords server (econ), generating events that are published on different topics.

Such a data collector must publish to the event specific topic.

The second task that the service performs is to subscribe to a topic that uniquely identifies tha server that the service is connected to.
This topic in the message broker has the purpose to feed commands into the specific Teeworlds server to beexecuted.