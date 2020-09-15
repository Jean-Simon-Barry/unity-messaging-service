# Deployment

To deploy, you'll need [Helm](https://helm.sh/docs/intro/install/) and [Minikube](https://kubernetes.io/docs/tasks/tools/install-minikube) (for convenience)

start your minikube
```
minikube start
```

Once started, run
```
helm install unity-msg-svc ./unity-msg-chart
```

Wait for the pods to be healthy with
```
kubectl get pods
```
**This might take a few seconds as rabbitmq is slow to boot**

Finally, once all pods healthy, you can access the app with
```
minikube service unity-msg-svc-unity-msg-chart
```

# Useful commands
Accessing the redis-cli
```
kubectl exec -it unity-msg-svc-redis-master-0 -- redis-cli
```

Accessing the rabbitmq service
```
minikube service unity-msg-svc-rabbitmq
```

Seeing the logs of the server
```
kubectl get pods | awk '/unity-msg-svc-unity-msg-chart/{print $1}' | xargs kubectl logs -f
```

## How the messaging works?
On startup, each instance of the app will register a RabbitMQ queue with a unique name and the instance will
listen for messages only for that queue. When users connect to this instance, their client id is associated with that queue via Redis.

Then, when messages requests are sent to /relay, the app figures out
which queues are being targeted (i.e. get the queues associated with each client id that was sent as a receiver of the msg)
and posts the messages to those queues (i.e. post msg A to queue XYZ). As messages are consumed from the queue, the app instance simply writes the msg
to the websocket of each of the receiver currently connected to its Hub.

## Why RabbitMQ and Redis?
The idea is that by using shared instances of RabbitMQ and Redis, we can scale the number
of message app instances, and allow the instances to communicate via these "global" bridges. So if a client A connects to HubA instance, 
and client B connects to HubB instance, client A can still see that B is connected (via Redis), and can message client B (via RabbitMQ). 
So as more and more users connect, if needed we can scale the number of instances to let users have quick and enjoyable chat experiences :).

## Strengths/Drawbacks/Thoughts
General thoughts I had during design/dev
1) What happens when a msg is sent to a user who is offline? There should be a way
to 'persist' the message until the user logs back in, where upon he should receive the message.

2) There should be an expiration on the connection/check in to the Hub. If a user has no activity,
then a user should be disconnected and checked out.

3) When a /relay request is sent, instead of posting all messages to the Rabbit queue we could first checked
if some of the receivers are connected to this particular hub instance. In that case, we could simply write directly 
to their and bypass the queue.

4) Even if a receiver is online, but the message delivery fails somehow, it should have a retry.
Or perhaps we should only ACK the queue message once it's been successfully written to the client, leaving the retry
up to RabbitMQ.

5) I am a big fan of testing, and testable code. But in this instance, I failed miserably ;_;. Especially since there are a lot of corner cases here.
As I wrote the app pieces, I just wanted to get it working and see the result. I did write _some_ api tests but most of the testing
was live debugging. So at least you could say I QAed :P.

6) I am a very novice Golanger, but I think it's pretty neat :)! I'd love to do more 
Go.
