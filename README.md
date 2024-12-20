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