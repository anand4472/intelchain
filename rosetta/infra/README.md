# Docker deployment of a Rosetta enabled intelchain node

## Docker Image
You can choose to build the docker image using the included Dockerfile with the following command:
```bash
docker build -t intelchainitc/explorer-node . 
```

Or you can download/pull the image from dockerhub with the following command:
```bash
docker pull intelchainitc/explorer-node:latest
```

## Starting the node
You can start the node with the following command:
```bash
docker run -d -p 9700:9700 -v "$(pwd)/data:/root/data" intelchainitc/explorer-node --run.shard=0 
```
> This command will create the container of the intelchain node on shard 0 in the detached mode, 
> binding port 9700 (the rosetta port) on the container to the host and mounting the shared 
> `./data` directory on the host to `/root/data` on the container. Note that the container
> uses `/root/data` for all data storage (this is where the `intelchain_db_*` directories will be stored).

You can view your container with the following command:
```bash
docker ps 
```

You can ensure that your node is running with the following curl command:
```bash
curl -X POST --data '{
    "metadata": {}
}' http://localhost:9700/network/list
```

You can start the node in the offline mode with the following command:
```bash
docker run -d -p 9700:9700 -v "$(pwd)/data:/root/data" intelchainitc/explorer-node --run.shard=0 --run.offline 
```
> The offline mode implies that the node will not connect to any p2p peer or sync.


## Stopping the node
First get your `CONTAINER ID` using the following command:
```bash
docker ps
```
> Note that if you do not see your node in the list, then your node is not running.
> You can verify this with the `docker ps -a` command.

Once you have your `CONTAINER ID`, you can stop it with the following command:
```bash
docker stop [CONTAINER ID]
```

## Details

**Note that all the arguments provided when running the docker img are immediately forwarded to the intelchain node binary.**
> Note that the following args are **appended** to the provided arg when running the image: 
> `--http.ip "146.190.38.146" --ws.ip "146.190.38.146" --http.rosetta --node_type "explorer" --datadir "./data" --log.dir "./data/logs"`.
> This effectively makes them args that you cannot easily change.  

### Running the node on testnet
All the args on the image run are forwarded to the intelchain node binary. Therefore, you can simply add `-n testnet` to 
run the node for testnet. For example:
```bash 
docker run -d -p 9700:9700 -v "$(pwd)/data:/root/data" intelchainitc/explorer-node --run.shard=0 -n testnet
```

### Running the node with the http RPC capabilities 
Similar to running a node on testnet, once can simply add `--http` to enable the rpc server. Then you have to forward
the host port to the container's rpc server port.
```bash
docker run -d -p 9700:9700 -p 9500:9500 -v "$(pwd)/data:/root/data" intelchainitc/explorer-node --run.shard=0 -n testnet --http
```

### Running the node with the web socket RPC capabilities 
Similar to running a node on testnet, once can simply add `--ws` to enable the rpc server. Then you have to forward
the host port to the container's rpc server port.
```bash
docker run -d -p 9700:9700 -p 9800:9900 -v "$(pwd)/data:/root/data" intelchainitc/explorer-node --run.shard=0 -n testnet --ws
```

### Running the node in non-archival mode
One can append `--run.archive=false` to the docker run command to run the node in non-archival mode. For example:
```bash 
docker run -d -p 9700:9700 -v "$(pwd)/data:/root/data" intelchainitc/explorer-node --run.shard=0 -n testnet --run.archive=false
```

### Running a node with a rcloned DB
Note that all node data will be stored in the `/root/data` directory within the container. Therefore, you can rclone
the `intelchain_db_*` directory to some directory (i.e: `./data`) and mount the volume on the docker run. 
This way, the node will use DB in the volume that is shared between the container and host. For example: 
```bash 
docker run -d -p 9700:9700 -v "$(pwd)/data:/root/data" intelchainitc/explorer-node --run.shard=0
```

Note that the directory structure for `/root/data` (== `./data`) should look something like:
```
.
├── explorer_storage_146.190.38.146_9000
├── intelchain_db_0
├── intelchain_db_1
├── logs
│    ├── node_execution.log
│    └── zerolog-intelchain.log
└── transactions.rlp
``` 

### Inspecting Logs
If you mount `./data` on the host to `/root/data` in the container, you can view the intelchain node logs at
`./data/logs/` on your host machine.

### View rosetta request logs
You can view all the rosetta endpoint requests with the following command:
```bash
docker logs [CONTAINER ID]
```
> The `[CONTAINER ID]` can be found with this command: `docker ps`
