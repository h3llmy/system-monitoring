# How to run

## development mode

```sh
go run src/main.go
```

## TODO: watch mode using air

```sh
air
```

## production mode

```sh
go build -o monitoring src/main.go
```

then run using [pm2](https://pm2.keymetrics.io/) (make sure pm2 is installed on your machine)

```sh
pm2 start ./monitoring
```
