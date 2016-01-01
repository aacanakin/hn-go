# hn-go
hn-go is a restful hacker news api & story data consolidator using gin framework

# Installation

1. Install the dependencies

```sh
sh install.sh
```

2. Install httpie // for cronjobs

```sh
sudo apt-get install httpie
```

3. [Install mongodb](https://docs.mongodb.org/v3.0/tutorial/install-mongodb-on-ubuntu/)

4. [Install memcache](https://www.digitalocean.com/community/tutorials/how-to-install-and-use-memcache-on-ubuntu-14-04)

5. Make the configuration

In the `config.toml` file, there exists various configuration options. Just change them according to your configuration.

6. Run the server
```
go run app.go
```

7. Run the jobs // schedule these jobs in crontab
```sh
sh crons.sh
```

# Routes
```
- GET /stories/:type
    Description: Returns the stories in json format of github.com/aacanakin/hn

- PUT /stories/:type
    Description: Retrieves the stories and saves it to db

Type could be the following;
    - new
    - top 
    - job
    - ask
    - show
```
