## Description

In this project a web application based on microservices is auto deployed into k8s on every commit to this reposotory, Deployed applcation microservices fetches creadentials from the vault cluster running inside K8s cluster to access mongoDB database and fetches data from the mongoDB Database runnning inside K8s cluster.

![image](AutoDeploymentUsingGitlab&traceMonitoringUsingJaeger.png)
<br/>

The main pupose of this project is to learn Otel instrumentation in an application and sending tracing data to a tracing backend (In this application e.g Jaeger) to application visualization.

### Technologies (Used in this project)

- Aws EKS (Elastic Kubernetes Service).
- Terraform
- Jaeger
- Otel
- Golang
- Docker
- Gitlab CI/CD

### Pipeline Flow

On Every commit to this repo followings happen sequencely:

- microservices container images are built and pushed to [my dockerhub repository](https://hub.docker.com/repository/docker/mdsahiloss/jaeger-tracing/general)
- Then container images of deployments are updated in rolling-update mode to achieve **zero down time** inside k8s cluster running as AWS EKS.

### Application Workflow inside K8s

Once application gets deployed inside K8s cluster then followings happen on access request to the application using http protocol:

- Appication microservices fetches credentials from Vault cluster running inside K8s to access MongoDB Database.
- And fetches data from the MongoDB.
- And send tracing data to Jaeger agent running in the same namespace of K8s cluster

### Proof of project 

![working-video](tracing-working-video.mp4)

## Microservices

Following explaining about each microservice which is being used in this application:

### A simple Web UI: videos-web

<hr/>
<br/>

Consider `videos-web` <br/>
It's an HTML application that lists a bunch of playlists with videos in them.

```
+------------+
| videos-web |
|            |
+------------+
```

<br/>

### A simple API: playlists-api

<hr/>
<br/>

For `videos-web` to get any content, it needs to make a call to `playlists-api`

```
+------------+     +---------------+
| videos-web +---->+ playlists-api |
|            |     |               |
+------------+     +---------------+

```

Playlists consist of data like `title`, `description` etc, and a list of `videos`. <br/>
Playlists are stored in a database. <br/>
`playlists-api` stores its data in a database

```
+------------+     +---------------+    +--------------+
| videos-web +---->+ playlists-api +--->+ playlists-db |
|            |     |               |    |              |
+------------+     +---------------+    +--------------+

```

<br/>

### A little complexity

<hr/>
<br/>

Each playlist item contains only a list of video id's. <br/>
A playlist does not have the full metadata of each video. <br/>

Example `playlist`:

```
{
  "id" : "playlist-01",
  "title": "Cool playlist",
  "videos" : [ "video-1", "video-x" , "video-b"]
}
```

Take not above `videos: []` is a list of video id's <br/>

Videos have their own `title` and `description` and other metadata. <br/>

To get this data, we need a `videos-api` <br/>
This `videos-api` has its own database too <br/>

```
+------------+       +-----------+
| videos-api +------>+ videos-db |
|            |       |           |
+------------+       +-----------+
```

For the `playlists-api` to load all the video data, it needs to call `videos-api` for each video ID it has.<br/>
<br/>

## Traffic flow

<hr/>
<br/>
A single `GET` request to the `playlists-api` will get all the playlists 
from its database with a single DB call <br/>

For every playlist and every video in each list, a separate `GET` call will be made to the `videos-api` which will
retrieve the video metadata from its database. <br/>

This will result in many network fanouts between `playlists-api` and `videos-api` and many call to its database. <br/>
This is intentional to demonstrate a busy network.

<br/>

## Full application architecture

<hr/>
<br/>

```

+------------+     +---------------+    +--------------+
| videos-web +---->+ playlists-api +--->+ playlists-db |
|            |     |               |    |    [redis]   |
+------------+     +-----+---------+    +--------------+
                         |
                         v
                   +-----+------+       +-----------+
                   | videos-api +------>+ videos-db |
                   |            |       |  [redis]  |
                   +------------+       +-----------+

```

<br/>

## Run the apps: Docker

<hr/>
<br/>
There is a `docker-compose.yaml`  in this directory. <br/>
Change your terminal to this folder and run:

```
cd tracing

docker-compose build

docker-compose up

```

You can access the app on `http://localhost`. <br/>
You should now see the complete architecture in the browser
<br/>

## Traces

<hr/>
<br/>

To see our traces, we can access the Jaeger UI on `http://localhost:16686`
