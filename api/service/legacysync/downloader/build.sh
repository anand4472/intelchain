
# if command fails, try this
# docker buildx create --use
docker buildx build -t colibyt/intelchain-proto:latest --platform linux/amd64,linux/arm64 -f Proto.Dockerfile --progress=plain .