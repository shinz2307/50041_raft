# Distributed Chatroom Application

This project is a distributed chatroom application implemented using the Raft consensus algorithm. It consists of multiple server nodes (`app0`, `app1`, etc.) that maintain consistency across the system and client nodes (`client0`, `client1`) that allow users to interact with the chatroom. The system is designed to tolerate node failures and ensure high availability.

---

## Prerequisites

Before running the project, ensure you have the following installed:

- **Docker** (version 20.10 or higher)
- **Docker Compose** (version 1.29 or higher)

---

## Folder Structure

The project contains the following key directories:

- `/logs`: Contains logs for chatroom messages. This folder is automatically created on the host machine when the client nodes are run.
- `/server`: Contains the server-side application code.
- `/client`: Contains the client-side application code.

---

## Running the Docker

To start the system, follow these steps:

1. Navigate to the root directory of the project.
2. Build and start the Docker containers:
   ```bash
   docker compose build
   docker compose up
   This will create and start the following containers:
   ```

5 server containers: app0, app1, app2, app3, and app4
2 client containers: client0 and client1
The server nodes will automatically start running upon executing the docker compose up command.

## Running the Client nodes

1. Open another terminal and run the following code to execute commands on the `client0` container.

```
docker exec -it client0 /bin/bash
```

2. Once connected to the container, run the main client file:

```
./client
```

The main chatroom will start up. a `logs` folder will also be created in the root folder of the host computer, which is connected to the /logs folder in the client container. This file stores all the messages that the client has in its chatroom. 3. To simulate client signing out of the chatroom, run the following code:

```
exit
```

Client will not be able to interact with the chatroom interface or get updates to the chatroom once it has signed out of the chatroom.

# Testing

### Simulate Leader node failures

Note: Replace app0 with the actual leader's container which can be found in the logs
To disconnect the leader node from the network, run :

```
docker network disconnect 50041_raft_default app0
```

To reconnect the leader node, run :

```
docker network connect 50041_raft_default app0
```

To crash the leader node's container, run

```
docker stop app0
```

### Simulate follower node failures

Note: `docker network disconnect` is not applicable for follower nodes and this is discuss in section 7.3 in our report as a limitation of the project.
Note: Replace app0 with the actual follower's container which can be found in the logs

To crash the follower node's container, run:

```
docker stop app1
```

To restart the follower node, run:

```
docker start app1
```

### Simulate client failures

To disconnect the client from the network, run (can be client0 or client1):

```
docker network disconnect 50041_raft_default client0
```

To reconnect the client from the network, run (Can be client0 or client1):

```
docker network connect 50041_raft_default client0
```
