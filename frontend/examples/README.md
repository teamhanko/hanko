# Hanko Examples

This directory contains examples that show

- integration of web component(s) provided through the`@teamhanko/hanko-elements` package (see [elements](../frontend/elements)).
- how to validate JSON Web Tokens (JWT) issued by the Hanko [API](../backend) in a custom backend

It contains:

- an example [express](express) backend - this is a simple version of the well-known todo app
- example frontend applications using the following frameworks:
  - [Angular](angular)
  - [Next.js](nextjs)
  - [React](react)
  - [Vue](vue)
  - [Svelte](svelte)

## How to run
### Manual
1. Start the Hanko API (see the instructions on how to run the API [in Docker](../backend/README.md#Docker) or [from Source](../backend/README.md#from-source))
2. Start the express backend (see the [README](express) for the express backend)
3. Start one of the frontend applications (see the README for the app of your choice)

### Docker Compose

#### React
```
docker compose -f deploy/docker-compose/base.yaml -f deploy/docker-compose/todo-react.yaml -p "hanko-todo-react" up --build
```
#### Angular
```
docker compose -f deploy/docker-compose/base.yaml -f deploy/docker-compose/todo-angular.yaml -p "hanko-todo-angular" up --build
```
#### Next.js
```
docker compose -f deploy/docker-compose/base.yaml -f deploy/docker-compose/todo-nextjs.yaml -p "hanko-todo-nextjs" up --build
```
#### Vue
```
docker compose -f deploy/docker-compose/base.yaml -f deploy/docker-compose/todo-vue.yaml -p "hanko-todo-vue" up --build
```
#### Svelte
```
docker compose -f deploy/docker-compose/base.yaml -f deploy/docker-compose/todo-svelte.yaml -p "hanko-todo-svelte" up --build
```
