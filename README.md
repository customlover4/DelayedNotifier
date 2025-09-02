# Startup

```
cd docker
docker-compose build
docker-compose up
```

site: `http://localhost:80`
doc: `http://localhost:80/swagger/index.html`

## docker

```
notifier:
    build: # files for building
        context: ../notifier
        dockerfile: ../notifier/Dockerfile
    ports: # if you want to change internal port, you should change notifier env var and sender env var, also in dockerfile
        - "80:8080"
    environment:
        - DEBUG=false # if false debug mode will off
        - PORT=8080
        - POSTGRES_PASSWORD=qqq
        - REDIS_PASSWORD=qqq
        - RABBIT_PASSWORD=password
        - CONFIG_PATH=../config/config.yml # path to config
    volumes:
        - ../config:/config # bind config from local to container

sender:
    build: # files for building
        context: ../sender
        dockerfile: ../sender/Dockerfile
    volumes:
        - ../config:/config
    environment:
        - NOTIFIER_PORT=8080 # internal port in notifier container
        - EMAIL_PASSWORD=lmky oyvu rnwj aamc # password for email account
        - BOT_TOKEN=8052892345:AAEdWZ8pvxab1vqecabjSlPC7WMb5qZMTNs # token for telegram bot
        - RABBIT_PASSWORD=password
        - CONFIG_PATH=../config/config.yml # path to config
```
