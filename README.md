# Running the Docker

On the root folder, run:

```
docker compose build
docker compose up
```

5 server containers (app0, app1, app2, app3, app4) and 2 client containers (client0, client1) will be built.
The server nodes will automatically run start up upon `docker compose up`.

### Running the Client nodes

Open another terminal and run the following code to execute commands on the `client0` container.

```
docker exec -it client0 /bin/bash
```

Once connected to the container, run the main client file:

```
./client
```

The main chatroom will start up. a `logs` folder will also be created in the root folder of the host computer, which is connected to the /logs folder in the client container. This file stores all the messages that the client has in its chatroom.
To simulate client signing out of the chatroom, run the following code:

```
exit
```

Client will not be able to interact with the chatroom interface or get updates to the chatroom once it has signed out of the chatroom.

# Testing

### Simulate Leader node failures

To disconnect the leader node from the network, run `docker network disconnect 50041_raft_default app0` (replace app0 with the actual leader's container)
To reconnect the leader node, run `docker network connect 50041_raft_default app0`(replace app0 with the actual leader's container)

To crash the leader node's container, run `docker stop app0`(replace app0 with the actual leader's container)

### Simulate follower node failures

Note: `docker network disconnect` is not applicable for follower nodes and this is discuss in section 7.3 in our report as a limitation of the project.

To crash the follower node's container, run `docker stop app1` (replace app1 with the actual follower's container)
To restart the follower node, run `docker start app1`(replace app1 with the actual follower's container)

### Simulate client failures

To disconnect the client from the network, run `docker network disconnect 50041_raft_default client0` (Can be client0 or client1)
To reconnect the client from the network, run `docker network connect 50041_raft_default client0`(Can be client0 or client1)
