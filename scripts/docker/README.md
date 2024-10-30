## Docker Image

Included in this repo is a Dockerfile that you can launch intelchain node for trying it out. Docker images are available on `intelchainitc/intelchain`.

You can build the docker image with the following commands:
```bash
make docker
```

If your build machine has an ARM-based chip, like Apple silicon (M1), the image is built for `linux/arm64` by default. To build for `x86_64`, apply the --platform arg:

```bash
docker build --platform linux/amd64 -t intelchainitc/intelchain -f Dockerfile .
```

Before start the docker, dump the default config `intelchain.conf` by running:

for testnet
```bash
docker run -v $(pwd)/config:/intelchain --rm --name intelchain intelchainitc/intelchain intelchain config dump --network testnet intelchain.conf
```
for mainnet
```bash
docker run -v $(pwd)/config:/intelchain --rm --name intelchain intelchainitc/intelchain intelchain config dump intelchain.conf
```

make your customization. `intelchain.conf` should be mounted into default `HOME` directory `/intelchain` inside the container. Assume `intelchain.conf`, `blskeys` and `itckey` are under `./config` in your current working directory, you can start your docker container with the following command:
```bash
docker run -v $(pwd)/config:/intelchain --rm --name intelchain -it intelchainitc/intelchain
```

If you need to open another shell, just do:
```bash
docker exec -it intelchainitc/intelchain /bin/bash
```

We also provide a `docker-compose` file for local testing

To use the container in kubernetes, you can use a configmap or secret to mount the `intelchain.conf` into the container
```bash
containers:
  - name: intelchain
    image: intelchainitc/intelchain
    ports:
      - name: p2p
        containerPort: 9000  
      - name: rpc
        containerPort: 9500
      - name: ws
        containerPort: 9800     
    volumeMounts:
      - name: config
        mountPath: /intelchain/intelchain.conf
  volumes:
    - name: config
      configMap:
        name: cm-intelchain-config
    
```

Your configmap `cm-intelchain-config` should look like this:
```
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm-intelchain-config
data:
  intelchain.conf: |
    ...
```