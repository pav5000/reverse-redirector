# Warning

There is no encryption of transferred data and no proper authentication. This is the responsibility of the next layer. So it's recommended to tunnel only ssh or vpn connections via this tool.

# Concepts

Server here is a machine with external ip which is accessible from the client machine.

Client is a machine which you want to forward a port to.

This tool provides a simple way accessing ports of the client machine without vpn and external ip.

# Starting a client

Copy `client.example.yml` to `client.yml` and edit this config according to your needs. Fill the `token` field with some random string (better to use a password generator). This token should be shared between the server and the client.

Add at least one server address and port to the `servers` array. Several servers are supported for redundancy. If one server is down, you may use others.

Run `docker-client` to build a docker image. Then run `start-client` to start a container. You may check it's status by running `logs-client` (ctrl+c to stop viewing log). `stop-client` will stop running the container.

# Starting a server

Copy `server.example.yml` to `server.yml` and edit this config according to your needs. Fill the `token` field with the same token as the client.

Configure the `listen` field, place an address which server should listen and accept connections from the client (only one client is supported for now). Addr may be with concrete ip or without it. This addr should be accessible from the client side.

Add as many `redirects` as you like.

Run `docker-server` to build a docker image. Then run `start-server` to start a container. You may check it's status by running `logs-server` (ctrl+c to stop viewing log). `stop-server` will stop running the container.

# Example

For example you have a raspberry pi on some remote location and want be able to access it via ssh. You may rent a VM from some hosting provider. Place a server there, select some token, fill the `listen` field with a port (for example `":8080"`). Place one redirect into the config `from: 127.0.0.1:8000`, `to: "127.0.0.1:22"`. This would tell server to listen port 8000 at it's localhost. When there is a connection to 8000, it will be forwarded to the 22nd port of raspberry pi's localhost. Keep in mind that `to` field is relative to client's side.

Then run a client on your raspberry. Take the same token as in server's config and add one server to the `servers` array. Put your VM's ip there with the port 8080.

Then you can ssh to your VM and run `ssh -p 8000 127.0.0.1` to connect to your raspberry.

If you want your raspberry pi's ssh to face the Internet, just remove 127.0.0.1 from the `from` field: `from: ":8000"`. After that just run `make restart-server` and connect to your raspberry from your laptop like this: `ssh -p 8000 your.vm.ip`.
